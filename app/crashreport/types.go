package crashreport

import (
	"time"
)

const (
	PINone     = "none"
	PIIndirect = "indirect"
	PIDirect   = "direct"
)

var PluginInvolvementStrings = map[string]string{
	PINone:     "None",
	PIIndirect: "Indirect",
	PIDirect:   "Direct",
}

// CrashReport ...
type CrashReport struct {
	Duplicate    bool

	Data       *ReportData
	ReportDate time.Time

	Version    *VersionString

	Error ReportError
}

// ReportData ...
type ReportData struct {
	Time              int64
	FormatVersion     int64 `json:"format_version"`
	Plugin            string
	PluginInvolvement string `json:"plugin_involvement"`
	General struct {
		Name        string
		BaseVersion string `json:"base_version"`
		Build       int
		IsDev       bool `json:"is_dev"`
		Protocol    int
		GIT         string
		Raklib      string
		Uname       string
		PHP         string
		Zend        string
		PHPOS       string `json:"php_os"`
		OS          string
	}
	Error            ReportError
	Code             map[string]string
	Plugins          interface{} `json:"plugins,omitempty"`
	PocketmineYML    string      `json:"pocketmine.yml"`
	ServerProperties string      `json:"server.properties"`
	Trace            []string
}

type ReportError struct {
	Type    string
	Message string
	Line    int
	File    string
}

// Report ...
type Report struct {
	ID                int `db:"id"`
	Plugin            string
	PluginInvolvement string `db:"pluginInvolvement"`
	Version           string
	Build             int
	File              string
	Message           string
	Line              int
	Type              string
	OS                string
	SubmitDate        int64  `db:"submitDate"`
	ReportDate        int64  `db:"reportDate"`
	Duplicate         bool
	ReporterName      string `db:"reporterName"`
	ReporterEmail     string `db:"reporterEmail"`
}
