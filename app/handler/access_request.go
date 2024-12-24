package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/user"
)

func AccessRequestPost(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			template.ErrorTemplate(w, r, "Please specify a report", http.StatusBadRequest)
			return
		}
		if err := r.ParseForm(); err != nil {
			log.Println("bad access request post from %s: %v", r.RemoteAddr, err)
			template.ErrorTemplate(w, r, "", http.StatusBadRequest)
			return
		}

		userInfo := user.GetUserInfo(r)
		if userInfo.Permission < user.Basic {
			log.Println("can't submit an access request as an anonymous user")
			template.ErrorTemplate(w, r, "You need to be logged in to submit a report access request", http.StatusUnauthorized)
			return
		}

		description := r.FormValue("description")
		redirectURL := r.FormValue("redirect_url")

		requestID, err := db.InsertUserReportAccessRequest(int64(reportID), userInfo, description)
		if err != nil {
			//TODO: this could be a duplicate access request
			log.Printf("failed to insert access request for %s: %v", userInfo.Name, err)
			template.ErrorTemplate(w, r, "", http.StatusInternalServerError)
			return
		}

		v := make(map[string]interface{})
		v["MessageTitle"] = "Request submitted successfully"
		v["Message"] = fmt.Sprintf("Your request to view crashdump #%d will be reviewed by an administrator soon. Your access request ID is #%d.", reportID, requestID)
		v["RedirectURL"] = redirectURL
		template.ExecuteTemplateParams(w, r, "message", v)
	}
}


