package utils

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed views/error/*
var ErrorTemplates embed.FS

func pageNotFound(rw http.ResponseWriter, r *http.Request) {

	rw.WriteHeader(http.StatusNotFound)
	tmpl, _ := template.ParseFS(ErrorTemplates, "../views/error/404.html")
	tmpl.Execute(rw, nil)
}

func internalServerError(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusInternalServerError)
	tmpl, _ := template.ParseFS(ErrorTemplates, "views/error/500.html")
	tmpl.Execute(rw, nil)
}
