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
	TypeGeneric         = "generic"
	TypeOperandType     = "operand_type"
	TypeOutOfMemory     = "out_of_memory"
	TypeUndefinedCall   = "undefined_call"
	TypeClassVisibility = "class_visibility"
	TypeInvalidArgument = "invalid_argument"
	TypeClassNotFound   = "class_not_found"
	TypeUnknown         = "unknown"

	reportBegin = "===BEGIN CRASH DUMP==="
	reportEnd   = "===END CRASH DUMP==="
)

func Parse(data string) (*CrashReport, error) {
	var r CrashReport
	r.ReportType = TypeGeneric
	r.CausedByPlugin = false

	if err := r.ReadCompressed(data); err != nil {
		return nil, fmt.Errorf("failed to read compressed data: %v", err)
	}

	r.parseDate()
	r.parseError()
	r.parseVersion()
	r.classifyMessage()
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
	m := r.Error.Message
	if strings.HasPrefix(m, "Unsupported operand types") {
		r.ReportType = TypeOperandType
	} else if strings.HasPrefix(m, "Allowed memory size of") {
		r.ReportType = TypeOutOfMemory
	} else if strings.HasPrefix(m, "Call to undefined") ||
		strings.HasPrefix(m, "Call to a member") ||
		strings.HasPrefix(m, "Trying to get property of non-object") ||
		strings.HasPrefix(m, "Access to undeclared static property") {
		r.ReportType = TypeUndefinedCall
	} else if strings.HasPrefix(m, "Call to private method") ||
		strings.HasPrefix(m, "Call to protected method") ||
		strings.HasPrefix(m, "Cannot access private property") ||
		strings.HasPrefix(m, "Cannot access protected property") {
		r.ReportType = TypeClassVisibility
	} else if strings.HasSuffix(m, " not found") {
		r.ReportType = TypeClassNotFound
	} else if strings.HasPrefix(m, "Argument") {
		r.ReportType = TypeInvalidArgument

		line := strings.Replace(r.Error.Message, "\\\\", "\\", -1)
		index1 := strings.Index(line, ", called in")
		r.Error.Message = line[0:index1]
	} else if r.Error.Type != "E_ERROR" &&
		r.Error.Type != "E_USER_ERROR" &&
		r.Error.Type != "1" {
		r.ReportType = TypeUnknown
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

// ReadCompressed reads the base64 encoded and zlib compressed report
func (r *CrashReport) ReadCompressed(report string) error {
	zlibBytes, err := base64.StdEncoding.DecodeString(extractBase64(report))
	if err != nil {
		return err
	}

	br := bytes.NewReader(zlibBytes)
	zr, err := zlib.NewReader(br)
	if err != nil {
		return err
	}
	defer zr.Close()

	err = json.NewDecoder(zr).Decode(&r.Data)
	return err
}

// clean is shoghi magic
func clean(v string) string {
	var re = regexp.MustCompile(`[^A-Za-z0-9_\-\.\,\;\:/\#\(\)\\ ]`)
	return re.ReplaceAllString(v, "")
}

// Encoded ...
func (r *CrashReport) Encoded() string {
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

	return fmt.Sprintf("%s\n%s\n%s", reportBegin, base64.StdEncoding.EncodeToString(zlibBuf.Bytes()), reportEnd)
}
