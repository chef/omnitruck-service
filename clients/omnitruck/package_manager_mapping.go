package omnitruck

import "strings"

var platformToPackageManager = map[string]string{
	"aix":           "bff",
	"amazon":        "rpm",
	"centos":        "rpm",
	"darwin":        "dmg",
	"debian":        "deb",
	"el":            "rpm",
	"fedora":        "rpm",
	"freebsd":       "sh",
	"ios_xr":        "rpm",
	"linux":         "rpm",
	"linux-kernel2": "rpm",
	"linuxmint":     "deb",
	"mac_os_x":      "dmg",
	"nexus":         "rpm",
	"opensuseleap":  "rpm",
	"redhat":        "rpm",
	"rocky":         "rpm",
	"sles":          "rpm",
	"solaris2":      "p5p",
	"suse":          "rpm",
	"ubuntu":        "deb",
	"windows":       "msi",
}

// platformToDbPlatform maps user-provided platforms to database platform keys
var platformToDbPlatform = map[string]string{
	// Linux variants all map to "linux" in database
	"amazon":        "linux",
	"centos":        "linux",
	"debian":        "linux",
	"el":            "linux",
	"fedora":        "linux",
	"ios_xr":        "linux",
	"linux":         "linux",
	"linux-kernel2": "linux",
	"linuxmint":     "linux",
	"nexus":         "linux",
	"opensuseleap":  "linux",
	"redhat":        "linux",
	"rocky":         "linux",
	"sles":          "linux",
	"suse":          "linux",
	"ubuntu":        "linux",

	// Darwin/macOS variants
	"darwin":   "darwin",
	"mac_os_x": "darwin",

	// Windows variants
	"windows": "windows",

	// Others remain as-is for now (may not exist in DB)
	"aix":      "aix",
	"freebsd":  "freebsd",
	"solaris2": "solaris2",
}

func NormalizePlatformForDatabase(platform string) string {
	normalized := strings.ToLower(strings.TrimSpace(platform))
	if dbPlatform, exists := platformToDbPlatform[normalized]; exists {
		return dbPlatform
	}
	// Return the normalized platform if no mapping is found
	return normalized
}

func DerivePackageManager(platform string) string {
	normalized := strings.ToLower(strings.TrimSpace(platform))
	return platformToPackageManager[normalized]
}
