package handler

import (
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/pressly/chi"

	"bitbucket.org/intyre/ca-pmmp/app"
	"bitbucket.org/intyre/ca-pmmp/app/crashreport"
	"bitbucket.org/intyre/ca-pmmp/app/template"
)

func ViewIDGet(app *app.App) http.HandlerFunc {
	const name = "view"
	tmpl, err := template.LoadTemplate(name, app.Config.Template)
	if err != nil {
		log.Fatalf("failed to load template %s: %v\n", name, err)
	}

	querySelect := `SELECT * FROM crash_reports WHERE id = ?`
	// queryTotal := `SELECT COUNT(*) FROM crash_reports`

	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			log.Printf("%v\n", err)
			return
		}

		data := crashreport.Report{}
		err = app.Database.Get(&data, querySelect, reportID)
		if err != nil {
			log.Println(err)
			return
		}
		report, jsonData, err := crashreport.ReadFile(int64(reportID))
		if err != nil {
			log.Println(err)
			return
		}

		v := make(map[string]interface{}, 0)
		v["Data"] = data
		v["Report"] = report
		v["Name"] = clean(jsonData["name"].(string))
		v["PocketMineVersion"] = report.Version.Get(true)
		v["AttachedIssue"] = "None"
		tmpl.ExecuteTemplate(w, "base.html", v)
	}
}

func clean(v string) string {
	var re = regexp.MustCompile("[^A-Za-z0-9_\\-\\.\\,\\;\\:/\\#\\(\\)\\\\ ]")
	return re.ReplaceAllString(v, "")
}
