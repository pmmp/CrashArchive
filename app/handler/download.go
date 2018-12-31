package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"../crashreport"
	"../database"
	"../template"
	"github.com/go-chi/chi"
	"log"
)

func DownloadGet(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			log.Println(err)
			template.ErrorTemplate(w, "", http.StatusBadRequest)
			return
		}

		reportJsonBlob, err := db.FetchRawReport(int64(reportID))
		if err != nil {
			template.ErrorTemplate(w, "Report not found", http.StatusNotFound)
			return
		}
		reportBytes, err := crashreport.JsonToCrashLog(reportJsonBlob)
		if err != nil {
			log.Printf("failed to encode crash report: %v", err)
			template.ErrorTemplate(w, "", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.log", reportID))
		w.Header().Set("Content-Length", strconv.Itoa(len(reportBytes)))
		w.Write(reportBytes)
	}
}
