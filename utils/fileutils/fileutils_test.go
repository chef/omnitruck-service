package fileutils_test

import (
	"errors"
	"testing"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/utils/fileutils"
	"github.com/stretchr/testify/assert"
)

const (
	HOST_NAME = "example.com"
	LICENSE_ID = "afd2c0a2-111f-4caf-1fa2-1211fe1212d"
)

func TestGetScript(t *testing.T) {
	type args struct {
		baseUrl  string
		params   *omnitruck.RequestParams
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "success when file is shell script",
			args: args{
				baseUrl: HOST_NAME,
				params: &omnitruck.RequestParams{
					LicenseId: LICENSE_ID,
				},
				filePath: "./testfiles/test_success.sh.tmpl",
			},
			want:    "#!/bin/sh",
			wantErr: nil,
		},
		{
			name: "success when file is pwershellscript",
			args: args{
				baseUrl: HOST_NAME,
				params: &omnitruck.RequestParams{
					LicenseId: LICENSE_ID,
				},
				filePath: "./testfiles/test_success.ps1.tmpl",
			},
			want:    "new-module -name TestFile",
			wantErr: nil,
		},
		{
			name: "error while parsing the file",
			args: args{
				baseUrl: HOST_NAME,
				params: &omnitruck.RequestParams{
					LicenseId: LICENSE_ID,
				},
				filePath: "./testfiles/test_success1.sh.tmpl",
			},
			want:    "",
			wantErr: errors.New("error while parsing the template files: open ./testfiles/test_success1.sh.tmpl: no such file or directory"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fu := &fileutils.FileUtilsImpl{}
			got, err := fu.GetScript(tt.args.baseUrl, tt.args.params, tt.args.filePath)
			if err != nil {
				assert.Equal(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}
