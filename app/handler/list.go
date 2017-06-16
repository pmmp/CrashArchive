package handler

import (
	"log"
	"net/http"

	"bitbucket.org/intyre/ca-pmmp/app"
	"bitbucket.org/intyre/ca-pmmp/app/crashreport"
	"bitbucket.org/intyre/ca-pmmp/app/template"
)

func ListGet(app *app.App) http.HandlerFunc {
	const name = "list"
	tmpl, err := template.LoadTemplate(name, app.Config.Template)
	if err != nil {
		log.Fatalf("failed to load template %s: %v\n", name, err)
	}

	querySelect := "SELECT id, version, message FROM crash_reports ORDER BY id DESC LIMIT 50"
	queryTotal := "SELECT COUNT(*) FROM crash_reports"

	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var reports []crashreport.Report

		err = app.Database.Select(&reports, querySelect)
		if err != nil {
			log.Println(err)
			return
		}

		var total int
		err = app.Database.Get(&total, queryTotal)
		if err != nil {
			log.Println(err)
			return
		}

		data := make(map[string]interface{})
		data["Data"] = reports
		data["ShowCount"] = len(reports)
		data["TotalCount"] = total
		tmpl.ExecuteTemplate(w, "base.html", data)
	}
}
