package handler

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/template"
)

func ViewIDGet(w http.ResponseWriter, r *http.Request) {
	reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
	if err != nil {
		template.ErrorTemplate(w, "Please specify a report", http.StatusNotFound)
		return
	}
	report, err := crashreport.ReadFile(int64(reportID))
	if err != nil {
		template.ErrorTemplate(w, "Report not found", http.StatusNotFound)
		return
	}

	report.ParseExtraStuff()

	v := make(map[string]interface{})
	v["Report"] = report
	v["Name"] = "PocketMine-MP"//clean(jsonData["name"].(string))
	v["PocketMineVersion"] = report.Version.Get(true)
	v["AttachedIssue"] = "None"
	v["ReportID"] = reportID

	template.ExecuteTemplate(w, "view", v)
}

var cleanRE = regexp.MustCompile(`[^A-Za-z0-9_\-\.\,\;\:/\#\(\)\\ ]`)

func clean(v string) string {
	return cleanRE.ReplaceAllString(v, "")
}
