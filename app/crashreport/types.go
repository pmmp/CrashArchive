package crashreport

import "time"

// CrashReport ...
type CrashReport struct {
	ReportType   string
	ErrorMessage string
	Valid        bool
	Duplicate    bool

	CausedByPlugin bool
	CausingPlugin  string

	Data       *ReportData
	ReportDate time.Time

	Version    *VersionString
	APIVersion string

	Error ReportError
}

// ReportData ...
type ReportData struct {
	Time    int64
	Plugin  interface{}
	General struct {
		Name     string
		Version  string
		Build    int
		Protocol int
		API      string
		GIT      string
		Raklib   string
		Uname    string
		PHP      string
		Zend     string
		PHPOS    string
		OS       string
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
	ID         int `db:"id"`
	Plugin     string
	Version    string
	Build      int
	File       string
	Message    string
	Line       int
	Type       string
	OS         string
	ReportType string `db:"reportType"`
	SubmitDate int64  `db:"submitDate"`
	ReportDate int64  `db:"reportDate"`
	Duplicate  bool
}
