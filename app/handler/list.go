package handler

import (
	"log"
	"net/http"

	"fmt"
	"strconv"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/template"
)

func ListGet(app *app.App) http.HandlerFunc {
	const pageSize = 50

	querySelect := "SELECT id, version, message FROM crash_reports WHERE duplicate = false ORDER BY id DESC LIMIT %d, %d"
	queryTotal := "SELECT COUNT(*) FROM crash_reports WHERE duplicate = false"

	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		var total int
		err = app.Database.Get(&total, queryTotal)
		if err != nil {
			log.Println(err)
			return
		}

		var pageId int

		params := r.URL.Query()

		pageParam := params.Get("page")
		if pageParam != "" {
			pageId, err = strconv.Atoi(pageParam)
			if err != nil || pageId <= 0 || (pageId-1)*pageSize > total {
				template.ErrorTemplate(w, "", http.StatusNotFound)
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

		template.ExecuteListTemplate(w, reports, r.URL.String(), pageId, rangeStart, total)
	}
}
