package omnitruck

import "testing"

func TestDerivePackageManager(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		want     string
	}{
		// Enterprise Linux
		{name: "almalinux", platform: "almalinux", want: "rpm"},
		{name: "alibabalinux", platform: "alibabalinux", want: "rpm"},
		{name: "alinux", platform: "alinux", want: "rpm"},
		{name: "bigip", platform: "bigip", want: "rpm"},
		{name: "centos", platform: "centos", want: "rpm"},
		{name: "clearos", platform: "clearos", want: "rpm"},
		{name: "cloudlinux", platform: "cloudlinux", want: "rpm"},
		{name: "el", platform: "el", want: "rpm"},
		{name: "enterpriseenterprise", platform: "enterpriseenterprise", want: "rpm"},
		{name: "ibm_powerkvm", platform: "ibm_powerkvm", want: "rpm"},
		{name: "nexus_centos", platform: "nexus_centos", want: "rpm"},
		{name: "ol", platform: "ol", want: "rpm"},
		{name: "oracle", platform: "oracle", want: "rpm"},
		{name: "parallels", platform: "parallels", want: "rpm"},
		{name: "redhat", platform: "redhat", want: "rpm"},
		{name: "rhel", platform: "rhel", want: "rpm"},
		{name: "rocky", platform: "rocky", want: "rpm"},
		{name: "sangoma", platform: "sangoma", want: "rpm"},
		{name: "scientific", platform: "scientific", want: "rpm"},
		{name: "virtuozzo", platform: "virtuozzo", want: "rpm"},
		{name: "xcp-ng", platform: "xcp-ng", want: "rpm"},
		{name: "xenenterprise", platform: "xenenterprise", want: "rpm"},
		{name: "xenserver", platform: "xenserver", want: "rpm"},

		// Fedora and Fedora-derived
		{name: "arista_eos", platform: "arista_eos", want: "rpm"},
		{name: "fedora", platform: "fedora", want: "rpm"},

		// Amazon Linux
		{name: "amazon", platform: "amazon", want: "rpm"},
		{name: "amzn", platform: "amzn", want: "rpm"},

		// SUSE
		{name: "opensuse", platform: "opensuse", want: "rpm"},
		{name: "opensuse-leap", platform: "opensuse-leap", want: "rpm"},
		{name: "opensuseleap", platform: "opensuseleap", want: "rpm"},
		{name: "sled", platform: "sled", want: "rpm"},
		{name: "sles", platform: "sles", want: "rpm"},
		{name: "sles_sap", platform: "sles_sap", want: "rpm"},
		{name: "suse", platform: "suse", want: "rpm"},

		// Debian and Debian-derived
		{name: "cumulus", platform: "cumulus", want: "deb"},
		{name: "cumulus-linux", platform: "cumulus-linux", want: "deb"},
		{name: "cumulus_linux", platform: "cumulus_linux", want: "deb"},
		{name: "cumulus_networks", platform: "cumulus_networks", want: "deb"},
		{name: "debian", platform: "debian", want: "deb"},
		{name: "kali", platform: "kali", want: "deb"},
		{name: "linuxmint", platform: "linuxmint", want: "deb"},
		{name: "pop", platform: "pop", want: "deb"},
		{name: "raspbian", platform: "raspbian", want: "deb"},
		{name: "ubuntu", platform: "ubuntu", want: "deb"},

		// Wind River Linux
		{name: "ios_xr", platform: "ios_xr", want: "rpm"},
		{name: "nexus", platform: "nexus", want: "rpm"},

		// Generic Linux
		{name: "linux", platform: "linux", want: "rpm"},
		{name: "linux-kernel2", platform: "linux-kernel2", want: "rpm"},

		// macOS
		{name: "darwin", platform: "darwin", want: "dmg"},
		{name: "mac_os_x", platform: "mac_os_x", want: "dmg"},
		{name: "macos", platform: "macos", want: "dmg"},

		// Windows
		{name: "windows", platform: "windows", want: "msi"},

		// Other UNIX
		{name: "aix", platform: "aix", want: "bff"},
		{name: "freebsd", platform: "freebsd", want: "sh"},
		{name: "solaris", platform: "solaris", want: "p5p"},
		{name: "solaris2", platform: "solaris2", want: "p5p"},

		// Input normalization
		{name: "platform with spaces", platform: "  windows ", want: "msi"},
		{name: "platform with mixed case", platform: "Ubuntu", want: "deb"},
		{name: "mixed case alias", platform: "AlmaLinux", want: "rpm"},
		{name: "unknown", platform: "unknown", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DerivePackageManager(tt.platform); got != tt.want {
				t.Errorf("DerivePackageManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizePlatformForDatabase(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		want     string
	}{
		// Enterprise Linux
		{name: "almalinux", platform: "almalinux", want: "linux"},
		{name: "alibabalinux", platform: "alibabalinux", want: "linux"},
		{name: "alinux", platform: "alinux", want: "linux"},
		{name: "bigip", platform: "bigip", want: "linux"},
		{name: "centos", platform: "centos", want: "linux"},
		{name: "clearos", platform: "clearos", want: "linux"},
		{name: "cloudlinux", platform: "cloudlinux", want: "linux"},
		{name: "el", platform: "el", want: "linux"},
		{name: "enterpriseenterprise", platform: "enterpriseenterprise", want: "linux"},
		{name: "ibm_powerkvm", platform: "ibm_powerkvm", want: "linux"},
		{name: "nexus_centos", platform: "nexus_centos", want: "linux"},
		{name: "ol", platform: "ol", want: "linux"},
		{name: "oracle", platform: "oracle", want: "linux"},
		{name: "parallels", platform: "parallels", want: "linux"},
		{name: "redhat", platform: "redhat", want: "linux"},
		{name: "rhel", platform: "rhel", want: "linux"},
		{name: "rocky", platform: "rocky", want: "linux"},
		{name: "sangoma", platform: "sangoma", want: "linux"},
		{name: "scientific", platform: "scientific", want: "linux"},
		{name: "virtuozzo", platform: "virtuozzo", want: "linux"},
		{name: "xcp-ng", platform: "xcp-ng", want: "linux"},
		{name: "xenenterprise", platform: "xenenterprise", want: "linux"},
		{name: "xenserver", platform: "xenserver", want: "linux"},

		// Fedora and Fedora-derived
		{name: "arista_eos", platform: "arista_eos", want: "linux"},
		{name: "fedora", platform: "fedora", want: "linux"},

		// Amazon Linux
		{name: "amazon", platform: "amazon", want: "linux"},
		{name: "amzn", platform: "amzn", want: "linux"},

		// SUSE
		{name: "opensuse", platform: "opensuse", want: "linux"},
		{name: "opensuse-leap", platform: "opensuse-leap", want: "linux"},
		{name: "opensuseleap", platform: "opensuseleap", want: "linux"},
		{name: "sled", platform: "sled", want: "linux"},
		{name: "sles", platform: "sles", want: "linux"},
		{name: "sles_sap", platform: "sles_sap", want: "linux"},
		{name: "suse", platform: "suse", want: "linux"},

		// Debian and Debian-derived
		{name: "cumulus", platform: "cumulus", want: "linux"},
		{name: "cumulus-linux", platform: "cumulus-linux", want: "linux"},
		{name: "cumulus_linux", platform: "cumulus_linux", want: "linux"},
		{name: "cumulus_networks", platform: "cumulus_networks", want: "linux"},
		{name: "debian", platform: "debian", want: "linux"},
		{name: "kali", platform: "kali", want: "linux"},
		{name: "linuxmint", platform: "linuxmint", want: "linux"},
		{name: "pop", platform: "pop", want: "linux"},
		{name: "raspbian", platform: "raspbian", want: "linux"},
		{name: "ubuntu", platform: "ubuntu", want: "linux"},

		// Wind River Linux
		{name: "ios_xr", platform: "ios_xr", want: "linux"},
		{name: "nexus", platform: "nexus", want: "linux"},

		// Generic Linux
		{name: "linux", platform: "linux", want: "linux"},
		{name: "linux-kernel2", platform: "linux-kernel2", want: "linux"},

		// macOS
		{name: "darwin", platform: "darwin", want: "darwin"},
		{name: "mac_os_x", platform: "mac_os_x", want: "darwin"},
		{name: "macos", platform: "macos", want: "darwin"},

		// Windows
		{name: "windows", platform: "windows", want: "windows"},

		// Other UNIX
		{name: "aix", platform: "aix", want: "aix"},
		{name: "freebsd", platform: "freebsd", want: "freebsd"},
		{name: "solaris", platform: "solaris", want: "solaris2"},
		{name: "solaris2", platform: "solaris2", want: "solaris2"},

		// Input normalization
		{name: "platform with spaces", platform: "  Ubuntu ", want: "linux"},
		{name: "platform with mixed case", platform: "Ubuntu", want: "linux"},
		{name: "mixed case alias", platform: "AlmaLinux", want: "linux"},
		{name: "unknown platform returned as-is", platform: "unknown", want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizePlatformForDatabase(tt.platform); got != tt.want {
				t.Errorf("NormalizePlatformForDatabase() = %v, want %v", got, tt.want)
			}
		})
	}
}

// A platform that derives a package manager but has no database mapping fails
// later, and less obviously, with "Product information not found" -- the
// un-normalized platform name reaches the database lookup. Keep the two maps in
// sync so a new alias is either fully supported or not accepted at all.
func TestEveryPackageManagerPlatformHasDbMapping(t *testing.T) {
	for platform := range platformToPackageManager {
		if _, ok := platformToDbPlatform[platform]; !ok {
			t.Errorf("platform %q is in platformToPackageManager but missing from platformToDbPlatform", platform)
		}
	}
}

func TestEveryDbPlatformDerivesAPackageManager(t *testing.T) {
	for platform := range platformToDbPlatform {
		if DerivePackageManager(platform) == "" {
			t.Errorf("platform %q is in platformToDbPlatform but missing from platformToPackageManager", platform)
		}
	}
}

// Aliases for the same underlying platform must agree on both the package format
// and the database key, otherwise the artifact a caller receives depends on which
// spelling their platform detection happened to report.
func TestPlatformAliasesAgree(t *testing.T) {
	aliasGroups := map[string][]string{
		"oracle linux":  {"oracle", "ol", "enterpriseenterprise"},
		"rhel":          {"redhat", "rhel"},
		"xenserver":     {"xenserver", "xenenterprise"},
		"alibaba linux": {"alibabalinux", "alinux"},
		"amazon linux":  {"amazon", "amzn"},
		"opensuse leap": {"opensuseleap", "opensuse-leap"},
		"cumulus linux": {"cumulus", "cumulus-linux", "cumulus_linux", "cumulus_networks"},
		"macos":         {"mac_os_x", "darwin", "macos"},
		"solaris":       {"solaris", "solaris2"},
		"sles":          {"sles", "sles_sap"},
	}

	for name, aliases := range aliasGroups {
		t.Run(name, func(t *testing.T) {
			wantPM := DerivePackageManager(aliases[0])
			wantDB := NormalizePlatformForDatabase(aliases[0])
			if wantPM == "" {
				t.Fatalf("alias group %q: reference platform %q derives no package manager", name, aliases[0])
			}
			for _, alias := range aliases[1:] {
				if got := DerivePackageManager(alias); got != wantPM {
					t.Errorf("DerivePackageManager(%q) = %q, want %q (same as %q)", alias, got, wantPM, aliases[0])
				}
				if got := NormalizePlatformForDatabase(alias); got != wantDB {
					t.Errorf("NormalizePlatformForDatabase(%q) = %q, want %q (same as %q)", alias, got, wantDB, aliases[0])
				}
			}
		})
	}
}
