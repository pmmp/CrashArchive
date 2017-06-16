package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/intyre/ca-pmmp/app"
)

func TestSubmitGet(t *testing.T) {
	app := &app.App{}
	req, err := http.NewRequest("GET", "/submit", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SubmitGet(app))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}

func TestSubmitPost(t *testing.T) {
	app := &app.App{}
	req, err := http.NewRequest("POST", "/submit", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SubmitPost(app))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
