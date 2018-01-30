package handler

import (
	"log"
	"net/http"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/template"
)

func HomeGet(app *app.App) http.HandlerFunc {
	const name = "home"
	tmpl, err := template.LoadTemplate(name, app.Config.Template)
	if err != nil {
		log.Fatalf("failed to load template %s: %v\n", name, err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "base.html", nil)
	}
}
