package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

type ConfigEditor interface {
	Name() string
	String() string
	Update(url.Values) error
	RenderConfig() template.HTML
}

type apiConfigEditor[T ConfigEditor] struct {
	ConfTypeName string
	Path         string
	Confs        []T
	initialized  bool
}

type tmplConfigEditor struct {
	System System
	Path   string
	Title  string
	Conf   ConfigEditor
}

func (gc *apiConfigEditor[T]) Init(r *mux.Router) {
	if gc.initialized {
		return
	}
	sr := r.PathPrefix(gc.Path).Subrouter()
	sr.HandleFunc("/{name}/Edit/", gc.api_EditConf)
	sr.HandleFunc("/Add/", gc.api_NewConf)
	gc.initialized = true
}

func (gc *apiConfigEditor[T]) api_EditConf(w http.ResponseWriter, r *http.Request) {
	templates, _ = template.ParseGlob(templatedir + "*") // TODO: remove once page debug is done
	vars := mux.Vars(r)
	targetName := vars["name"]
	conf, ok := gc.findConfByName(targetName)
	if !ok {
		log.Printf("Could not find '%s'", targetName)
		return
	}

	if r.Method == "POST" {
		// might have form data to deal with.
		err := r.ParseForm()
		if err != nil {
			log.Printf("Problem parsing form: %v", err)
			return
		}
		err = (*conf).Update(r.Form)
		if err != nil {
			log.Printf("Problem saving '%s': %v", targetName, err)
			return
		}
	}

	gc.editConf(*conf, w, r)
}

func (gc *apiConfigEditor[T]) editConf(conf ConfigEditor, w http.ResponseWriter, r *http.Request) {

	dat := tmplConfigEditor{
		System: system,
		Path:   gc.Path,
		Title:  fmt.Sprintf("Editing %s", gc.ConfTypeName),
		Conf:   conf,
	}
	err := templates.ExecuteTemplate(w, "Historian_Generic.html", dat)
	if err != nil {
		log.Printf("problem with template. %v", err)
	}
}

func (gc *apiConfigEditor[T]) api_NewConf(w http.ResponseWriter, r *http.Request) {
	templates, _ = template.ParseGlob(templatedir + "*") // TODO: remove once page debug is done
	var conf T
	system.Changes = true

	gc.Confs = append(gc.Confs, conf)

	gc.editConf(conf, w, r)
}

func (gc *apiConfigEditor[T]) findConfByName(name string) (*T, bool) {
	for i := range gc.Confs {
		if gc.Confs[i].Name() == name {
			return &gc.Confs[i], true
		}
	}

	log.Printf("Could not find '%s'", name)
	return new(T), false
}
