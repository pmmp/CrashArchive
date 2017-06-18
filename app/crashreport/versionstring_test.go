package crashreport

import "testing"

func TestVersionString(t *testing.T) {
	testVersions := [...]string{"1.6.2dev-178", "1.6", "1.6dev", "1.6.2dev", "1.6.2-1800"}

	for _, versionText := range testVersions {
		version := NewVersionString(versionText, 0);
		if(version.Get(true) != versionText){
			t.Fatalf("Bad version string, expected %s, got %s", versionText, version.Get(true));
		}
	}
}
