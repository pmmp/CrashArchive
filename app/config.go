package app

import (
	"encoding/json"
	"os"

	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
)

type Config struct {
	ListenAddress string           `json:"ListenAddress"`
	Database      *database.Config `json:"Database"`
	Template      *template.Config `json:"Template"`
	SlackURL      string	       `json:"SlackUrl"`
}

func LoadConfig(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config Config
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
