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
)

func Parse(data string) (*CrashReport, error) {
	report := trimHead(data)
	if report == "" {
		return nil, errors.New("report is empty")
	}

	var r CrashReport
	r.ReportType = TypeGeneric
	r.CausedByPlugin = false
	r.Valid = true

	if err := r.ReadCompressed(report); err != nil {
		return nil, fmt.Errorf("failed to read compressed data: %v\n", err)
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
		if plugin {
			r.CausedByPlugin = true
		}
	case string:
		r.CausingPlugin = clean(plugin)
	}

	r.Error.Type = r.Data.Error.Type
	r.Error.Message = r.Data.Error.Message
	r.Error.Line = r.Data.Error.Line
	r.Error.File = r.Data.Error.File
}

// ParseVersion ...
func (r *CrashReport) parseVersion() {
	if r.Data.General.Version == "" {
		r.Valid = false
		return
	}

	general := r.Data.General
	r.APIVersion = general.API
	r.Version = NewVersionString(general.Version, general.Build)
}

// ClassifyMessage ...
func (r *CrashReport) classifyMessage() {
	if r.Error.Message == "" {
		r.Valid = false
		return
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
		index1 := strings.Index(line, "called in")

		var index int
		if index = strings.Index(line, "src/"); index > 0 {
			r.ErrorMessage = fmt.Sprintf("%s %s", line[0:index1+10], line[:index])
		} else if index = strings.Index(line, "plugins/"); index > 0 {
			r.ErrorMessage = fmt.Sprintf("%s %s", line[0:index1+10], line[:index])
			r.CausedByPlugin = true
		}
	} else if r.Error.Type != "E_ERROR" &&
		r.Error.Type != "E_USER_ERROR" &&
		r.Error.Type != "1" {
		r.ReportType = TypeUnknown
	}
}

func trimHead(data string) string {
	x := strings.Trim(data, "\r\n\t` ")
	i := strings.Index(x, "===BEGIN CRASH DUMP===")
	if i == -1 {
		return data
	}
	x = x[i:]
	return x
}

// ReadCompressed reads the base64 encoded and zlib compressed report
func (r *CrashReport) ReadCompressed(report string) error {
	b64Enc := strings.Replace(report, "===BEGIN CRASH DUMP===", "", -1)
	b64Enc = strings.Replace(b64Enc, "===END CRASH DUMP===", "", -1)
	b64Dec, err := base64.StdEncoding.DecodeString(b64Enc)
	if err != nil {
		return err
	}

	br := bytes.NewReader(b64Dec)
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

	return fmt.Sprintf("===BEGIN CRASH DUMP===\n%s\n===END CRASH DUMP===", base64.StdEncoding.EncodeToString(zlibBuf.Bytes()))
}
