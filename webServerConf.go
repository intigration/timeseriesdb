package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type tmplServerConfData struct {
	System System
	Title  string
	Conf   ConfigGeneral
}

func api_ServerConf(w http.ResponseWriter, r *http.Request) {
	templates, _ = template.ParseGlob(templatedir + "*") // TODO: remove once page debug is done

	if r.Method == "POST" {
		// parse the form and update the working config.

		err := r.ParseForm()
		if err != nil {
			log.Printf("problem parsing form: %v", err)
			return
		}
		rate_str := r.FormValue("RestartDelay")
		rate_str = strings.ReplaceAll(rate_str, " ", "") // get rid of spaces for parsing
		rate, err := time.ParseDuration(rate_str)
		if err != nil {
			log.Printf("invalid rete %s: %v", rate_str, err)
			return
		}
		system.WorkingConfig.General.RestartDelay = rate
		system.WorkingConfig.General.Host = r.FormValue("Host")

		port, err := strconv.Atoi(r.FormValue("Port"))
		if err != nil {
			log.Printf("invalid port %s not an int: %v", r.FormValue("Port"), err)
			return
		}
		system.WorkingConfig.General.Port = port

		system.Changes = true

	}

	dat := tmplServerConfData{
		System: system,
		Title:  "ServerConf",
		Conf:   system.WorkingConfig.General,
	}
	err := templates.ExecuteTemplate(w, "Server.html", dat)
	if err != nil {
		log.Printf("problem with template. %v", err)
	}
}
