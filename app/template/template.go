package template

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"net/http"
	"github.com/pmmp/CrashArchive/app/crashreport"
)

type Config struct {
	Folder    string
	Extension string
}

func LoadTemplate(name string, cfg *Config) (*template.Template, error) {
	templateList := []string{
		"base",
		name,
		"footer",
		"menu",
	}

	for i, name := range templateList {
		path, err := filepath.Abs(cfg.Folder + string(os.PathSeparator) + name + "." + cfg.Extension)
		if err != nil {
			log.Fatalf("Template Path Error: %v\n", path)
			return nil, err
		}
		templateList[i] = path
	}

	tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(templateList...)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func ExecuteErrorTemplate(w http.ResponseWriter, cfg *Config, message string, backURL string) error {
	errorTmpl, err := LoadTemplate("error", cfg)
	if err != nil {
		return err
	}
	errorTmpl.ExecuteTemplate(w, "base.html", map[string]interface{}{
		"Message": message,
		"URL":     backURL,
	})
	return nil
}

func ExecuteListTemplate(w http.ResponseWriter, cfg *Config, reports []crashreport.Report, searchUrl string, pageId int, rangeStart int, total int) {
	const templateName = "list"

	tmpl, err := LoadTemplate(templateName, cfg)
	if err != nil {
		log.Printf("failed to load template %s: %v\n", templateName, err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	reportCount := len(reports)

	data := make(map[string]interface{})
	data["Data"] = reports
	if reportCount <= 0 {
		data["RangeStart"] = 0
	} else {
		data["RangeStart"] = rangeStart + 1
	}

	data["RangeEnd"] = rangeStart + reportCount
	data["ShowCount"] = reportCount
	data["TotalCount"] = total
	data["SearchUrl"] = searchUrl
	if rangeStart <= 0 {
		data["PrevPage"] = 0
	} else {
		data["PrevPage"] = pageId - 1
	}
	if rangeStart + reportCount >= total {
		data["NextPage"] = 0
	} else {
		data["NextPage"] = pageId + 1
	}
	tmpl.ExecuteTemplate(w, "base.html", data)
}
