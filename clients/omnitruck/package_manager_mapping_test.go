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
		{name: "darwin", platform: "darwin", want: "dmg"},
		{name: "debian", platform: "debian", want: "deb"},
		{name: "el", platform: "el", want: "rpm"},
		{name: "freebsd", platform: "freebsd", want: "sh"},
		{name: "ios_xr", platform: "ios_xr", want: "rpm"},
		{name: "linux", platform: "linux", want: "rpm"},
		{name: "linux-kernel2", platform: "linux-kernel2", want: "rpm"},
		{name: "mac_os_x", platform: "mac_os_x", want: "dmg"},
		{name: "nexus", platform: "nexus", want: "rpm"},
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

func TestIsUniversalPackageManager(t *testing.T) {
	tests := []struct {
		name           string
		packageManager string
		want           bool
	}{
		{name: "tar", packageManager: "tar", want: true},
		{name: "zip", packageManager: "zip", want: true},
		{name: "tar mixed case", packageManager: "Tar", want: true},
		{name: "rpm", packageManager: "rpm", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUniversalPackageManager(tt.packageManager); got != tt.want {
				t.Errorf("IsUniversalPackageManager() = %v, want %v", got, tt.want)
			}
		})
	}
}
