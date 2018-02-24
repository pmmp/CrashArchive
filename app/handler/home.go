package handler

import (
	"net/http"

	"github.com/pmmp/CrashArchive/app/template"
)

func HomeGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		template.ExecuteTemplate(w, "home", nil)
	}
}
