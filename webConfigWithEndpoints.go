package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
)

type ConfigEditorWithEndpoints interface {
	Name() string
	String() string
	Update(url.Values) error
	Endpoints() []any
	NewEndpoint()
	UpdateEndpoint(url.Values) error
	RemoveEndpoint(int)
	RenderEndpoints() template.HTML
	RenderConfig() template.HTML
}

type apiConfigEditorWithEndpoints[T ConfigEditorWithEndpoints] struct {
	ConfTypeName string
	Path         string
	Confs        []T
	initialized  bool
}

type tmplConfigEditorWithEndpoints struct {
	System System
	Path   string
	Title  string
	Conf   ConfigEditorWithEndpoints
}

func (gc *apiConfigEditorWithEndpoints[T]) Init(r *mux.Router) {
	if gc.initialized {
		return
	}
	sr := r.PathPrefix(gc.Path).Subrouter()
	sr.HandleFunc("/{name}/Edit/", gc.api_EditConf)
	sr.HandleFunc("/{name}/EditEndpoint/", gc.api_EditEndpoint)
	sr.HandleFunc("/{name}/NewEndpoint/", gc.api_NewEndpoint)
	sr.HandleFunc("/Add/", gc.api_NewConf)
	gc.initialized = true
}

func (gc *apiConfigEditorWithEndpoints[T]) api_EditConf(w http.ResponseWriter, r *http.Request) {
	templates, _ = template.ParseGlob(templatedir + "*") // TODO: remove once page debug is done
	vars := mux.Vars(r)
	targetName := vars["name"]
	conf, ok := gc.findConfByName(targetName)
	if !ok {
		log.Printf("Could not find '%s'", targetName)
		return
	}
	gc.editConf(*conf, w, r)
}

func (gc *apiConfigEditorWithEndpoints[T]) editConf(conf ConfigEditorWithEndpoints, w http.ResponseWriter, r *http.Request) {

	dat := tmplConfigEditorWithEndpoints{
		System: system,
		Path:   gc.Path,
		Title:  fmt.Sprintf("Editing %s", gc.ConfTypeName),
		Conf:   conf,
	}
	err := templates.ExecuteTemplate(w, "Provider_Generic.html", dat)
	if err != nil {
		log.Printf("problem with template. %v", err)
	}
}

func (gc *apiConfigEditorWithEndpoints[T]) api_NewConf(w http.ResponseWriter, r *http.Request) {
	templates, _ = template.ParseGlob(templatedir + "*") // TODO: remove once page debug is done
	var conf T
	system.Changes = true

	gc.Confs = append(gc.Confs, conf)

	gc.editConf(conf, w, r)
}

func (gc *apiConfigEditorWithEndpoints[T]) api_EditEndpoint(w http.ResponseWriter, r *http.Request) {
	templates, _ = template.ParseGlob(templatedir + "*") // TODO: remove once page debug is done
	vars := mux.Vars(r)
	targetName := vars["name"]
	conf, ok := gc.findConfByName(targetName)
	log.Printf("Editing Endpoint %s", r.URL)
	if !ok {
		log.Printf("Could not find '%s'", targetName)
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Printf("problem parsing form: %v", err)
		return
	}

	c := *conf
	endpoints := c.Endpoints()

	// see what we're editing.
	index_str := r.FormValue("Index")
	index, err := strconv.Atoi(index_str)
	if err != nil {
		log.Printf("invalid endpoint %s not an int: %v", index_str, err)
		return
	} else if index < 0 || index >= len(endpoints) {
		log.Printf("invalid endpoint %d.  must be 0.. %d", index, len(endpoints))
		return
	}

	err = c.UpdateEndpoint(r.Form)
	if err != nil {
		log.Printf("Could not update endpoint '%s': %v", targetName, err)
		return
	}

	gc.editConf(*conf, w, r)

}

func (gc *apiConfigEditorWithEndpoints[T]) api_NewEndpoint(w http.ResponseWriter, r *http.Request) {
	templates, _ = template.ParseGlob(templatedir + "*") // TODO: remove once page debug is done
	vars := mux.Vars(r)
	targetName := vars["name"]
	conf, ok := gc.findConfByName(targetName)
	log.Printf("Editing Endpoint %s", r.URL)
	if !ok {
		log.Printf("Could not find '%s'", targetName)
		return
	}

	system.Changes = true
	c := *conf
	c.NewEndpoint()

	gc.editConf(*conf, w, r)

}

func (gc *apiConfigEditorWithEndpoints[T]) findConfByName(name string) (*T, bool) {
	for i := range gc.Confs {
		if gc.Confs[i].Name() == name {
			return &gc.Confs[i], true
		}
	}

	log.Printf("Could not find '%s'", name)
	return new(T), false
}
