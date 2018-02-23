package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/pmmp/CrashArchive/app"
)

func TestViewIDGet(t *testing.T) {
	context := &app.App{}
	req, err := http.NewRequest("GET", "/view/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ViewIDGet(context))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
