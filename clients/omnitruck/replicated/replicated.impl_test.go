package replicated_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/logger"
	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (mc MockClient) Do(req *http.Request) (*http.Response, error) {
	return mc.DoFunc(req)
}

func TestSearchCustomerByEmailSuccess(t *testing.T) {
	repImp := replicated.ReplicatedImpl{
		Client: &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`{
					"query": "email:test@progress.com",
					"total_hits": 1,
					"customers": [
						{
							"id": "2eDnZGGGSjwJN91WZh74A",
							"teamId": "6IopgEQswci9pZWVGGNHU4NyoRbaWe7d",
							"name": "[DEV] Test",
							"email": "test@progress.com",
							"customId": ""
						}
					]
				}`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, nil
			},
		},
		ReplicatedConfig: config.ReplicatedConfig{},
		Logger:           logger.NewLogrusStandardLogger(),
	}
	customers, err := repImp.SearchCustomersByEmail("s-no@progress.com", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(customers))

}

func TestSearchCustomerByEmailError(t *testing.T) {
	repImp := replicated.ReplicatedImpl{
		Client: &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("catastrophic failure")
			},
		},
		ReplicatedConfig: config.ReplicatedConfig{},
		Logger:           logger.NewLogrusStandardLogger(),
	}
	_, err := repImp.SearchCustomersByEmail("s-no@progress.com", "")
	assert.Error(t, err)
}

func TestSearchCustomerByEmail500(t *testing.T) {
	repImp := replicated.ReplicatedImpl{
		Client: &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`{
					"message": "error occurred"
				}`)))
				return &http.Response{
					StatusCode: 500,
					Body:       responseBody,
				}, nil
			},
		},
		ReplicatedConfig: config.ReplicatedConfig{},
		Logger:           logger.NewLogrusStandardLogger(),
	}
	_, err := repImp.SearchCustomersByEmail("s-no@progress.com", "")
	assert.Error(t, err)
}

// func TestReplicatedImpl_SearchCustomersByEmail(t *testing.T) {
// 	rc := config.ReplicatedConfig{
// 		URL:   "https://api.replicated.com/vendor/v3",
// 		Token: "e69364a39827225d1042609c4e461e336093305e8a7f58cd47da97107075738d",
// 		AppID: "2dbKte6a9ecfZo6Mn0KTjRvDak4",
// 	}
// 	r := replicated.NewReplicatedImpl(rc, logger.NewLogrusStandardLogger())
// 	customers, err := r.SearchCustomersByEmail("george.westwater@progress.com", "abc123")
// 	assert.NoError(t, err)
// 	assert.Equal(t, 1, len(customers))
// }

