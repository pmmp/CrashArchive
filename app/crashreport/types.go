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
	Duplicate bool `json:"duplicate"`

	Data       *ReportData `json:"data"`
	ReportDate time.Time   `json:"report_date"`

	Version *VersionString `json:"version"`

	Error ReportError `json:"error"`
}

// ReportData ...
type ReportData struct {
	Time              float64 `json:"time"`
	Uptime            float64 `json:"uptime"`
	FormatVersion     int64   `json:"format_version"`
	Plugin            string  `json:"plugin"`
	PluginInvolvement string  `json:"plugin_involvement"`
	General           struct {
		Name              string            `json:"name"`
		BaseVersion       string            `json:"base_version"`
		Build             int               `json:"build"`
		IsDev             bool              `json:"is_dev"`
		Protocol          int               `json:"protocol"`
		GIT               string            `json:"git"`
		Uname             string            `json:"uname"`
		PHP               string            `json:"php"`
		Zend              string            `json:"zend"`
		PHPOS             string            `json:"php_os"`
		OS                string            `json:"os"`
		ComposerLibraries map[string]string `json:"composer_libraries"`
	}
	Error            ReportError                     `json:"error"`
	Code             MapStringStringAllowsEmptyArray `json:"code"`
	Plugins          interface{}                     `json:"plugins,omitempty"`
	PocketmineYML    string                          `json:"pocketmine.yml"`
	ServerProperties string                          `json:"server.properties"`
	Trace            []string                        `json:"trace"`
}

type ReportError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Line    int    `json:"line"`
	File    string `json:"file"`
}

// Report ...
type Report struct {
	ID                int    `json:"id" db:"id"`
	Plugin            string `json:"plugin"`
	PluginInvolvement string `json:"plugin_involvement" db:"pluginInvolvement"`
	Version           string `json:"version"`
	Build             int    `json:"build"`
	File              string `json:"file"`
	Message           string `json:"message"`
	Line              int    `json:"line"`
	Type              string `json:"type"`
	OS                string `json:"os"`
	SubmitDate        int64  `json:"submit_date" db:"submitDate"`
	ReportDate        int64  `json:"report_date" db:"reportDate"`
	Duplicate         bool   `json:"duplicate"`
	ReporterName      string `db:"reporterName"`
	ReporterEmail     string `db:"reporterEmail"`
}
