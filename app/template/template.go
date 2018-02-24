package template

import (
	"html/template"
	"io"
	"net/http"
	"path/filepath"

	"github.com/pmmp/CrashArchive/app/crashreport"
)

type Config struct {
	Folder    string
	Extension string
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

func ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	if tmpl, ok := t[name]; ok {
		return tmpl.ExecuteTemplate(w, "base.html", data)
	}
	return ErrorTemplate(w, "whoops")
}

func ErrorTemplate(w io.Writer, message string) error {
	return t["error"].Execute(w, struct{ Message string }{message})
}

func ExecuteListTemplate(w http.ResponseWriter, reports []crashreport.Report, url string, id int, start int, total int) {
	cnt := len(reports)

	data := map[string]interface{}{
		"RangeEnd":   start + cnt,
		"ShowCount":  cnt,
		"TotalCount": total,
		"SearchUrl":  url,
		"Data":       reports,
		"RangeStart": 0,
		"PrevPage":   0,
	}

	if cnt > 0 {
		data["RangeStart"] = start + 1
	}

	if start > 0 {
		data["PrevPage"] = id - 1
	}

	if start+cnt >= total {
		data["NextPage"] = 0
	} else {
		data["NextPage"] = id + 1
	}
	ExecuteTemplate(w, "list", data)
}
