package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/go-chi/chi"
)

func DownloadGet(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
		if err != nil {
			template.ExecuteErrorTemplate(w, app.Config.Template, "Please specify a report")
			return
		}

		_, jsonData, err := crashreport.ReadFile(int64(reportID))
		if err != nil {
			template.ExecuteErrorTemplate(w, app.Config.Template, "Report not found")
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.log", reportID))
		w.Header().Set("Content-Length", strconv.Itoa(len(jsonData["report"].(string))))
		w.Write([]byte(jsonData["report"].(string)))
	}
}
