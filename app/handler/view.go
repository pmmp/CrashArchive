package handler

import (
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
)

func ViewIDGet(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			template.ErrorTemplate(w, "Please specify a report", http.StatusNotFound)
			return
		}

		var reporterName string
		err = db.Get(&reporterName, "SELECT reporterName FROM crash_reports WHERE id = ?", reportID)
		if err != nil {
			log.Printf("can't find report %d in database: %v", reportID, err)
			template.ErrorTemplate(w, "Report not found", http.StatusNotFound)
			return
		}

		report, err := crashreport.ReadFile(int64(reportID))
		if err != nil {
			template.ErrorTemplate(w, "Report not found", http.StatusNotFound)
			return
		}

		v := make(map[string]interface{})
		v["Report"] = report
		v["Name"] = clean(reporterName)
		v["PocketMineVersion"] = report.Version.Get(true)
		v["AttachedIssue"] = "None"
		v["ReportID"] = reportID

		template.ExecuteTemplate(w, "view", v)
	}
}

var cleanRE = regexp.MustCompile(`[^A-Za-z0-9_\-\.\,\;\:/\#\(\)\\ ]`)

func clean(v string) string {
	return cleanRE.ReplaceAllString(v, "")
}
