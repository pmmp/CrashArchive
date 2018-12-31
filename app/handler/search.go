package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"../crashreport"
	"../database"
	"../template"
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
