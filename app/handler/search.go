package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/template"
)

func SearchGet(w http.ResponseWriter, r *http.Request) {
	template.ExecuteTemplate(w, "search", nil)
}

func SearchIDGet(w http.ResponseWriter, r *http.Request) {
	reportID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/view/%d", reportID), http.StatusMovedPermanently)
}

func SearchPluginGet(app *app.App) http.HandlerFunc {
	query := "SELECT id, version, message FROM crash_reports WHERE plugin = ? ORDER BY id DESC"
	return func(w http.ResponseWriter, r *http.Request) {
		plugin := r.URL.Query().Get("plugin")
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

		template.ExecuteListTemplate(w, reports, r.URL.String(), 1, 0, len(reports))
	}
}

func SearchBuildGet(app *app.App) http.HandlerFunc {
	query := "SELECT id, version, message FROM crash_reports WHERE build"
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		buildID, err := strconv.Atoi(params.Get("build"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			log.Println(err)
			return
		}

		operator := "="
		typ := params.Get("type")
		if typ == "greater" {
			operator = ">"
		} else if typ == "less" {
			operator = "<"
		}

		var reports []crashreport.Report
		err = app.Database.Select(&reports, fmt.Sprintf("%s %s ? ORDER BY id DESC", query, operator), buildID)
		if err != nil {
			log.Println(err)
			return
		}

		template.ExecuteListTemplate(w, reports, r.URL.String(), 1, 0, len(reports))
	}
}
func SearchReportGet(app *app.App) http.HandlerFunc {
	query := "SELECT * FROM crash_reports WHERE id = ?"
	queryDupe := "SELECT id, version, message FROM crash_reports WHERE message = ? AND file = ? and line = ? ORDER BY id DESC"
	return func(w http.ResponseWriter, r *http.Request) {

		reportID, err := strconv.Atoi(r.URL.Query().Get("id"))
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

		template.ExecuteListTemplate(w, reports, r.URL.String(), 1, 0, len(reports))
	}
}
