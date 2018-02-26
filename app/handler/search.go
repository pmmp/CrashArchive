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
		log.Println(err)
		template.ErrorTemplate(w, "", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/view/%d", reportID), http.StatusMovedPermanently)
}

func SearchPluginGet(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plugin := r.URL.Query().Get("plugin")
		if plugin == "" {
			log.Println("empty plugin name")
			template.ErrorTemplate(w, "", http.StatusBadRequest)
			return
		}

		ListFilteredReports(w, r, app.Database, "WHERE plugin = ?", plugin)
	}
}

func SearchBuildGet(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		buildID, err := strconv.Atoi(params.Get("build"))
		if err != nil {
			log.Println(err)
			template.ErrorTemplate(w, "", http.StatusBadRequest)
			return
		}

		operator := "="
		typ := params.Get("type")
		if typ == "greater" {
			operator = ">"
		} else if typ == "less" {
			operator = "<"
		}

		ListFilteredReports(w, r, app.Database, fmt.Sprintf("WHERE build %s ?", operator), buildID)
	}
}
func SearchReportGet(app *app.App) http.HandlerFunc {
	query := "SELECT * FROM crash_reports WHERE id = ?"
	return func(w http.ResponseWriter, r *http.Request) {

		reportID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			log.Println(err)
			template.ErrorTemplate(w, "", http.StatusBadRequest)
			return
		}

		var report crashreport.Report
		err = app.Database.Get(&report, query, reportID)
		if err != nil {
			log.Println(err)
			template.ErrorTemplate(w, "Report not found", http.StatusNotFound)
			return
		}

		ListFilteredReports(w, r, app.Database, "WHERE message = ? AND file = ? and line = ?", report.Message, report.File, report.Line)
	}
}
