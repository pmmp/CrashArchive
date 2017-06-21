package template

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
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
