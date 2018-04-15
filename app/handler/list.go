package handler

import (
	"log"
	"net/url"
	"net/http"

	"fmt"
	"math"
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

func parseUintParam(v url.Values, paramName string, defaultValue uint64, w http.ResponseWriter) (uint64, error) {
	param := v.Get(paramName)
	if param != "" {
		retval, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			log.Println(err)
			template.ErrorTemplate(w, "", http.StatusBadRequest)
			return 0, err
		}

		return retval, nil
	}

	return defaultValue, nil
}

func ListFilteredReports(w http.ResponseWriter, r *http.Request, db *database.DB, filter string, filterParams ...interface{}) {
	var err error

	params := r.URL.Query()

	filterMinId, err := parseUintParam(params, "min", 0, w)
	if err != nil {
		return
	}
	filterMaxId, err := parseUintParam(params, "max", math.MaxUint64, w)
	if err != nil {
		return
	}

	if filterMinId > filterMaxId {
		log.Println("request tried to ask for min bound larger than max bound")
		template.ErrorTemplate(w, "", http.StatusBadRequest)
		return
	}

	if filter != "" {
		filter = fmt.Sprintf("%s AND id BETWEEN ? AND ?", filter)
	} else {
		filter = "WHERE id BETWEEN ? AND ?"
	}
	filterParams = append(filterParams, filterMinId, filterMaxId)

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
