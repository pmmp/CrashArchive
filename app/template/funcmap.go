package template

import (
	"fmt"
	"html/template"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var funcMap = template.FuncMap{
	"shorten":   shorten,
	"split":     split,
	"sortcode":  sortcode,
	"shorthash": shorthash,
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

func sortcode(a map[string]string) template.HTML {
	s := make([]int, 0)
	for k, _ := range a {
		c, _ := strconv.Atoi(k)
		s = append(s, c)

	}

	r := make([]string, 0)
	sort.Ints(s)
	for _, v := range s {
		r = append(r, fmt.Sprintf("%2d %s", v, template.HTMLEscapeString(a[strconv.Itoa(v)])))
	}
	return template.HTML(strings.Join(r, "<br/>"))
}
