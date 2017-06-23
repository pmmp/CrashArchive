package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"bitbucket.org/intyre/ca-pmmp/app"
	"bitbucket.org/intyre/ca-pmmp/app/crashreport"
	"bitbucket.org/intyre/ca-pmmp/app/template"
	"github.com/go-chi/chi"
)

func DownloadGet(app *app.App) http.HandlerFunc {
	name := "error"
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

		_, jsonData, err := crashreport.ReadFile(int64(reportID))
		if err != nil {
			errorTmpl.ExecuteTemplate(w, "base.html", map[string]interface{}{
				"Message": "Report not found",
				"URL":     "/",
			})
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.log", reportID))
		w.Header().Set("Content-Length", strconv.Itoa(len(jsonData["report"].(string))))
		w.Write([]byte(jsonData["report"].(string)))
	}

}
