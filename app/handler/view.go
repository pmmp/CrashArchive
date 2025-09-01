package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/user"
)

func ViewIDGet(db *database.DB, config *app.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			template.ErrorTemplate(w, r, "Please specify a report", http.StatusNotFound)
			return
		}

		var mainTableInfo crashreport.Report
		err = db.Get(&mainTableInfo, "SELECT reporterName, fork, modified FROM crash_reports WHERE id = ?", reportID)
		if err != nil {
			log.Printf("can't find report %d in database: %v", reportID, err)
			template.ErrorTemplate(w, r, "Report not found", http.StatusNotFound)
			return
		}

		report, expectedAccessToken, err := db.FetchReport(int64(reportID))
		if err != nil {
			log.Printf("error fetching report: %v", err)
			template.ErrorTemplate(w, r, "Report not found", http.StatusNotFound)
			return
		}

		givenAccessToken := r.URL.Query().Get("access_token")
		if config.ViewReportRequiresAuth && !user.GetUserInfo(r).CheckReportAccess(expectedAccessToken, givenAccessToken) {
			template.ErrorTemplate(w, r, "Administrator login is required to view this report", http.StatusUnauthorized)
			return
		}

		v := make(map[string]interface{})
		v["Report"] = report
		v["Name"] = clean(mainTableInfo.ReporterName)
		v["PocketMineVersion"] = report.Version.Get(true)
		v["ReportID"] = reportID
		v["HasDeletePerm"] = user.GetUserInfo(r).HasDeletePerm()
		//do not leak the access token if this instance allows unauthenticated viewing
		v["AccessToken"] = givenAccessToken //needed to allow deleting without admin perms
		v["Fork"] = mainTableInfo.Fork
		v["Modified"] = mainTableInfo.Modified

		issueQueryParams := url.Values{}
		issueQueryParams.Add("title", report.Error.Message)
		if config.GitHubCrashIssueForm {
			issueQueryParams.Add("template", "crash.yml")
			issueQueryParams.Add("crashdump-url", fmt.Sprintf("%s/view/%d", config.Domain, reportID))
		} else {
			issueQueryParams.Add("body", fmt.Sprintf("Link to crashdump: %s/view/%d\n\n### Additional comments\n", config.Domain, reportID))
		}
		v["ReportIssueURL"] = config.GitHubRepo + "/issues/new?" + issueQueryParams.Encode()
		v["GitTreeURL"] = config.GitHubRepo + "/tree/" + report.Data.General.GIT

		template.ExecuteTemplateParams(w, r, "view", v)

		db.Exec("UPDATE crash_reports SET viewed = true WHERE id = ?", reportID)
	}
}

func ViewIDRawGet(db *database.DB, requireAdminToView bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			template.ErrorTemplate(w, r, "Please specify a report", http.StatusNotFound)
			return
		}

		report, expectedAccessToken, err := db.FetchRawReport(int64(reportID))
		if err != nil {
			log.Printf("error fetching report: %v", err)
			template.ErrorTemplate(w, r, "Report not found", http.StatusNotFound)
			return
		}

		if requireAdminToView && !user.GetUserInfo(r).CheckReportAccess(expectedAccessToken, r.URL.Query().Get("access_token")) {
			template.ErrorTemplate(w, r, "This crash archive requires admin login to view reports without an access token", http.StatusUnauthorized)
			return
		}

		var buffer bytes.Buffer
		json.Indent(&buffer, report, "", "    ")
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write(buffer.Bytes())
	}
}

var cleanRE = regexp.MustCompile(`[^A-Za-z0-9_\-\.\,\;\:/\#\(\)\\ +]`)

func clean(v string) string {
	return cleanRE.ReplaceAllString(v, "")
}
