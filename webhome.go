package main

import (
	"html/template"
	"log"
	"net/http"
)

type tmplHomeData struct {
	System System
	Title  string
}

func api_Home(w http.ResponseWriter, r *http.Request) {
	templates, _ = template.ParseGlob(templatedir + "*") // TODO: remove once page debug is done
	dat := tmplHomeData{
		System: system,
		Title:  "Home",
	}
	err := templates.ExecuteTemplate(w, "Home.html", dat)
	if err != nil {
		log.Printf("problem with template. %v", err)
	}

}
