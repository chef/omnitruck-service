package omnitruck

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/chef/omnitruck-service/clients"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

func TestRequestParams_UrlParams(t *testing.T) {
	rp := &RequestParams{
		Version:         "1.2.3",
		Platform:        "ubuntu",
		PlatformVersion: "22.04",
		Architecture:    "x86_64",
		PackageManager:  "apt",
		Eol:             "true",
		LicenseId:       "LIC123",
	}

	params := rp.UrlParams()
	assert.Equal(t, "1.2.3", params.Get("v"))
	assert.Equal(t, "ubuntu", params.Get("p"))
	assert.Equal(t, "22.04", params.Get("pv"))
	assert.Equal(t, "x86_64", params.Get("m"))
	assert.Equal(t, "apt", params.Get("pm"))
	assert.Equal(t, "true", params.Get("eol"))
	assert.Equal(t, "LIC123", params.Get("license_id"))
}

func TestNew(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	client := New(log, "https://omnitruck.chef.io")
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.log)
}

func TestOmnitruck_Get_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"message":"ok"}`))
	}))
	defer ts.Close()

	ot := New(logrus.NewEntry(logrus.New()), ts.URL)
	ot.client = ts.Client()

	resp := ot.Get(ts.URL)
	assert.True(t, resp.Ok)
	assert.Equal(t, 200, resp.Code)
}

func TestOmnitruck_Get_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`not found`))
	}))
	defer ts.Close()

	ot := New(logrus.NewEntry(logrus.New()), ts.URL)
	ot.client = ts.Client()

	resp := ot.Get(ts.URL)
	assert.False(t, resp.Ok)
	assert.Equal(t, 400, resp.Code)
}

func TestOmnitruck_Get_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`error`))
	}))
	defer ts.Close()

	ot := New(logrus.NewEntry(logrus.New()), ts.URL)
	ot.client = ts.Client()

	resp := ot.Get(ts.URL)
	assert.False(t, resp.Ok)
	assert.Equal(t, 500, resp.Code)
}

func TestOmnitruck_Platforms(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/platforms", r.URL.Path)
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	ot := New(logrus.NewEntry(logrus.New()), ts.URL)
	ot.client = ts.Client()

	req := ot.Platforms()
	assert.True(t, req.Ok)
	assert.Equal(t, 200, req.Code)
}

func TestOmnitruck_Architectures(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/architectures", r.URL.Path)
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	ot := New(logrus.NewEntry(logrus.New()), ts.URL)
	ot.client = ts.Client()

	req := ot.Architectures()
	assert.True(t, req.Ok)
	assert.Equal(t, 200, req.Code)
}

func TestOmnitruck_LatestVersion(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/stable/test-product/versions/latest")
		w.WriteHeader(200)
		w.Write([]byte(`{"version":"1.2.3"}`))
	}))
	defer ts.Close()

	ot := New(logrus.NewEntry(logrus.New()), ts.URL)
	ot.client = ts.Client()

	p := &RequestParams{
		Channel: "stable",
		Product: "test-product",
	}

	req := ot.LatestVersion(p)
	assert.True(t, req.Ok)
	assert.Equal(t, 200, req.Code)
}

func TestOmnitruck_ProductVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/stable/test-product/versions/all")
		w.WriteHeader(200)
		w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	ot := New(logrus.NewEntry(logrus.New()), ts.URL)
	ot.client = ts.Client()

	p := &RequestParams{
		Channel: "stable",
		Product: "test-product",
	}

	req := ot.ProductVersions(p)
	assert.True(t, req.Ok)
	assert.Equal(t, 200, req.Code)
}

func TestOmnitruck_ProductPackages(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/stable/test-product/packages")
		assert.Contains(t, r.URL.RawQuery, "v=1.2.3")
		w.WriteHeader(200)
		w.Write([]byte(`[]`))
	}))
	defer ts.Close()

	ot := New(logrus.NewEntry(logrus.New()), ts.URL)
	ot.client = ts.Client()

	p := &RequestParams{
		Channel: "stable",
		Product: "test-product",
		Version: "1.2.3",
	}

	req := ot.ProductPackages(p)
	assert.True(t, req.Ok)
	assert.Equal(t, 200, req.Code)
}

func TestOmnitruck_ProductMetadata(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/stable/test-product/metadata")
		assert.Contains(t, r.URL.RawQuery, "v=1.2.3")
		assert.Contains(t, r.URL.RawQuery, "p=ubuntu")
		assert.Contains(t, r.URL.RawQuery, "pv=22.04")
		assert.Contains(t, r.URL.RawQuery, "m=x86_64")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	ot := New(logrus.NewEntry(logrus.New()), ts.URL)
	ot.client = ts.Client()

	p := &RequestParams{
		Channel:         "stable",
		Product:         "test-product",
		Version:         "1.2.3",
		Platform:        "ubuntu",
		PlatformVersion: "22.04",
		Architecture:    "x86_64",
	}

	req := ot.ProductMetadata(p)
	assert.True(t, req.Ok)
	assert.Equal(t, 200, req.Code)
}
