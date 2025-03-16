package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/user"
)

func DeletePost(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Printf("Bad delete post from %s: %v", r.RemoteAddr, err)
			template.ErrorTemplate(w, r, "", http.StatusBadRequest)
			return
		}

		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			template.ErrorTemplate(w, r, "Please specify a report", http.StatusNotFound)
			return
		}

		var expectedAccessToken string
		err = db.Get(&expectedAccessToken, "SELECT access_token FROM crash_report_blobs WHERE id = ?;", reportID)
		if err != nil {
			log.Printf("Error fetching report access token for report %d: %v", reportID, err)
			template.ErrorTemplate(w, r, "", http.StatusBadRequest)
			return
		}

		userInfo := user.GetUserInfo(r)

		if !userInfo.CheckReportAccess(expectedAccessToken, r.FormValue("access_token")) {
			template.ErrorTemplate(w, r, "", http.StatusUnauthorized)
			return
		}

		db.Exec("DELETE FROM crash_reports WHERE id = ?", reportID)
		log.Printf("user %s deleted crash report %d", userInfo.Name, reportID)
		redirectUrl := r.FormValue("redirect_url")
		if redirectUrl == "" {
			redirectUrl = "/list"
		}
		http.Redirect(w, r, redirectUrl, http.StatusMovedPermanently)
	}
}
