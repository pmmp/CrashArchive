package crashreport

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
)

// Types ...
const (

	reportBegin = "===BEGIN CRASH DUMP==="
	reportEnd   = "===END CRASH DUMP==="
)

func Parse(data string) (*CrashReport, error) {
	var r CrashReport

	if err := r.ReadCrashLog(data); err != nil {
		return nil, fmt.Errorf("failed to read compressed data: %v", err)
	}

	return &r, nil
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

// ReadCrashLog reads the base64 encoded and zlib compressed report
func (r *CrashReport) ReadCrashLog(report string) error {
	zlibBytes, err := base64.StdEncoding.DecodeString(extractBase64(report))
	if err != nil {
		return err
	}

	return r.ReadZlib(zlibBytes)
}

func (r *CrashReport) ReadZlib(zlibBytes []byte) error {
	zr, err := zlib.NewReader(bytes.NewReader(zlibBytes))
	if err != nil {
		return err
	}
	defer zr.Close()

	err = json.NewDecoder(zr).Decode(&r.Data)
	if err == nil {
		r.parseDate()
		r.parseError()
		r.parseVersion()
		r.classifyMessage()
	}

	return err
}

// WriteCrashLog generates a crashdump log file
func (r *CrashReport) WriteCrashLog() string {
	return fmt.Sprintf("%s\n%s\n%s", reportBegin, base64.StdEncoding.EncodeToString(r.WriteZlib()), reportEnd)
}

// WriteZlib json-encodes and zlib-compresses the crash report
func (r *CrashReport) WriteZlib() []byte {
	var jsonBuf bytes.Buffer
	jw := json.NewEncoder(&jsonBuf)
	err := jw.Encode(r.Data)
	if err != nil {
		log.Fatal(err)
	}

	var zlibBuf bytes.Buffer
	zw := zlib.NewWriter(&zlibBuf)
	_, err = zw.Write(jsonBuf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	zw.Close()

	return zlibBuf.Bytes()
}
