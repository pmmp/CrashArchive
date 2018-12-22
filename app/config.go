package app

import (
	"encoding/json"
	"os"
	"log"

	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
)

type Config struct {
	ListenAddress      string
	Database           *database.Config
	Template           *template.Config
	SlackURLs          []string
	SlackHookInterval  uint32
	PluginBlacklist    []string
	PluginBlacklistMap map[string]string
	IpBanlist          []string
	IpBanlistMap       map[string]string

	//old fields, for backwards compatibility
	SlackURL           string
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

	config.PluginBlacklistMap = make(map[string]string)
	for _, v := range config.PluginBlacklist {
		config.PluginBlacklistMap[v] = v
	}
	config.PluginBlacklist = nil

	config.IpBanlistMap = make(map[string]string)
	for _, v := range config.IpBanlist {
		config.IpBanlistMap[v] = v
	}
	config.IpBanlist = nil

	if config.SlackURLs == nil && config.SlackURL != "" {
		log.Println("`SlackURL` config is deprecated, use `SlackURLs` instead (supports multiple hooks)")
		config.SlackURLs = make([]string, 0)
		config.SlackURLs = append(config.SlackURLs, config.SlackURL)
	}

	return &config, nil
}
