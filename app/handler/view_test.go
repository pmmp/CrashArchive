package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/intyre/ca-pmmp/app"
)

func TestViewIDGet(t *testing.T) {
	app := &app.App{}
	req, err := http.NewRequest("GET", "/view/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ViewIDGet(app))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
