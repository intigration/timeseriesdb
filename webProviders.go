package main

import (
	"html/template"
	"log"
	"net/http"
)

type tmplProvidersData struct {
	System System
	Title  string
	Conf   ConfigDataProviders
}

func api_ProidersConf(w http.ResponseWriter, r *http.Request) {
	templates, _ = template.ParseGlob(templatedir + "*") // TODO: remove once page debug is done
	dat := tmplProvidersData{
		System: system,
		Title:  "ServerConf",
		Conf:   system.WorkingConfig.DataProviders,
	}
	err := templates.ExecuteTemplate(w, "Providers.html", dat)
	if err != nil {
		log.Printf("problem with template. %v", err)
	}
}
