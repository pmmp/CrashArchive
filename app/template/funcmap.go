package template

import (
	"github.com/pmmp/CrashArchive/app/crashreport"

	"fmt"
	"html/template"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"net/url"
	"log"
	"time"
)

var funcMap = template.FuncMap{
	"shorten":   shorten,
	"split":     split,
	"sortcode":  sortcode,
	"shorthash": shorthash,
	"pagenum":   pagenum,
	"add":       add,
	"pluginInvolvementToString": pluginInvolvementToString,
	"isDirectPluginCrash": isDirectPluginCrash,
	"isIndirectPluginCrash": isIndirectPluginCrash,
	"renderDuration": renderDuration,
	"renderJitMode": renderJitMode,
	"date": date,
}

func shorthash(s string) string {
	return s[:8]
}

func shorten(s string) string {
	if len(s) > 50 {
		return s[:50] + "..."
	}
	return s
}

var splitRegex = regexp.MustCompile(`(.*)=(.*)`)

func split(x string) template.HTML {
	r := make([]string, 0)
	for _, v := range strings.Split(x, "\n") {
		m := splitRegex.FindStringSubmatch(v)
		if len(m) == 3 {
			r = append(r, fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>", m[1], m[2]))
		}
	}
	return template.HTML(strings.Join(r, ""))
}

type SortedCode struct {
	StartLine int
	Lines     []string
}

func sortcode(a map[string]string) SortedCode {
	startLine := math.MaxUint32

	s := make([]int, 0)
	for k, _ := range a {
		c, _ := strconv.Atoi(k)
		s = append(s, c)
		if c < startLine {
			startLine = c
		}
	}

	r := make([]string, 0)
	sort.Ints(s)
	for _, v := range s {
		r = append(r, a[strconv.Itoa(v)])
	}
	return SortedCode{StartLine: startLine, Lines: r}
}

func pagenum(base string, page int) string {
	parsed, err := url.Parse(base)
	if err != nil {
		log.Println(err)
		return base
	}

	params := parsed.Query()
	params.Set("page", strconv.Itoa(page))
	parsed.RawQuery = params.Encode()
	return parsed.String()
}

func add(num1 int, num2 int) int {
	return num1 + num2
}

func pluginInvolvementToString (ctype string) string {
	f := crashreport.PluginInvolvementStrings[ctype]
	if f == "" {
		return "Unknown"
	}
	return f
}

func isDirectPluginCrash (ctype string) bool {
	return ctype == crashreport.PIDirect
}

func isIndirectPluginCrash (ctype string) bool {
	return ctype == crashreport.PIIndirect
}

func renderDuration(duration float64) string {
	durationInt := int64(duration)
	days := durationInt / (3600 * 24)
	hours := (durationInt % (3600 * 24)) / 3600
	minutes := (durationInt % 3600) / 60
	seconds := durationInt % 60
	return fmt.Sprintf("%d days %d hours %d minutes %d seconds", days, hours, minutes, seconds)
}

func renderJitMode(mode *int) string {
	if mode != nil {
		if *mode == 0 {
			return "Disabled"
		}
		return fmt.Sprintf("Enabled (%d)", *mode)
	} else {
		return "Not available"
	}
}

func date(timestamp int64) string {
	time := time.Unix(timestamp, 0)
	//Out with Golang and its stupid date formatting codes
	return fmt.Sprintf("%d-%02d-%02d", time.Year(), time.Month(), time.Day())
}
