package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/template"
)

func TestListGet(t *testing.T) {
	context := &app.App{
		Config: &app.Config{
			Template: &template.Config{
				Folder:    "../../templates",
				Extension: "html",
			},
		},
	}
	req, err := http.NewRequest("GET", "/list", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ListGet(context))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
