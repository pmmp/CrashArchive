package crashreport

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"io/ioutil"
)

// Types ...
const (

	reportBegin = "===BEGIN CRASH DUMP==="
	reportEnd   = "===END CRASH DUMP==="
)

func DecodeCrashReport(data string) (*CrashReport, error) {
	jsonBytes, err := JsonFromCrashLog(data)
	if err != nil {
		return nil, err
	}

	return FromJson(jsonBytes)
}

func (r *CrashReport) EncodeCrashReport() (string, error) {
	jsonBytes, err := r.ToJson()
	if err != nil {
		return "", err
	}

	return JsonToCrashLog(jsonBytes)
}

// ParseDate parses  the unix date to time.Time
func (r *CrashReport) parseDate() {
	if r.Data.Time == 0 {
		r.Data.Time = time.Now().Unix()
	}
	r.ReportDate = time.Unix(r.Data.Time, 0)
}

// ParseError ...
func (r *CrashReport) parseError() {
	switch plugin := r.Data.Plugin.(type) {
	case bool:
		r.CausedByPlugin = plugin
	case string:
		r.CausingPlugin = clean(plugin)
		r.CausedByPlugin = true
	}

	r.Error.Type = r.Data.Error.Type
	r.Error.Message = r.Data.Error.Message
	r.Error.Line = r.Data.Error.Line
	r.Error.File = r.Data.Error.File
}

// ParseVersion ...
func (r *CrashReport) parseVersion() {
	if r.Data.General.Version == "" {
		panic(errors.New("version is null"))
	}

	general := r.Data.General
	r.APIVersion = general.API
	r.Version = NewVersionString(general.Version, general.Build)
}

// ClassifyMessage ...
func (r *CrashReport) classifyMessage() {
	if r.Error.Message == "" {
		panic(errors.New("error message is empty"))
	}

	if strings.HasPrefix(r.Error.Message, "Argument") {
		index1 := strings.Index(r.Error.Message, ", called in")
		if index1 != -1 {
			r.Error.Message = r.Error.Message[0:index1]
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


func JsonToCrashLog(jsonBytes []byte) (string, error) {
	var zlibBuf bytes.Buffer
	zw := zlib.NewWriter(&zlibBuf)
	_, err := zw.Write(jsonBytes)
	if err != nil {
		return "", err
	}

	zw.Close()

	return fmt.Sprintf("%s\n%s\n%s", reportBegin, base64.StdEncoding.EncodeToString(zlibBuf.Bytes()), reportEnd), nil
}

func JsonFromCrashLog(report string) ([]byte, error) {
	zlibBytes, err := base64.StdEncoding.DecodeString(extractBase64(report))
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
	reader := bytes.NewReader([]byte(jsonBytes))

	var r CrashReport
	err := json.NewDecoder(reader).Decode(&r.Data)

	if err == nil {
		r.parseDate()
		r.parseError()
		r.parseVersion()
		r.classifyMessage()
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
