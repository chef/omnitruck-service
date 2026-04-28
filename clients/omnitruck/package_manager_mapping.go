package omnitruck

import "strings"

var platformToPackageManager = map[string]string{
	"aix":           "bff",
	"amazon":        "rpm",
	"darwin":        "dmg",
	"debian":        "deb",
	"el":            "rpm",
	"freebsd":       "sh",
	"ios_xr":        "rpm",
	"linux":         "rpm",
	"linux-kernel2": "rpm",
	"mac_os_x":      "dmg",
	"nexus":         "rpm",
	"rocky":         "rpm",
	"sles":          "rpm",
	"solaris2":      "p5p",
	"suse":          "rpm",
	"ubuntu":        "deb",
	"windows":       "msi",
}

func DerivePackageManager(platform string) string {
	normalized := strings.ToLower(strings.TrimSpace(platform))
	return platformToPackageManager[normalized]
}

func IsUniversalPackageManager(packageManager string) bool {
	switch strings.ToLower(strings.TrimSpace(packageManager)) {
	case "tar", "zip":
		return true
	default:
		return false
	}
}
