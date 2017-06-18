package crashreport

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
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

// CrashReport ...
type CrashReport struct {
	ReportType   string
	ErrorMessage string
	Valid        bool

	CausedByPlugin bool
	CausingPlugin  string

	Data       *ReportData
	ReportDate time.Time

	Version    *VersionString
	APIVersion string

	Error struct {
		Type    string
		Message string
		Line    int
		File    string
	}
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
	Error struct {
		Type    string
		Message string
		Line    int
		File    string
	}
	Code             map[string]string
	Plugins          interface{} `json:"plugins,omitempty"`
	PocketmineYML    string      `json:"pocketmine.yml"`
	ServerProperties string      `json:"server.properties"`
	Trace            []string
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
}

func FromString(data string) (*CrashReport, error) {
	return parse(data)
}
func FromFile(f *multipart.FileHeader) (*CrashReport, error) {

	return nil, nil
}

func parse(data string) (*CrashReport, error) {
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

	// TODO: clean this up!
	r.Error.Type = r.Data.Error.Type
	r.Error.Message = r.Data.Error.Message
	r.Error.Line = r.Data.Error.Line
	r.Error.File = r.Data.Error.File
	// log.Println("DONE: ParseError!")
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
	// log.Println("DONE: ParseVersion!")
}

// ClassifyMessage ...
func (r *CrashReport) classifyMessage() {
	// log.Println("TODO: ClassifyMessage")
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
		strings.HasPrefix(m, "Trying to get property of non-onject") ||
		strings.HasPrefix(m, "Acces to undeclared static property") {
		r.ReportType = TypeUndefinedCall
	} else if strings.HasPrefix(m, "Call to private method") ||
		strings.HasPrefix(m, "Call to protected method") ||
		strings.HasPrefix(m, "Cannot access private property") ||
		strings.HasPrefix(m, "Cannot access protected propery") {
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
		return ""
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
	//var x map[string]interface{}
	//err = json.NewDecoder(zr).Decode(&x)
	//log.Printf("%#v\n", x)
	//buf, _ := json.MarshalIndent(x, "", "    ")
	//fmt.Printf("%s\n", buf)

	//err = json.Unmarshal(buf, &r.Data)
	err = json.NewDecoder(zr).Decode(&r.Data)
	if err != nil {
		return err
	}

	return nil
}

// clean is shoghi magic
func clean(v string) string {
	var re = regexp.MustCompile("[^A-Za-z0-9_\\-\\.\\,\\;\\:/\\#\\(\\)\\\\ ]")
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
	//log.Println(hex.Dump(jsonBuf.Bytes()))
	var zlibBuf bytes.Buffer
	zw := zlib.NewWriter(&zlibBuf)
	defer zw.Close()
	_, err = zw.Write(jsonBuf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	zw.Flush()
	//log.Println(hex.Dump(zlibBuf.Bytes()))

	return fmt.Sprintf("===BEGIN CRASH DUMP===\n%s\n===END CRASH DUMP===", base64.StdEncoding.EncodeToString(zlibBuf.Bytes()))
}
