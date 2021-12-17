package crashreport

import (
	"encoding/json"
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

type MapStringStringAllowsEmptyArray map[string]string

func (this *MapStringStringAllowsEmptyArray) UnmarshalJSON(data []byte) error {
	var decoded interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	arr, ok := decoded.([]interface{})
	if ok && len(arr) == 0 {
		return nil //empty arrays should be treated as empty maps
	}

	casted := decoded.(map[string]interface{})
	*this = make(MapStringStringAllowsEmptyArray)
	for k, v := range casted {
		(*this)[k] = v.(string)
	}
	return nil
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
	Time              float64
	Uptime            float64
	FormatVersion     int64 `json:"format_version"`
	Plugin            string
	PluginInvolvement string `json:"plugin_involvement"`
	General struct {
		Name              string
		BaseVersion       string `json:"base_version"`
		Build             int
		IsDev             bool `json:"is_dev"`
		Protocol          int
		GIT               string
		Uname             string
		PHP               string
		Zend              string
		PHPOS             string `json:"php_os"`
		OS                string
		ComposerLibraries map[string]string `json:"composer_libraries"`
	}
	Error            ReportError
	Code             MapStringStringAllowsEmptyArray
	Plugins          interface{} `json:"plugins,omitempty"`
	Extensions       MapStringStringAllowsEmptyArray
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
