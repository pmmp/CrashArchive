package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/template"
	"log"
	"github.com/pmmp/CrashArchive/app/database"
)

func DownloadGet(db *database.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			log.Println(err)
			template.ErrorTemplate(w, "", http.StatusBadRequest)
			return
		}

		reportJson, err := db.FetchReportJson(int64(reportID))
		if err != nil {
			template.ErrorTemplate(w, "Report not found", http.StatusNotFound)
			return
		}

		reportBytes, err := crashreport.JsonToCrashLog(reportJson)
		if err != nil {
			log.Printf("failed to encode report for downloading: %v", err)
			template.ErrorTemplate(w, "", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.log", reportID))
		w.Header().Set("Content-Length", strconv.Itoa(len(reportBytes)))
		w.Write([]byte(reportBytes))
	}
}
