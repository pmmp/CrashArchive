package crashreport

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

// Types ...
const (
	reportBegin = "===BEGIN CRASH DUMP==="
	reportEnd   = "===END CRASH DUMP==="
	currentFormatVersion = 4
)

type errorCleanPattern struct {
	pattern     *regexp.Regexp
	replacement string
}

var errorCleanPatterns []errorCleanPattern = nil

func PrepareErrorCleanPatterns(patterns map[string]string) {
	for pattern, replacement := range patterns {
		var compiled = regexp.MustCompile(pattern)
		errorCleanPatterns = append(errorCleanPatterns, errorCleanPattern{
			pattern:     compiled,
			replacement: replacement,
		})
	}
}

// ParseDate parses  the unix date to time.Time
func (r *CrashReport) parseDate() {
	if r.Data.Time == 0 {
		r.Data.Time = float64(time.Now().Unix())
	}
	r.ReportDate = time.Unix(int64(r.Data.Time), 0)
}

// ParseError ...
func (r *CrashReport) parseError() {
	r.Error.Type = r.Data.Error.Type
	r.Error.Message = r.Data.Error.Message
	r.Error.Line = r.Data.Error.Line
	r.Error.File = r.Data.Error.File
}

// ParseVersion ...
func (r *CrashReport) parseVersion() {
	if r.Data.General.BaseVersion == "" {
		panic(errors.New("version is null"))
	}

	var err error
	general := r.Data.General
	r.Version, err = NewVersionString(general.BaseVersion, general.Build, general.IsDev)
	if err != nil {
		panic(err)
	}
}

// ClassifyMessage ...
func (r *CrashReport) ClassifyMessage() {
	if r.Error.Message == "" {
		panic(errors.New("error message is empty"))
	}

	if strings.HasPrefix(r.Error.Message, "Argument") {
		index1 := strings.Index(r.Error.Message, ", called in")
		if index1 != -1 {
			r.Error.Message = r.Error.Message[0:index1]
		}
	}

	if errorCleanPatterns != nil {
		for _, scrub := range errorCleanPatterns {
			r.Error.Message = scrub.pattern.ReplaceAllString(
				r.Error.Message,
				scrub.replacement,
			)
		}
	}
}

// extractBase64 returns the base64 between ===BEGIN CRASH DUMP=== and ===END CRASH DUMP===
func extractBase64(data string) string {
	reportBeginIndex := strings.Index(data, reportBegin)
	if reportBeginIndex == -1 {
		return data
	}
	reportEndIndex := strings.Index(data, reportEnd)
	if reportEndIndex == -1 {
		return data
	}

	return strings.Trim(data[reportBeginIndex+len(reportBegin):reportEndIndex], "\r\n\t` ")
}

// clean is shoghi magic
func clean(v string) string {
	var re = regexp.MustCompile(`[^A-Za-z0-9_\-\.\,\;\:/\#\(\)\\ ]`)
	return re.ReplaceAllString(v, "")
}

func DecodeCrashReport(data []byte) (*CrashReport, error) {
	jsonBytes, err := JsonFromCrashLog(data)
	if err != nil {
		return nil, err
	}

	return FromJson(jsonBytes)
}

func (r *CrashReport) EncodeCrashReport() ([]byte, error) {
	jsonBytes, err := r.ToJson()
	if err != nil {
		return nil, err
	}

	return JsonToCrashLog(jsonBytes)
}

func JsonToCrashLog(jsonBytes []byte) ([]byte, error) {
	var zlibBuf bytes.Buffer
	zw := zlib.NewWriter(&zlibBuf)
	_, err := zw.Write(jsonBytes)
	if err != nil {
		return nil, err
	}

	zw.Close()

	return []byte(fmt.Sprintf("%s\n%s\n%s", reportBegin, base64.StdEncoding.EncodeToString(zlibBuf.Bytes()), reportEnd)), nil
}

func JsonFromCrashLog(report []byte) ([]byte, error) {
	zlibBytes, err := base64.StdEncoding.DecodeString(extractBase64(string(report)))
	if err != nil {
		return nil, err
	}

	zr, err := zlib.NewReader(bytes.NewReader(zlibBytes))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	jsonBytes, err := ioutil.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %v", err)
	}

	return jsonBytes, nil
}


// FromJson decodes crash report JSON bytes into a CrashReport structure
func FromJson(jsonBytes []byte) (*CrashReport, error) {
	reader := bytes.NewReader(jsonBytes)

	var r CrashReport
	err := json.NewDecoder(reader).Decode(&r.Data)

	if err == nil {
		if r.Data.FormatVersion != currentFormatVersion {
			return nil, fmt.Errorf("incompatible crashdump format version %d", r.Data.FormatVersion)
		}
		switch(r.Data.PluginInvolvement) {
			case PIDirect:
			case PIIndirect:
			case PINone:
				break
			default:
				return nil, fmt.Errorf("unknown plugin involvement \"%s\"", r.Data.PluginInvolvement)
		}
		r.parseDate()
		r.parseError()
		r.parseVersion()
	}

	return &r, err
}

func (r *CrashReport) ToJson() ([]byte, error) {
	var jsonBuf bytes.Buffer
	jw := json.NewEncoder(&jsonBuf)
	err := jw.Encode(r.Data)
	if err != nil {
		return nil, err
	}

	return jsonBuf.Bytes(), nil
}
