// Package web provides utilities for handling web-related tasks.
package web

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed views/error/*
var ErrorTemplates embed.FS

//go:embed dashboard.html
var DashboardHTML []byte

func PageNotFound(ctx *gin.Context) {
	ctx.Status(http.StatusNotFound)
	tmpl, _ := template.ParseFS(ErrorTemplates, "../views/error/404.html")
	tmpl.Execute(ctx.Writer, nil)
}

func InternalServerError(ctx *gin.Context) {
	ctx.Status(http.StatusInternalServerError)
	tmpl, _ := template.ParseFS(ErrorTemplates, "views/error/500.html")
	tmpl.Execute(ctx.Writer, nil)
}

func GHbexDashboard(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	tmpl, _ := template.New("dashboard").Parse(string(DashboardHTML))
	tmpl.Execute(ctx.Writer, nil)
}
