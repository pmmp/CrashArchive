package handler

import (
	"net/http"

	"github.com/pmmp/CrashArchive/app/template"
)

func HomeGet(w http.ResponseWriter, r *http.Request) {
	template.ExecuteTemplate(w, r, "home")
}
