package crashreport

import (
	"fmt"
	"regexp"
	"strconv"
)

// VersionString ...
type VersionString struct {
	Major       int
	Minor       int
	Patch       int
	Suffix      string
	Development bool
	Build       int
}

var reVersion = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)(?:-(.*))?$`)

// NewVersionString ...
func NewVersionString(version string, build int, development bool) (*VersionString, error) {
	v := &VersionString{}

	v.Build = build
	v.Development = development

	matches := reVersion.FindStringSubmatch(version)
	if len(matches) >= 4 {
		v.Major, _ = strconv.Atoi(matches[1])
		v.Minor, _ = strconv.Atoi(matches[2])
		v.Patch, _ = strconv.Atoi(matches[3])
		v.Suffix = matches[4]
		return v, nil
	}

	return nil, fmt.Errorf("failed to parse version string %s", version)
}

// Get ...
func (v *VersionString) Get(withBuild bool) string {
	var suffix string
	if v.Development {
		suffix = "+dev"
		if (v.Build > 0) && withBuild {
			suffix = fmt.Sprintf("%s.%d", suffix, v.Build)
		}
	}

	return fmt.Sprintf("%s%s", v.baseVersion(), suffix)
}

// release ...
func (v *VersionString) baseVersion() string {
	var suffix string
	if v.Suffix != "" {
		suffix = fmt.Sprintf("-%s", v.Suffix)
	}
	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Patch, suffix)
}