func TestPlatformVersionsAll(t *testing.T) {
	type fields struct {
		ReplicatedConfig config.ReplicatedConfig
		Client           replicated.HTTPClient
		Logger           logger.Logger
	}
	type args struct {
		req        *omnitruck.RequestParams
		serverMode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []omnitruck.ProductVersion
		wantErr error
	}{
		{
			name: "success for getting the versions all",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stable",
					Product:         "chef-360",
					Version:         "latest",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 2,
			},
			want:    []omnitruck.ProductVersion{"latest"},
			wantErr: nil,
		},
		{
			name: "chef-360 failure on the non-commercial server",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stable",
					Product:         "chef-360",
					Version:         "latest",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 1,
			},
			want:    nil,
			wantErr: errors.New(constants.PLATFORM_ERROR),
		},
		{
			name: "validation failure",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stale",
					Product:         "chef-360",
					Version:         "1.2",
					Platform:        "",
					PlatformVersion: "1.2",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 2,
			},
			want:    nil,
			wantErr: errors.New("Channel can only be stable or current"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &replicated.ReplicatedImpl{
				ReplicatedConfig: tt.fields.ReplicatedConfig,
				Client:           tt.fields.Client,
				Logger:           tt.fields.Logger,
			}
			got, err := r.PlatformVersionsAll(tt.args.req, tt.args.serverMode)
			if err != nil {
				assert.Equal(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

func TestPlatformVersionLatest(t *testing.T) {
	type fields struct {
		ReplicatedConfig config.ReplicatedConfig
		Client           replicated.HTTPClient
		Logger           logger.Logger
	}
	type args struct {
		req        *omnitruck.RequestParams
		serverMode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    omnitruck.ProductVersion
		wantErr error
	}{
		{
			name: "success for getting the versions latest",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stable",
					Product:         "chef-360",
					Version:         "latest",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 2,
			},
			want:    "latest",
			wantErr: nil,
		},
		{
			name: "chef-360 failure on the non-commercial server",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stable",
					Product:         "chef-360",
					Version:         "latest",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 1,
			},
			want:    "",
			wantErr: errors.New(constants.PLATFORM_ERROR),
		},
		{
			name: "validation failure",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stale",
					Product:         "chef-360",
					Version:         "1.2",
					Platform:        "",
					PlatformVersion: "1.2",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 2,
			},
			want:    "",
			wantErr: errors.New("Channel can only be stable or current"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &replicated.ReplicatedImpl{
				ReplicatedConfig: tt.fields.ReplicatedConfig,
				Client:           tt.fields.Client,
				Logger:           tt.fields.Logger,
			}
			got, err := r.PlatformVersionLatest(tt.args.req, tt.args.serverMode)
			if err != nil {
				assert.Equal(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

func TestPlatformMetadata(t *testing.T) {
	type fields struct {
		ReplicatedConfig config.ReplicatedConfig
		Client           replicated.HTTPClient
		Logger           logger.Logger
	}
	type args struct {
		req        *omnitruck.RequestParams
		serverMode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    omnitruck.PackageMetadata
		wantErr error
	}{
		{
			name: "success for getting the metadata for the chef-360",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stable",
					Product:         "chef-360",
					Version:         "latest",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 2,
			},
			want:    omnitruck.PackageMetadata{
				Sha1: "",
				Sha256: "",
				Url: "",
				Version: "latest",
			},
			wantErr: nil,
		},
		{
			name: "chef-360 failure on the non-commercial server",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stable",
					Product:         "chef-360",
					Version:         "latest",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 1,
			},
			want:    omnitruck.PackageMetadata{},
			wantErr: errors.New(constants.PLATFORM_ERROR),
		},
		{
			name: "validation failure",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stale",
					Product:         "chef-360",
					Version:         "1.2",
					Platform:        "",
					PlatformVersion: "1.2",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 2,
			},
			want:    omnitruck.PackageMetadata{},
			wantErr: errors.New("Channel can only be stable or current"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &replicated.ReplicatedImpl{
				ReplicatedConfig: tt.fields.ReplicatedConfig,
				Client:           tt.fields.Client,
				Logger:           tt.fields.Logger,
			}
			got, err := r.PlatformMetadata(tt.args.req, tt.args.serverMode)
			if err != nil {
				assert.Equal(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

func TestPlatformPackages(t *testing.T) {
	type fields struct {
		ReplicatedConfig config.ReplicatedConfig
		Client           replicated.HTTPClient
		Logger           logger.Logger
	}
	type args struct {
		req        *omnitruck.RequestParams
		serverMode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    omnitruck.PackageList
		wantErr error
	}{
		{
			name: "success for getting the chef-360 packages details",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stable",
					Product:         "chef-360",
					Version:         "latest",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 2,
			},
			want:    map[string]omnitruck.PlatformVersionList{
				"linux": {
					"pv": omnitruck.ArchList{
						"amd64": omnitruck.PackageMetadata{
							Sha1:    "",
							Sha256:  "",
							Version: "latest",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "chef-360 failure on the non-commercial server",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stable",
					Product:         "chef-360",
					Version:         "latest",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 1,
			},
			want:    omnitruck.PackageList{},
			wantErr: errors.New(constants.PLATFORM_ERROR),
		},
		{
			name: "validation failure",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stale",
					Product:         "chef-360",
					Version:         "1.2",
					Platform:        "",
					PlatformVersion: "1.2",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 2,
			},
			want:    omnitruck.PackageList{},
			wantErr: errors.New("Channel can only be stable or current"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &replicated.ReplicatedImpl{
				ReplicatedConfig: tt.fields.ReplicatedConfig,
				Client:           tt.fields.Client,
				Logger:           tt.fields.Logger,
			}
			got, err := r.PlatformPackages(tt.args.req, tt.args.serverMode)
			if err != nil {
				assert.Equal(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

func TestPlatformFilename(t *testing.T) {
	type fields struct {
		ReplicatedConfig config.ReplicatedConfig
		Client           replicated.HTTPClient
		Logger           logger.Logger
	}
	type args struct {
		req        *omnitruck.RequestParams
		serverMode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr error
	}{
		{
			name: "success for getting the correct file name",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stable",
					Product:         "chef-360",
					Version:         "latest",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 2,
			},
			want:    "chef-360.zip",
			wantErr: nil,
		},
		{
			name: "failed for non commercial server",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stable",
					Product:         "chef-360",
					Version:         "latest",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 1,
			},
			want:    "",
			wantErr: errors.New(constants.PLATFORM_ERROR),
		},
		{
			name: "validation failure",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &omnitruck.RequestParams{
					Channel:         "stale",
					Product:         "chef-360",
					Version:         "1.2",
					Platform:        "",
					PlatformVersion: "1.2",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 2,
			},
			want:    "",
			wantErr: errors.New("Channel can only be stable or current"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &replicated.ReplicatedImpl{
				ReplicatedConfig: tt.fields.ReplicatedConfig,
				Client:           tt.fields.Client,
				Logger:           tt.fields.Logger,
			}
			got, err := r.PlatformFilename(tt.args.req, tt.args.serverMode)
			if err != nil {
				assert.Equal(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}