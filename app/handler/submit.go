package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"bitbucket.org/intyre/ca-pmmp/app"
	"bitbucket.org/intyre/ca-pmmp/app/crashreport"
	"bitbucket.org/intyre/ca-pmmp/app/template"
)

func SubmitGet(app *app.App) http.HandlerFunc {
	const name = "submit"
	tmpl, err := template.LoadTemplate(name, app.Config.Template)
	if err != nil {
		log.Fatalf("failed to load template %s: %v\n", name, err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "base.html", nil)
	}
}

func SubmitPost(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("report") != "yes" {
			log.Println("invalid report")
			return
		}

		reportStr, err := ParseMultipartForm(r)
		if err != nil {
			log.Println(err)
			return
		}

		report, err := crashreport.Parse(reportStr)
		if err != nil {
			log.Println(err)
			return
		}

		if report.Data.General.Name != "PocketMine-MP" {
			log.Printf("spoon detected from: %s\n", r.RemoteAddr)
			return
		}

		id, err := app.Database.InsertReport(report)
		if err != nil {
			log.Println(err)
			return
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		if err = report.WriteFile(id, name, email); err != nil {
			log.Printf("failed to write file: %d\n", id)
			return
		}

		if !strings.HasSuffix(r.RequestURI, "/api") {
			http.Redirect(w, r, fmt.Sprintf("/view/%d", id), http.StatusMovedPermanently)
			return
		}

		jsonResponse(w, map[string]interface{}{
			"crashId":  id,
			"crashUrl": fmt.Sprintf("https://crash.pmmp.io/view/%d", id),
		})

	}
}

func jsonResponse(w http.ResponseWriter, data map[string]interface{}) {
	json.NewEncoder(w).Encode(data)
}

func ParseMultipartForm(r *http.Request) (string, error) {
	if err := r.ParseMultipartForm(1024 * 256); err != nil {
		return "", err
	}

	var report string
	form := r.MultipartForm
	if reportPaste, ok := form.Value["reportPaste"]; ok && reportPaste[0] != "" {
		report = reportPaste[0]
	} else if reportFile, ok := form.File["reportFile"]; ok && reportFile[0] != nil {
		f, err := reportFile[0].Open()
		if err != nil {
			return "", errors.New("could not open file")
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return "", errors.New("could not read file")
		}
		f.Close()
		report = string(b)
	}

	return report, nil
}
