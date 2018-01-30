package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/template"
)

func TestHomeGet(t *testing.T) {
	app := &app.App{
		Config: &app.Config{
			Template: &template.Config{
				Folder:    "../../templates",
				Extension: "html",
			},
		},
	}
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HomeGet(app))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
