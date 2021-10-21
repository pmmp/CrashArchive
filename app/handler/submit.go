package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/template"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/webhook"
)

func SubmitGet(w http.ResponseWriter, r *http.Request) {
	template.ExecuteTemplate(w, r, "submit")
}

func SubmitPost(db *database.DB, wh *webhook.Webhook, config *app.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, banned := config.IpBanlistMap[r.RemoteAddr]; banned {
			log.Printf("rejected submission from banned IP: %s\n", r.RemoteAddr);
			sendError(w, r, "", http.StatusTeapot, true)
			return
		}
		if r.FormValue("report") != "yes" {
			http.Redirect(w, r, "/submit", http.StatusMovedPermanently)
			return
		}

		if err := r.ParseMultipartForm(1024 * 256); err != nil {
			http.Redirect(w, r, "/submit", http.StatusMovedPermanently)
			return
		}

		reportStr, err := ParseMultipartForm(r.MultipartForm)
		if err != nil {
			http.Redirect(w, r, "/submit", http.StatusMovedPermanently)
			return
		}

		isAPI := strings.HasSuffix(r.RequestURI, "/api")

		defer func() {
			if recovered := recover(); recovered != nil {
				err, ok := recovered.(error)
				if !ok {
					err = fmt.Errorf("%v", recovered)
				}

				log.Printf("got invalid crash report from: %s (%v)", r.RemoteAddr, err)
				sendError(w, r, "This crash report is not valid", http.StatusUnprocessableEntity, isAPI)
			}
		}()

		report, err := crashreport.DecodeCrashReport([]byte(reportStr))
		if err != nil {
			//this panic will be recovered in the above deferred function
			panic(err)
		}

		if report.Data.General.Build != 0 && uint32(report.Data.General.Build) < config.MinBuildNumber {
			log.Printf("too-old version %d, minimum is %d, from: %s\n", report.Data.General.Build, config.MinBuildNumber, r.RemoteAddr)
			sendError(w, r, "This crash report is from an outdated version", http.StatusUnprocessableEntity, isAPI)

			return
		}

		if report.Data.General.Name != "PocketMine-MP" {
			log.Printf("spoon detected from: %s\n", r.RemoteAddr)
			sendError(w, r, "", http.StatusTeapot, isAPI)
			return
		}

		if report.Data.General.GIT == strings.Repeat("00", 20) || strings.HasSuffix(report.Data.General.GIT, "-dirty") {
			log.Printf("invalid git hash %s in report from: %s\n", report.Data.General.GIT, r.RemoteAddr)
			sendError(w, r, "", http.StatusTeapot, isAPI)
			return
		}

		pluginsList, ok := report.Data.Plugins.(map[string]interface{})
		if ok {
			for v, _ := range pluginsList {
				if _, blacklisted := config.PluginBlacklistMap[v]; blacklisted {
					log.Printf("blacklisted plugin \"%s\" in report from: %s\n", v, r.RemoteAddr)
					sendError(w, r, "", http.StatusTeapot, isAPI)
					return
				}
			}
		}

		for _, pattern := range(config.CompiledErrorBlacklistPatterns) {
			if pattern.MatchString(report.Data.Error.Message) {
				log.Printf("blacklisted error pattern match in report from: %s", "", r.RemoteAddr)
				sendError(w, r, "This crashdump is blacklisted", http.StatusUnprocessableEntity, isAPI)
				return
			}
		}

		report.ClassifyMessage() //we use the classified message to get a better hit on dupes without useless information

		dupes, err := db.CheckDuplicate(report)
		report.Duplicate = dupes
		if dupes {
			snippet := report.Data.Error.Message
			if len(snippet) > 80 {
				snippet = snippet[:80]
			}
			log.Printf("duplicate report from: %s, message is \"%s\"", r.RemoteAddr, snippet)
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		jsonBytes, _ := crashreport.JsonFromCrashLog([]byte(reportStr)) //this should have given us an error earlier if it was going to

		id, err := db.InsertReport(report, name, email, jsonBytes)
		if err != nil {
			log.Printf("failed to insert report into database: %v", err)
			sendError(w, r, "", http.StatusInternalServerError, isAPI)
			return
		}

		if wh != nil {
			if !report.Duplicate {
				go wh.Post(webhook.ReportListEntry{
					ReportId: uint64(id),
					Message: report.Error.Message,
				})
			} else {
				wh.BumpDupeCounter()
			}
		}

		if isAPI {
			jsonResponse(w, map[string]interface{}{
				"crashId":  id,
				"crashUrl": fmt.Sprintf("%s/view/%d", config.Domain, id),
			})
		} else {
			http.Redirect(w, r, fmt.Sprintf("/view/%d", id), http.StatusMovedPermanently)
		}

	}
}

func sendError(w http.ResponseWriter, r *http.Request, message string, status int, isAPI bool) {
	if isAPI {
		w.WriteHeader(status)
		if message == "" {
			message = http.StatusText(status)
		}
		jsonResponse(w, map[string]interface{}{
			"error": message,
		})
	} else {
		template.ErrorTemplate(w, r, message, status)
	}
}

func jsonResponse(w http.ResponseWriter, data map[string]interface{}) {
	json.NewEncoder(w).Encode(data)
}

func ParseMultipartForm(form *multipart.Form) (string, error) {
	var report string
	if reportPaste, ok := form.Value["reportPaste"]; ok && reportPaste[0] != "" {
		report = reportPaste[0]
	} else if reportFile, ok := form.File["reportFile"]; ok && reportFile[0] != nil {
		f, err := reportFile[0].Open()
		if err != nil {
			return "", errors.New("could not open file")
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return "", errors.New("could not read file")
		}
		f.Close()
		report = string(b)
	}

	return report, nil
}
