package main

import (
	"html/template"
	"log"
	"net/http"
)

type tmplHistoriansData struct {
	System System
	Title  string
	Conf   ConfigHistorians
}

func api_HistoriansConf(w http.ResponseWriter, r *http.Request) {
	templates, _ = template.ParseGlob(templatedir + "*") // TODO: remove once page debug is done
	dat := tmplHistoriansData{
		System: system,
		Title:  "ServerConf",
		Conf:   system.WorkingConfig.Historians,
	}
	err := templates.ExecuteTemplate(w, "Historians.html", dat)
	if err != nil {
		log.Printf("problem with template. %v", err)
	}
}
