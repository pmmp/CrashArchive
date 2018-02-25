package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pmmp/CrashArchive/app"
)

func TestSubmitGet(t *testing.T) {
	req, err := http.NewRequest("GET", "/submit", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SubmitGet)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}

func TestSubmitPost(t *testing.T) {
	context := &app.App{}
	req, err := http.NewRequest("POST", "/submit", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SubmitPost(context))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
