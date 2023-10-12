package omnitruck

import (
	"reflect"
	"testing"

	"github.com/chef/omnitruck-service/clients"
)

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

func TestValidateRequest(t *testing.T) {
	type args struct {
		p     *RequestParams
		flags RequestParamsFlags
	}
	tests := []struct {
		name string
		args args
		want *clients.Request
	}{
		{
			name: "platform version params missing",
			args: args{
				p: &RequestParams{
					Channel:         "",
					Product:         "",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
				flags: RequestParamsFlags{
					PlatformVersion: true,
				},
			},
			want: &clients.Request{
				Code:    400,
				Message: "Platform Version (pv) params cannot be empty",
				Ok:      false,
			},
		},
		{
			name: "channel params missing",
			args: args{
				p: &RequestParams{
					Channel:         "",
					Product:         "",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
				flags: RequestParamsFlags{
					Channel: true,
				},
			},
			want: &clients.Request{
				Code:    400,
				Message: "Channel can only be stable or current",
				Ok:      false,
			},
		},
		{
			name: "channel params incorrect",
			args: args{
				p: &RequestParams{
					Channel:         "st",
					Product:         "",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
				flags: RequestParamsFlags{
					Channel: true,
				},
			},
			want: &clients.Request{
				Code:    400,
				Message: "Channel can only be stable or current",
				Ok:      false,
			},
		},
		{
			name: "platform  params missing",
			args: args{
				p: &RequestParams{
					Channel:         "",
					Product:         "",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
				flags: RequestParamsFlags{
					Platform: true,
				},
			},
			want: &clients.Request{
				Code:    400,
				Message: "Platfrom (p) params cannot be empty",
				Ok:      false,
			},
		},
		{
			name: "bom  params missing",
			args: args{
				p: &RequestParams{
					Channel:         "",
					Product:         "",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
				flags: RequestParamsFlags{
					BOM: true,
				},
			},
			want: &clients.Request{
				Code:    400,
				Message: "BOM (bom) params cannot be empty",
				Ok:      false,
			},
		},
		{
			name: "architecture  params missing",
			args: args{
				p: &RequestParams{
					Channel:         "",
					Product:         "",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
				flags: RequestParamsFlags{
					Architecture: true,
				},
			},
			want: &clients.Request{
				Code:    400,
				Message: "Architecture (m) params cannot be empty",
				Ok:      false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateRequest(tt.args.p, tt.args.flags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
