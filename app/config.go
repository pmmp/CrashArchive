package app

import (
	"encoding/json"
	"log"
	"os"
	"regexp"

	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/template"
)

type Config struct {
	Domain            string
	ListenAddress     string
	Public            bool
	Database          *database.Config
	Template          *template.Config
	SlackURLs         []string
	SlackHookInterval uint32

	PluginBlacklist    []string
	PluginBlacklistMap map[string]string

	SubmitAllowedIps    []string
	SubmitAllowedIpsMap map[string]string

	IpBanlist    []string
	IpBanlistMap map[string]string

	ErrorCleanPatterns     map[string]string
	ErrorBlacklistPatterns []string

	CompiledErrorBlacklistPatterns []*regexp.Regexp

	//old fields, for backwards compatibility
	SlackURL string
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

	config.SubmitAllowedIpsMap = make(map[string]string)
	for _, v := range config.SubmitAllowedIps {
		config.SubmitAllowedIpsMap[v] = v
	}
	config.SubmitAllowedIps = nil

	if config.SlackURLs == nil && config.SlackURL != "" {
		log.Println("`SlackURL` config is deprecated, use `SlackURLs` instead (supports multiple hooks)")
		config.SlackURLs = make([]string, 0)
		config.SlackURLs = append(config.SlackURLs, config.SlackURL)
	}

	for _, patternStr := range config.ErrorBlacklistPatterns {
		var compiled = regexp.MustCompile(patternStr)
		config.CompiledErrorBlacklistPatterns = append(config.CompiledErrorBlacklistPatterns, compiled)
	}

	return &config, nil
}
