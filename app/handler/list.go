package handler

import (
	"log"
	"net/http"

	"fmt"
	"strconv"

	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
)

func ListGet(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ListFilteredReports(w, r, db, "WHERE duplicate = false")
	}
}

const pageSize = 50

func ListFilteredReports(w http.ResponseWriter, r *http.Request, db *database.DB, filter string, filterParams ...interface{}) {
	var err error

	var total int

	queryCount := fmt.Sprintf("SELECT COUNT(*) FROM crash_reports %s", filter)
	err = db.Get(&total, queryCount, filterParams...)
	if err != nil {
		log.Println(err)
		log.Println(queryCount)
		template.ErrorTemplate(w, "", http.StatusInternalServerError)
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
	querySelect := fmt.Sprintf("SELECT id, version, message FROM crash_reports %s ORDER BY id DESC LIMIT %d, %d", filter, rangeStart, pageSize)
	err = db.Select(&reports, querySelect, filterParams...)
	if err != nil {
		log.Println(err)
		log.Println(querySelect)
		template.ErrorTemplate(w, "", http.StatusInternalServerError)
		return
	}

	template.ExecuteListTemplate(w, reports, r.URL.String(), pageId, rangeStart, total)
}
