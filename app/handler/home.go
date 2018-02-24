package handler

import (
	"net/http"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/template"
)

func HomeGet(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		template.ExecuteTemplate(w, "home", nil)
	}
}
