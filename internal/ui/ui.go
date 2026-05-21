package ui

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed index.html
var indexHTML embed.FS

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	content, err := indexHTML.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Could not load page", http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("index").Parse(string(content))
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
