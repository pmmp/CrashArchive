package template

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/csrf"
	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/user"
)

type Config struct {
	Folder    string
	Extension string
}

var t map[string]*template.Template
var githubAppClientId string

func Preload(cfg *Config, githubAppClientId_ string) error {
	githubAppClientId = githubAppClientId_

	t = make(map[string]*template.Template)
	abs, _ := filepath.Abs(cfg.Folder)

	layoutFiles, err := filepath.Glob(filepath.Join(abs, "layout", "*."+cfg.Extension))
	if err != nil {
		return err
	}

	pageFiles, err := filepath.Glob(filepath.Join(abs, "*."+cfg.Extension))
	if err != nil {
		return err
	}

	for _, page := range pageFiles {
		templateFiles := append(layoutFiles, page)
		_, fname := filepath.Split(page)

		name := fname[:len(fname)-len(cfg.Extension)-1]
		tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(templateFiles...)
		if err != nil {
			return err
		}
		t[name] = tmpl
	}
	return nil
}

func ExecuteTemplate(w http.ResponseWriter, r *http.Request, name string) error {
	return ExecuteTemplateParams(w, r, name, make(map[string]interface{}))
}

func addContextTemplateParams(data map[string]interface{}, r *http.Request) map[string]interface{} {
	data["ActiveUserName"] = user.GetUserInfo(r).Name
	data["GitHubAppClientId"] = githubAppClientId
	data[csrf.TemplateTag] = csrf.TemplateField(r)
	return data
}

func ExecuteTemplateParams(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) error {
	addContextTemplateParams(data, r)
	if tmpl, ok := t[name]; ok {
		err := tmpl.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			log.Printf("error executing template %s: %v", name, err)
		}
		return err
	}
	return ErrorTemplate(w, r, "", http.StatusInternalServerError)
}

func ErrorTemplate(w http.ResponseWriter, r *http.Request, message string, status int) error {
	w.WriteHeader(status)
	if message == "" {
		message = http.StatusText(status)
	}
	return t["error"].ExecuteTemplate(w, "base.html", addContextTemplateParams(map[string]interface{}{
		"Message": message,
	}, r))
}

type SearchBoxParams struct {
	Message string
	ErrorType string
	PluginInvolvements map[string]string
	Plugin string
	Versions map[string]string
	Duplicates bool
	Forks bool
	Modified bool
}

func ExecuteListTemplate(w http.ResponseWriter, r *http.Request, reports []crashreport.Report, url string, id int, start int, total int, searchBoxParams *SearchBoxParams, knownVersions []string) error {
	cnt := len(reports)

	log.Printf("searchbox params: %+v", searchBoxParams)


	data := map[string]interface{}{
		"RangeStart": 0,
		"RangeEnd":   start + cnt,
		"ShowCount":  cnt,
		"TotalCount": total,
		"SearchUrl":  url,
		"Data":       reports,
		"PrevPage":   0,
		"NextPage":   0,
		"Search":     searchBoxParams,
		"PluginInvolvementOptions": crashreport.PluginInvolvementOptions,
		"KnownVersions": knownVersions,
	}

	if cnt > 0 {
		data["RangeStart"] = start + 1
	}

	if start > 0 {
		data["PrevPage"] = id - 1
	}

	if start+cnt < total {
		data["NextPage"] = id + 1
	}
	return ExecuteTemplateParams(w, r, "list", data)
}
