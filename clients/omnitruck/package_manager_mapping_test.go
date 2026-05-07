package omnitruck

import "testing"

func TestDerivePackageManager(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		want     string
	}{
		{name: "aix", platform: "aix", want: "bff"},
		{name: "amazon", platform: "amazon", want: "rpm"},
		{name: "centos", platform: "centos", want: "rpm"},
		{name: "darwin", platform: "darwin", want: "dmg"},
		{name: "debian", platform: "debian", want: "deb"},
		{name: "el", platform: "el", want: "rpm"},
		{name: "fedora", platform: "fedora", want: "rpm"},
		{name: "freebsd", platform: "freebsd", want: "sh"},
		{name: "ios_xr", platform: "ios_xr", want: "rpm"},
		{name: "linux", platform: "linux", want: "rpm"},
		{name: "linux-kernel2", platform: "linux-kernel2", want: "rpm"},
		{name: "linuxmint", platform: "linuxmint", want: "deb"},
		{name: "mac_os_x", platform: "mac_os_x", want: "dmg"},
		{name: "nexus", platform: "nexus", want: "rpm"},
		{name: "opensuseleap", platform: "opensuseleap", want: "rpm"},
		{name: "redhat", platform: "redhat", want: "rpm"},
		{name: "rocky", platform: "rocky", want: "rpm"},
		{name: "sles", platform: "sles", want: "rpm"},
		{name: "solaris2", platform: "solaris2", want: "p5p"},
		{name: "suse", platform: "suse", want: "rpm"},
		{name: "ubuntu", platform: "ubuntu", want: "deb"},
		{name: "windows", platform: "windows", want: "msi"},
		{name: "platform with spaces", platform: "  windows ", want: "msi"},
		{name: "platform with mixed case", platform: "Ubuntu", want: "deb"},
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
		{name: "amazon", platform: "amazon", want: "linux"},
		{name: "centos", platform: "centos", want: "linux"},
		{name: "debian", platform: "debian", want: "linux"},
		{name: "el", platform: "el", want: "linux"},
		{name: "fedora", platform: "fedora", want: "linux"},
		{name: "ios_xr", platform: "ios_xr", want: "linux"},
		{name: "linux", platform: "linux", want: "linux"},
		{name: "linux-kernel2", platform: "linux-kernel2", want: "linux"},
		{name: "linuxmint", platform: "linuxmint", want: "linux"},
		{name: "nexus", platform: "nexus", want: "linux"},
		{name: "opensuseleap", platform: "opensuseleap", want: "linux"},
		{name: "redhat", platform: "redhat", want: "linux"},
		{name: "rocky", platform: "rocky", want: "linux"},
		{name: "sles", platform: "sles", want: "linux"},
		{name: "suse", platform: "suse", want: "linux"},
		{name: "ubuntu", platform: "ubuntu", want: "linux"},
		{name: "darwin", platform: "darwin", want: "darwin"},
		{name: "mac_os_x", platform: "mac_os_x", want: "darwin"},
		{name: "windows", platform: "windows", want: "windows"},
		{name: "aix", platform: "aix", want: "aix"},
		{name: "freebsd", platform: "freebsd", want: "freebsd"},
		{name: "solaris2", platform: "solaris2", want: "solaris2"},
		{name: "platform with spaces", platform: "  Ubuntu ", want: "linux"},
		{name: "platform with mixed case", platform: "Ubuntu", want: "linux"},
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
