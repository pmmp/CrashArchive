package handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

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
	query := `INSERT INTO crash_reports
		(plugin, version, build, file, message, line, type, os, reportType, submitDate, reportDate)
	VALUES
		(:plugin, :version, :build, :file, :message, :line, :type, :os, :reportType, :submitDate, :reportDate)`
	return func(w http.ResponseWriter, r *http.Request) {
		// var err error
		if r.FormValue("report") != "yes" {
			log.Println("invalid report")
			return
		}
		// log.Println(r.FormValue("name"), r.FormValue("email"))
		report, err := parseMultipartForm(r)
		if err != nil {
			log.Println(err)
			return
		}
		res, err := app.Database.NamedExec(query, &crashreport.Report{
			Plugin:     report.CausingPlugin,
			Version:    report.Version.Get(true),
			Build:      report.Version.Build,
			File:       report.Error.File,
			Message:    report.Error.Message,
			Line:       report.Error.Line,
			Type:       report.Error.Type,
			OS:         report.Data.General.OS,
			ReportType: report.ReportType,
			SubmitDate: time.Now().Unix(),
			ReportDate: report.ReportDate.Unix(),
		})
		if err != nil {
			log.Printf("failed to exec: %v\n", err)
			return
		}

		id, err := res.LastInsertId()
		if err != nil {
			log.Println(err)
			return
		}
		data := map[string]interface{}{
			"report":        report.Encoded(),
			"reportId":      id,
			"email":         r.FormValue("email"),
			"name":          r.FormValue("name"),
			"attachedIssue": false,
		}
		err = crashreport.WriteFile(id, data)
		if err != nil {
			log.Printf("failed to write file: %d %v\n", id, data)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/view/%d", id), http.StatusMovedPermanently)
	}
}

// parseMultipartForm ...
func parseMultipartForm(r *http.Request) (*crashreport.CrashReport, error) {
	var err error

	if err = r.ParseMultipartForm(1024 * 256); err != nil {
		return nil, err
	}

	f := r.MultipartForm
	if files, ok := f.File["reportFile"]; ok {
		for _, file := range files {
			m, _ := file.Open()
			b, _ := ioutil.ReadAll(m)
			f.Value["reportPaste"][0] = string(b)
			m.Close()
		}
	}

	if v, ok := f.Value["reportPaste"]; ok {
		if v[0] == "" {
			return nil, fmt.Errorf("reportPaste is empty: %+v\n", f)
		}

		report, err := crashreport.FromString(v[0])
		if err != nil {
			log.Println(err)
			return nil, err
		}

		return report, nil
	}

	return nil, errors.New("no valid MultipartForm data found")
}
