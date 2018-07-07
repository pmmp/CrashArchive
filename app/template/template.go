package template

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/pmmp/CrashArchive/app/crashreport"
)

type Config struct {
	Folder    string `json:"Folder"`
	Extension string `json:"Extension"`
}

var t map[string]*template.Template

func Preload(cfg *Config) error {
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

func ExecuteTemplate(w http.ResponseWriter, name string, data interface{}) error {
	if tmpl, ok := t[name]; ok {
		return tmpl.ExecuteTemplate(w, "base.html", data)
	}
	return ErrorTemplate(w, "", http.StatusInternalServerError)
}

func ErrorTemplate(w http.ResponseWriter, message string, status int) error {
	w.WriteHeader(status)
	if message == "" {
		message = http.StatusText(status)
	}
	return t["error"].ExecuteTemplate(w, "base.html", struct{ Message string }{message})
}

func ExecuteListTemplate(w http.ResponseWriter, reports []crashreport.Report, url string, id int, start int, total int) {
	cnt := len(reports)

	data := map[string]interface{}{
		"RangeStart": 0,
		"RangeEnd":   start + cnt,
		"ShowCount":  cnt,
		"TotalCount": total,
		"SearchUrl":  url,
		"Data":       reports,
		"PrevPage":   0,
		"NextPage":   0,
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
	ExecuteTemplate(w, "list", data)
}
