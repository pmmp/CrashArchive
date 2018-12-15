package handler

import (
	"log"
	"net/url"
	"net/http"

	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
)

func ListGet(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter, filterParams, err := buildSearchQuery(r.URL.Query())
		if err != nil {
			template.ErrorTemplate(w, "", http.StatusBadRequest)
			return
		}
		log.Printf("search generated query: %s\n", filter)
		ListFilteredReports(w, r, db, filter, filterParams...)
	}
}

const pageSize = 50

func parseUintParam(v url.Values, paramName string, defaultValue uint64) (uint64, error) {
	param := v.Get(paramName)
	if param != "" {
		retval, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("Invalid value for search parameter \"%s\"", paramName)
		}

		return retval, nil
	}

	return defaultValue, nil
}

func buildSearchQuery(params url.Values) (string, []interface{}, error) {
	filters := make([]string, 0)
	filterParams := make([]interface{}, 0)
	var filter string

	if params.Get("duplicates") != "true" {
		filters = append(filters, "duplicate = false")
	}

	if params.Get("min") != "" || params.Get("max") != "" {
		//check ranges
		filterMinId, err := parseUintParam(params, "min", 0)
		if err != nil {
			return "", nil, err
		}
		filterMaxId, err := parseUintParam(params, "max", math.MaxUint64)
		if err != nil {
			return "", nil, err
		}

		if filterMinId > filterMaxId {
			return "", nil, errors.New("Invalid min/max ID bounds")
		}

		filters = append(filters, "id BETWEEN ? AND ?")
		filterParams = append(filterParams, filterMinId, filterMaxId)
	}

	//filter by message
	message := params.Get("message")
	if message != "" {
		filters = append(filters, "message LIKE ?")
		filterParams = append(filterParams, "%" + message + "%")
	}

	cause := params.Get("cause")
	if cause == "core" {
		filters = append(filters, "plugin = \"\"")
	} else if cause == "plugin" {
		//filter by plugin
		plugin := params.Get("plugin")
		if plugin != "" {
			filters = append(filters, "plugin = ?")
			filterParams = append(filterParams, plugin)
		} else { //any plugin but not core crashes
			filters = append(filters, "plugin <> \"\"")
		}
	}

	//filter by build number
	if params.Get("build") != "" {
		buildID, err := parseUintParam(params, "build", math.MaxUint64)
		if err != nil {
			return "", nil, err
		}

		operator := "="
		typ := params.Get("buildtype")
		if typ == "greater" {
			operator = ">"
		} else if typ == "less" {
			operator = "<"
		}

		filters = append(filters, fmt.Sprintf("build %s ?", operator))
		filterParams = append(filterParams, buildID)
	}

	filter = strings.Join(filters[:], " AND ")
	if filter != "" {
		filter = "WHERE " + filter
	}
	return filter, filterParams, nil
}

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
	querySelect := fmt.Sprintf("SELECT id, version, plugin, message FROM crash_reports %s ORDER BY id DESC LIMIT %d, %d", filter, rangeStart, pageSize)
	err = db.Select(&reports, querySelect, filterParams...)
	if err != nil {
		log.Println(err)
		log.Println(querySelect)
		template.ErrorTemplate(w, "", http.StatusInternalServerError)
		return
	}

	template.ExecuteListTemplate(w, reports, r.URL.String(), pageId, rangeStart, total)
}
