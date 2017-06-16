package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"bitbucket.org/intyre/ca-pmmp/app"
	"bitbucket.org/intyre/ca-pmmp/app/crashreport"
	"bitbucket.org/intyre/ca-pmmp/app/template"
)

func SearchGet(app *app.App) http.HandlerFunc {
	const name = "search"
	tmpl, err := template.LoadTemplate(name, app.Config.Template)
	if err != nil {
		log.Fatalf("failed to load template %s: %v\n", name, err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "base.html", nil)
	}
}
func SearchIDPost(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/view/%d", reportID), http.StatusMovedPermanently)
	}
}
func SearchPluginPost(app *app.App) http.HandlerFunc {
	const name = "list"
	tmpl, err := template.LoadTemplate(name, app.Config.Template)
	if err != nil {
		log.Fatalf("failed to load template %s: %v\n", name, err)
	}

	query := "SELECT id, version, message FROM crash_reports WHERE plugin = ? ORDER BY id DESC"
	queryTotal := "SELECT COUNT(*) FROM crash_reports"
	return func(w http.ResponseWriter, r *http.Request) {
		plugin := r.FormValue("plugin")
		if plugin == "" {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		var reports []crashreport.Report
		err := app.Database.Select(&reports, query, plugin)
		if err != nil {
			log.Println(err)
			return
		}
		var total int
		err = app.Database.Get(&total, queryTotal)
		if err != nil {
			log.Println(err)
			return
		}

		data := make(map[string]interface{})
		data["Data"] = reports
		data["ShowCount"] = len(reports)
		data["TotalCount"] = total
		tmpl.ExecuteTemplate(w, "base.html", data)
	}
}
func SearchBuildPost(app *app.App) http.HandlerFunc {
	const name = "list"
	tmpl, err := template.LoadTemplate(name, app.Config.Template)
	if err != nil {
		log.Fatalf("failed to load template %s: %v\n", name, err)
	}

	query := "SELECT id, version, message FROM crash_reports WHERE build"
	queryTotal := "SELECT COUNT(*) FROM crash_reports"
	return func(w http.ResponseWriter, r *http.Request) {

		buildID, err := strconv.Atoi(r.FormValue("build"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		operator := "="
		typ := r.FormValue("type")
		if typ == "greater" {
			operator = ">"
		} else if typ == "less" {
			operator = "<"
		}

		var reports []crashreport.Report
		err = app.Database.Select(&reports, fmt.Sprintf("%s %s ?", query, operator), buildID)
		if err != nil {
			log.Println(err)
			return
		}
		var total int
		err = app.Database.Get(&total, queryTotal)
		if err != nil {
			log.Println(err)
			return
		}
		data := make(map[string]interface{})
		data["Data"] = reports
		data["ShowCount"] = len(reports)
		data["TotalCount"] = total
		tmpl.ExecuteTemplate(w, "base.html", data)
	}
}
func SearchReportPost(app *app.App) http.HandlerFunc {
	const name = "list"
	tmpl, err := template.LoadTemplate(name, app.Config.Template)
	if err != nil {
		log.Fatalf("failed to load template %s: %v\n", name, err)
	}

	query := "SELECT * FROM crash_reports WHERE id = ?"
	queryDupe := "SELECT id, version, message FROM crash_reports WHERE message = ? AND file = ? and line = ? ORDER BY id DESC"
	queryTotal := "SELECT COUNT(*) FROM crash_reports"
	return func(w http.ResponseWriter, r *http.Request) {

		reportID, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		var report crashreport.Report
		err = app.Database.Get(&report, query, reportID)
		if err != nil {
			log.Println(err)
			return
		}

		var reports []crashreport.Report
		err = app.Database.Select(&reports, queryDupe, report.Message, report.File, report.Line)
		if err != nil {
			log.Println(err)
			return
		}
		var total int
		err = app.Database.Get(&total, queryTotal)
		if err != nil {
			log.Println(err)
			return
		}
		data := make(map[string]interface{})
		data["Data"] = reports
		data["ShowCount"] = len(reports)
		data["TotalCount"] = total
		tmpl.ExecuteTemplate(w, "base.html", data)
	}
}
