package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/user"
	"log"
)

func DownloadGet(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			log.Println(err)
			template.ErrorTemplate(w, r, "", http.StatusBadRequest)
			return
		}
		givenAccessToken := r.URL.Query().Get("access")

		reportJsonBlob, expectedAccessToken, err := db.FetchRawReport(int64(reportID))
		if err != nil {
			template.ErrorTemplate(w, r, "Report not found", http.StatusNotFound)
			return
		}

		if !user.GetUserInfo(r).CheckReportAccess(expectedAccessToken, givenAccessToken) {
			template.ErrorTemplate(w, r, "", http.StatusUnauthorized)
			return
		}
		reportBytes, err := crashreport.JsonToCrashLog(reportJsonBlob)
		if err != nil {
			log.Printf("failed to encode crash report: %v", err)
			template.ErrorTemplate(w, r, "", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.log", reportID))
		w.Header().Set("Content-Length", strconv.Itoa(len(reportBytes)))
		w.Write(reportBytes)
	}
}
