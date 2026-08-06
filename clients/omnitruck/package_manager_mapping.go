package omnitruck

import "strings"

// platformToPackageManager maps a caller-supplied platform to the package format
// we publish for it.
//
// Entries are grouped by platform family. Within a family every alias resolves to
// the same package format, because the family is what determines the artifact.
//
// Aliases exist because callers legitimately report different names for the same
// platform:
//
//   - Ohai's `node['platform']` (what Chef Infra Client reports).
//   - The raw `ID=` field from /etc/os-release. Chef's own install.sh falls back to
//     `platform=$ID` for any distro without an older release file, and Ohai
//     normalizes several of these (e.g. `ol` -> `oracle`, `opensuse-leap` ->
//     `opensuseleap`, `rhel` -> `redhat`).
//   - LSB `DISTRIB_ID`, which install.sh uses when /etc/lsb-release is present.
//
// install.sh is explicit that this normalization belongs here rather than in the
// client: "remapping platform and mangling platform version numbers is now the
// complete responsibility of the server-side endpoints".
var platformToPackageManager = map[string]string{
	// Enterprise Linux (RHEL and rebuilds/derivatives)
	"almalinux":            "rpm",
	"alibabalinux":         "rpm",
	"alinux":               "rpm", // os-release ID for Alibaba Cloud Linux
	"bigip":                "rpm",
	"centos":               "rpm",
	"clearos":              "rpm",
	"cloudlinux":           "rpm",
	"el":                   "rpm",
	"enterpriseenterprise": "rpm", // Oracle Linux LSB distributor ID
	"ibm_powerkvm":         "rpm",
	"nexus_centos":         "rpm",
	"ol":                   "rpm", // os-release ID for Oracle Linux
	"oracle":               "rpm",
	"parallels":            "rpm",
	"redhat":               "rpm",
	"rhel":                 "rpm", // os-release ID for RHEL
	"rocky":                "rpm",
	"sangoma":              "rpm",
	"scientific":           "rpm",
	"virtuozzo":            "rpm",
	"xcp-ng":               "rpm",
	"xenenterprise":        "rpm", // os-release ID for XenServer 7.5+
	"xenserver":            "rpm",

	// Fedora and Fedora-derived
	"arista_eos": "rpm",
	"fedora":     "rpm",

	// Amazon Linux
	"amazon": "rpm",
	"amzn":   "rpm", // os-release ID for Amazon Linux

	// SUSE
	"opensuse":      "rpm",
	"opensuse-leap": "rpm", // os-release ID for openSUSE Leap 15+
	"opensuseleap":  "rpm",
	"sled":          "rpm",
	"sles":          "rpm",
	"sles_sap":      "rpm", // os-release ID for SLES for SAP
	"suse":          "rpm",

	// Debian and Debian-derived
	"cumulus":          "deb",
	"cumulus-linux":    "deb", // os-release ID for Cumulus Linux
	"cumulus_linux":    "deb", // LSB DISTRIB_ID for Cumulus Linux
	"cumulus_networks": "deb", // LSB DISTRIB_ID for Cumulus Networks
	"debian":           "deb",
	"kali":             "deb",
	"linuxmint":        "deb",
	"pop":              "deb",
	"raspbian":         "deb",
	"ubuntu":           "deb",

	// Wind River Linux (Cisco network operating systems)
	"ios_xr": "rpm",
	"nexus":  "rpm",

	// Generic Linux
	"linux":         "rpm",
	"linux-kernel2": "rpm",

	// macOS
	"darwin":   "dmg",
	"mac_os_x": "dmg",
	"macos":    "dmg",

	// Windows
	"windows": "msi",

	// Other UNIX
	"aix":      "bff",
	"freebsd":  "sh",
	"solaris":  "p5p",
	"solaris2": "p5p",
}

// platformToDbPlatform maps user-provided platforms to database platform keys.
//
// Every platform in platformToPackageManager must also appear here, otherwise an
// otherwise-valid alias resolves a package manager but then fails the lookup with
// "Product information not found" because the un-normalized name reaches the
// database. TestEveryPackageManagerPlatformHasDbMapping enforces this.
var platformToDbPlatform = map[string]string{
	// Linux variants all map to "linux" in database
	"almalinux":            "linux",
	"alibabalinux":         "linux",
	"alinux":               "linux",
	"amazon":               "linux",
	"amzn":                 "linux",
	"arista_eos":           "linux",
	"bigip":                "linux",
	"centos":               "linux",
	"clearos":              "linux",
	"cloudlinux":           "linux",
	"cumulus":              "linux",
	"cumulus-linux":        "linux",
	"cumulus_linux":        "linux",
	"cumulus_networks":     "linux",
	"debian":               "linux",
	"el":                   "linux",
	"enterpriseenterprise": "linux",
	"fedora":               "linux",
	"ibm_powerkvm":         "linux",
	"ios_xr":               "linux",
	"kali":                 "linux",
	"linux":                "linux",
	"linux-kernel2":        "linux",
	"linuxmint":            "linux",
	"nexus":                "linux",
	"nexus_centos":         "linux",
	"ol":                   "linux",
	"opensuse":             "linux",
	"opensuse-leap":        "linux",
	"opensuseleap":         "linux",
	"oracle":               "linux",
	"parallels":            "linux",
	"pop":                  "linux",
	"raspbian":             "linux",
	"redhat":               "linux",
	"rhel":                 "linux",
	"rocky":                "linux",
	"sangoma":              "linux",
	"scientific":           "linux",
	"sled":                 "linux",
	"sles":                 "linux",
	"sles_sap":             "linux",
	"suse":                 "linux",
	"ubuntu":               "linux",
	"virtuozzo":            "linux",
	"xcp-ng":               "linux",
	"xenenterprise":        "linux",
	"xenserver":            "linux",

	// Darwin/macOS variants
	"darwin":   "darwin",
	"mac_os_x": "darwin",
	"macos":    "darwin",

	// Windows variants
	"windows": "windows",

	// Others remain as-is for now (may not exist in DB)
	"aix":      "aix",
	"freebsd":  "freebsd",
	"solaris":  "solaris2",
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
