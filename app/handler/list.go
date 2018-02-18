package handler

import (
	"log"
	"net/http"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/go-chi/chi"
	"strconv"
	"fmt"
)

func ListGet(app *app.App) http.HandlerFunc {
	const name = "list"
	const pageSize = 50
	tmpl, err := template.LoadTemplate(name, app.Config.Template)
	if err != nil {
		log.Fatalf("failed to load template %s: %v\n", name, err)
	}

	querySelect := "SELECT id, version, message FROM crash_reports ORDER BY id DESC LIMIT %d, %d"
	queryTotal := "SELECT COUNT(*) FROM crash_reports"

	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		var total int
		err = app.Database.Get(&total, queryTotal)
		if err != nil {
			log.Println(err)
			return
		}

		var pageId int

		pageParam := chi.URLParam(r, "pageID")
		if pageParam != "" {
			pageId, err = strconv.Atoi(pageParam)
			if err != nil || pageId < 0 || (pageId - 1) * pageSize > total {
				err = template.ExecuteErrorTemplate(w, app.Config.Template, "Page not found", "/list")
				if err != nil {
					log.Println(err)
				}
				return
			}
		} else {
			pageId = 1
		}

		rangeStart := (pageId - 1) * pageSize

		var reports []crashreport.Report

		err = app.Database.Select(&reports, fmt.Sprintf(querySelect, rangeStart, pageSize))
		if err != nil {
			log.Println(err)
			return
		}

		data := make(map[string]interface{})
		data["Data"] = reports
		data["RangeStart"] = rangeStart + 1
		data["RangeEnd"] = rangeStart + len(reports)
		data["ShowCount"] = len(reports)
		data["TotalCount"] = total
		if rangeStart <= 0 {
			data["PrevPage"] = 0
		} else {
			data["PrevPage"] = pageId - 1
		}
		if rangeStart + pageSize >= total {
			data["NextPage"] = 0
		} else {
			data["NextPage"] = pageId + 1
		}
		tmpl.ExecuteTemplate(w, "base.html", data)
	}
}
