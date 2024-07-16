package omnitruck

import (
	"errors"
	"testing"

	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/logger"
	"github.com/stretchr/testify/assert"
)

func TestPlatformVersionsAll(t *testing.T) {
	type fields struct {
		Logger logger.Logger
	}
	type args struct {
		req        *RequestParams
		serverMode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []ProductVersion
		wantErr error
	}{
		{
			name: "success for getting the versions all",
			fields: fields{
				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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
			want:    []ProductVersion{"latest"},
			wantErr: nil,
		},
		{
			name: "chef-360 failure on the non-commercial server",
			fields: fields{
				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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

				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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
			r := &PlatformServices{
				Logger: tt.fields.Logger,
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
		Logger logger.Logger
	}
	type args struct {
		req        *RequestParams
		serverMode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ProductVersion
		wantErr error
	}{
		{
			name: "success for getting the versions latest",
			fields: fields{
				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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

				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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

				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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
			r := &PlatformServices{
				Logger: tt.fields.Logger,
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
		Logger logger.Logger
	}
	type args struct {
		req        *RequestParams
		serverMode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    PackageMetadata
		wantErr error
	}{
		{
			name: "success for getting the metadata for the chef-360",
			fields: fields{
				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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
			want: PackageMetadata{
				Sha1:    "",
				Sha256:  "",
				Url:     "",
				Version: "latest",
			},
			wantErr: nil,
		},
		{
			name: "chef-360 failure on the non-commercial server",
			fields: fields{

				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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
			want:    PackageMetadata{},
			wantErr: errors.New(constants.PLATFORM_ERROR),
		},
		{
			name: "validation failure",
			fields: fields{

				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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
			want:    PackageMetadata{},
			wantErr: errors.New("Channel can only be stable or current"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &PlatformServices{
				Logger: tt.fields.Logger,
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
		Logger logger.Logger
	}
	type args struct {
		req        *RequestParams
		serverMode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    PackageList
		wantErr error
	}{
		{
			name: "success for getting the chef-360 packages details",
			fields: fields{
				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
					Channel:         "stable",
					Product:         "chef-360",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
				serverMode: 2,
			},
			want: map[string]PlatformVersionList{
				"linux": {
					"pv": ArchList{
						"amd64": PackageMetadata{
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
				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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
			want:    PackageList{},
			wantErr: errors.New(constants.PLATFORM_ERROR),
		},
		{
			name: "validation failure",
			fields: fields{
				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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
			want:    PackageList{},
			wantErr: errors.New("Channel can only be stable or current"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &PlatformServices{
				Logger: tt.fields.Logger,
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
		Logger logger.Logger
	}
	type args struct {
		req        *RequestParams
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
				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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
				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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

				Logger: logger.NewLogrusStandardLogger(),
			},
			args: args{
				req: &RequestParams{
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
			r := &PlatformServices{
				Logger: tt.fields.Logger,
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
