package handler

import (
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi"

	"bitbucket.org/intyre/ca-pmmp/app"
	"bitbucket.org/intyre/ca-pmmp/app/crashreport"
	"bitbucket.org/intyre/ca-pmmp/app/template"
)

func ViewIDGet(app *app.App) http.HandlerFunc {
	name := "view"
	tmpl, err := template.LoadTemplate(name, app.Config.Template)
	if err != nil {
		log.Fatalf("failed to load template %s: %v\n", name, err)
	}

	name = "error"
	errorTmpl, err := template.LoadTemplate(name, app.Config.Template)
	if err != nil {
		log.Fatalf("failed to load template %s: %v\n", name, err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			errorTmpl.ExecuteTemplate(w, "base.html", map[string]interface{}{
				"Message": "Please specify a report",
				"URL":     "/",
			})
			return
		}
		report, jsonData, err := crashreport.ReadFile(int64(reportID))
		if err != nil {
			errorTmpl.ExecuteTemplate(w, "base.html", map[string]interface{}{
				"Message": "Report not found",
				"URL":     "/",
			})
			return
		}

		v := make(map[string]interface{})
		v["Report"] = report
		v["Name"] = clean(jsonData["name"].(string))
		v["PocketMineVersion"] = report.Version.Get(true)
		v["AttachedIssue"] = "None"
		v["ReportID"] = reportID

		if err = tmpl.ExecuteTemplate(w, "base.html", v); err != nil {
			return
		}
	}
}

var cleanRE = regexp.MustCompile(`[^A-Za-z0-9_\-\.\,\;\:/\#\(\)\\ ]`)

func clean(v string) string {
	return cleanRE.ReplaceAllString(v, "")
}
