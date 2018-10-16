package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/database"
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

func SearchMessageGet(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		message := r.URL.Query().Get("message")
		if message == "" {
			log.Println("no message in search")
			template.ErrorTemplate(w, "", http.StatusBadRequest)
			return
		}

		ListFilteredReports(w, r, db, "WHERE message LIKE ?", "%" + message + "%")
	}
}

func SearchPluginGet(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plugin := r.URL.Query().Get("plugin")
		if plugin == "" {
			log.Println("empty plugin name")
			template.ErrorTemplate(w, "", http.StatusBadRequest)
			return
		}

		ListFilteredReports(w, r, db, "WHERE plugin = ?", plugin)
	}
}

func SearchBuildGet(db *database.DB) http.HandlerFunc {
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

		ListFilteredReports(w, r, db, fmt.Sprintf("WHERE build %s ?", operator), buildID)
	}
}
func SearchReportGet(db *database.DB) http.HandlerFunc {
	query := "SELECT * FROM crash_reports WHERE id = ?"
	return func(w http.ResponseWriter, r *http.Request) {

		reportID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			log.Println(err)
			template.ErrorTemplate(w, "", http.StatusBadRequest)
			return
		}

		var report crashreport.Report
		err = db.Get(&report, query, reportID)
		if err != nil {
			log.Println(err)
			template.ErrorTemplate(w, "Report not found", http.StatusNotFound)
			return
		}

		ListFilteredReports(w, r, db, "WHERE message = ? AND file = ? and line = ?", report.Message, report.File, report.Line)
	}
}
