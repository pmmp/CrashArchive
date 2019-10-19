package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/user"
)

func DeleteGet(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userInfo := user.GetUserInfo(r)
		if(!userInfo.HasDeletePerm()){
			log.Printf("access denied to %s (%s) for endpoint %s", userInfo.Name, r.RemoteAddr, r.RequestURI)
			template.ErrorTemplate(w, r, "You don't have permission to do that", http.StatusUnauthorized)
			return
		}
		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			template.ErrorTemplate(w, r, "Please specify a report", http.StatusNotFound)
			return
		}

		db.Exec("DELETE FROM crash_reports WHERE id = ?", reportID)
		log.Printf("user %s deleted crash report %d", userInfo.Name, reportID)
		redirectUrl := r.URL.Query().Get("redirect")
		if redirectUrl == "" {
			redirectUrl = "/list"
		}
		http.Redirect(w, r, redirectUrl, http.StatusMovedPermanently)
	}
}