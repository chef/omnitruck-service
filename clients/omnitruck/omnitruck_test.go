package omnitruck

import "testing"

func TestPackageList_UpdatePackages(t *testing.T) {
	tests := []struct {
		name        string
		pl          PackageList
		updater     PackageListUpdater
		wantVersion string
		wantUrl     string
	}{
		{
			name: "basic",
			pl: PackageList{
				"a": PlatformVersionList{
					"1": ArchList{
						"el": PackageMetadata{
							Version: "1.0",
							Url:     "https://oldurl.com",
						},
					},
				},
				"b": PlatformVersionList{
					"1": ArchList{
						"el": PackageMetadata{
							Version: "1.0",
							Url:     "https://old2url.com",
						},
					},
				},
			},
			updater: func(p string, pv string, arch string, m PackageMetadata) PackageMetadata {
				m.Url = "https://newurl.com"

				return m
			},
			wantVersion: "1.0",
			wantUrl:     "https://newurl.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.pl.UpdatePackages(tt.updater)
			for _, versions := range tt.pl {
				for _, arches := range versions {
					for _, metadata := range arches {
						if got := metadata.Version; got != tt.wantVersion {
							t.Errorf("Metadata version not updated, got %v, wanted %v", got, tt.wantVersion)
						}

						if got := metadata.Url; got != tt.wantUrl {
							t.Errorf("Metadata url not updated, got %v, wanted %v", got, tt.wantUrl)
						}
					}
				}
			}
		})
	}
}
