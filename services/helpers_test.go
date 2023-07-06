package services

import (
	"reflect"
	"testing"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	_ "github.com/chef/omnitruck-service/docs"
)

func Test_buildEndpointUrl(t *testing.T) {
	type args struct {
		baseUrl  string
		endpoint string
		params   omnitruck.RequestParams
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "download endpoint",
			args: args{
				baseUrl:  "https://commercial.chef.io",
				endpoint: "download",
				params: omnitruck.RequestParams{
					Channel: "stable",
					Product: "chef",
					Eol:     "true",
					Version: "1.0",
				},
			},
			want: "https://commercial.chef.io/stable/chef/download?eol=true&v=1.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildEndpointUrl(tt.args.baseUrl, tt.args.endpoint, &tt.args.params).String(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildEndpointUrl() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

type testContext struct {
	params  map[string]string
	query   map[string]string
	baseUrl string
}

func (tc *testContext) Params(k string, defaultValues ...string) string {
	return tc.params[k]
}

func (tc *testContext) Query(k string, defaultValues ...string) string {
	return tc.query[k]
}

func (tc *testContext) BaseURL() string {
	return tc.baseUrl
}

func Test_getDownloadUrl(t *testing.T) {
	type args struct {
		params *omnitruck.RequestParams
		c      omnitruck.FiberContext
	}

	tests := []struct {
		name string
		args *testContext
		want string
	}{
		// TODO: Add test cases.
		{
			name: "default",
			args: &testContext{
				baseUrl: "https://commercial.chef.io",
				params: map[string]string{
					"channel": "stable",
					"product": "chef",
				},
				query: map[string]string{
					"v":          "1.0",
					"p":          "el",
					"license_id": "12345",
				},
			},
			want: "https://commercial.chef.io/stable/chef/download?license_id=12345&p=el&v=1.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDownloadUrl(getRequestParams(tt.args), tt.args); got != tt.want {
				t.Errorf("getDownloadUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRequestParams(t *testing.T) {
	type args struct {
		c omnitruck.FiberContext
	}
	tests := []struct {
		name string
		ctx  omnitruck.FiberContext
		want *omnitruck.RequestParams
	}{
		{
			name: "default",
			ctx: &testContext{
				params: map[string]string{
					"channel": "stable",
					"product": "chef",
				},
				query: map[string]string{
					"v":          "1.0",
					"p":          "el",
					"pv":         "2.0",
					"m":          "x86",
					"eol":        "false",
					"license_id": "12345",
				},
			},
			want: &omnitruck.RequestParams{
				Channel:         "stable",
				Product:         "chef",
				Version:         "1.0",
				Platform:        "el",
				PlatformVersion: "2.0",
				Architecture:    "x86",
				Eol:             "false",
				LicenseId:       "12345",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRequestParams(tt.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRequestParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
