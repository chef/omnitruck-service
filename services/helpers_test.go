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
