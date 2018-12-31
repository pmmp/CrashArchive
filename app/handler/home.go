package handler

import (
	"net/http"

	"../template"
)

func HomeGet(w http.ResponseWriter, r *http.Request) {
	template.ExecuteTemplate(w, "home", nil)
}
