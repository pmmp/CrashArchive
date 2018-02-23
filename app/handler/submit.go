package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/template"
)

func SubmitGet(w http.ResponseWriter, r *http.Request) {
	template.ExecuteTemplate(w, "submit", nil)
}

func SubmitPost(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("report") != "yes" {
			http.Redirect(w, r, "/submit", http.StatusMovedPermanently)
			return
		}

		if err := r.ParseMultipartForm(1024 * 256); err != nil {
			http.Redirect(w, r, "/submit", http.StatusMovedPermanently)
			return
		}

		reportStr, err := ParseMultipartForm(r.MultipartForm)
		if err != nil {
			http.Redirect(w, r, "/submit", http.StatusMovedPermanently)
			return
		}

		isAPI := strings.HasSuffix(r.RequestURI, "/api")

		defer func() {
			if recovered := recover(); recovered != nil {
				err, ok := recovered.(error)
				if !ok {
					err = fmt.Errorf("%v", recovered)
				}

				log.Printf("got invalid crash report from: %s (%v)", r.RemoteAddr, err)
				sendError(w, "This crash report is not valid", http.StatusUnprocessableEntity, isAPI)
			}
		}()

		report, err := crashreport.Parse(reportStr)
		if err != nil {
			//this panic will be recovered in the above deferred function
			log.Panic(err)
		}

		if report.Data.General.Name != "PocketMine-MP" {
			log.Printf("spoon detected from: %s\n", r.RemoteAddr)
			sendError(w, "", http.StatusTeapot, isAPI)
			return
		}

		if report.Data.General.GIT == strings.Repeat("00", 20) || strings.HasSuffix(report.Data.General.GIT, "-dirty") {
			log.Printf("invalid git hash %s in report from: %s\n", report.Data.General.GIT, r.RemoteAddr)
			sendError(w, "", http.StatusTeapot, isAPI)
			return
		}

		dupes, err := app.Database.CheckDuplicate(report)
		report.Duplicate = dupes > 0
		if dupes > 0 {
			log.Printf("found %d duplicates of report from: %s", dupes, r.RemoteAddr)
		}

		name := r.FormValue("name")
		email := r.FormValue("email")

		id, err := app.Database.InsertReport(report, name, email)
		if err != nil {
			log.Printf("failed to insert report into database: %v", err)
			sendError(w, "", http.StatusInternalServerError, isAPI)
			return
		}

		if err = report.WriteFile(id); err != nil {
			log.Printf("failed to write file: %d\n", id)
			sendError(w,"", http.StatusInternalServerError, isAPI)
			return
		}

		if !report.Duplicate {
			go app.ReportToSlack(name, id, report.Error.Message)
		}

		if isAPI {
			jsonResponse(w, map[string]interface{}{
				"crashId":  id,
				"crashUrl": fmt.Sprintf("https://crash.pmmp.io/view/%d", id),
			})
		} else {
			http.Redirect(w, r, fmt.Sprintf("/view/%d", id), http.StatusMovedPermanently)
		}

	}
}

func sendError(w http.ResponseWriter, message string, status int, isAPI bool) {
	if isAPI {
		w.WriteHeader(status)
		if message == "" {
			message = http.StatusText(status)
		}
		jsonResponse(w, map[string]interface{}{
			"error": message,
		})
	} else {
		template.ErrorTemplate(w, message, status)
	}
}

func jsonResponse(w http.ResponseWriter, data map[string]interface{}) {
	json.NewEncoder(w).Encode(data)
}

func ParseMultipartForm(form *multipart.Form) (string, error) {
	var report string
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
