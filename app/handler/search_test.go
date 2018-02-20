package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pmmp/CrashArchive/app"
)

func TestSearchGet(t *testing.T) {
	app := &app.App{}
	req, err := http.NewRequest("GET", "/search", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SearchGet(app))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
func TestSearchIDPost(t *testing.T) {
	app := &app.App{}
	req, err := http.NewRequest("GET", "/list", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SearchIDGet(app))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
func TestSearchPluginPost(t *testing.T) {
	app := &app.App{}
	req, err := http.NewRequest("GET", "/list", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SearchPluginGet(app))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
func TestSearchBuildPost(t *testing.T) {
	app := &app.App{}
	req, err := http.NewRequest("GET", "/list", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SearchBuildGet(app))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
func TestSearchReportPost(t *testing.T) {
	app := &app.App{}
	req, err := http.NewRequest("GET", "/list", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SearchReportGet(app))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
