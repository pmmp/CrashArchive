package crashreport

import (
	"fmt"
	"regexp"
	"strconv"
)

// VersionString ...
type VersionString struct {
	Minor       int
	Major       int
	Generation  int
	Development bool
	Build       int
}

var reVersion = regexp.MustCompile(`([0-9]*)\.([0-9]*)\.{0,1}([0-9]*)(dev|)(-[0-9]{1,}|)`)

// NewVersionString ...
func NewVersionString(version string, build int) *VersionString {
	v := &VersionString{}
	if build > 0 {
		version = fmt.Sprintf("%s-%d", version, build)
	}
	if version, err := strconv.Atoi(version); err == nil {
		v.Minor = version & 0x1f
		v.Major = (version >> 5) & 0x0f
		v.Generation = (version >> 9) & 0x0f
		return v
	}

	matches := reVersion.FindStringSubmatch(version)
	if len(matches) > 0 {
		v.Generation, _ = strconv.Atoi(matches[1])
		v.Major, _ = strconv.Atoi(matches[2])
		v.Minor, _ = strconv.Atoi(matches[3])
		if matches[4] == "dev" {
			v.Development = true
		}
		if matches[5] != "" {
			v.Build, _ = strconv.Atoi(matches[5][1:])
		}
	}
	return v
}

// Get ...
func (v *VersionString) Get(b bool) string {
	var dev string
	var build string
	if v.Development {
		dev = "dev"
	}
	if (v.Build > 0) && b {
		build = fmt.Sprintf("-%d", v.Build)
	}
	return fmt.Sprintf("%s%s%s", v.release(), dev, build)
}

// release ...
func (v *VersionString) release() string {
	var minor string
	if v.Minor > 0 {
		minor = fmt.Sprintf(".%d", v.Minor)
	}
	return fmt.Sprintf("%d.%d%s", v.Generation, v.Major, minor)
}

// build ...
func (v *VersionString) build(b bool) string {
	var build string
	if (v.Build > 0) && b {
		build = fmt.Sprintf("-%d", v.Build)
	}
	return fmt.Sprintf("%d.%d%s", v.Generation, v.Major, build)
}
