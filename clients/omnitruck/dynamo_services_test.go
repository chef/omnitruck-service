package omnitruck

import (
	"errors"
	"reflect"
	"testing"

	"github.com/chef/omnitruck-service/dboperations"
	"github.com/chef/omnitruck-service/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewDBServices(t *testing.T) {
	mockDbService := new(dboperations.MockIDbOperations)
	type args struct {
		db  dboperations.IDbOperations
		log *logrus.Entry
	}
	tests := []struct {
		name string
		args args
		want DynamoServices
	}{
		{
			name: "test",
			args: args{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDynamoServices(tt.args.db, tt.args.log)
			assert.NotNil(t, got)
		})
	}
}

func TestProducts(t *testing.T) {
	mockDbService := new(dboperations.MockIDbOperations)
	type fields struct {
		db  dboperations.IDbOperations
		log *logrus.Entry
	}
	type args struct {
		p   []string
		eol string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "eol false",
			fields: fields{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			},
			args: args{
				p:   []string{"new"},
				eol: "false",
			},
			want: []string{"habitat", "new"},
		},
		{
			name: "eol true",
			fields: fields{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			},
			args: args{
				p:   []string{"new"},
				eol: "true",
			},
			want: []string{"automate-1", "habitat", "new"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DynamoServices{
				db:  tt.fields.db,
				log: tt.fields.log,
			}
			if got := svc.Products(tt.args.p, tt.args.eol); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.Products() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatforms(t *testing.T) {
	mockDbService := new(dboperations.MockIDbOperations)
	type fields struct {
		db  dboperations.IDbOperations
		log *logrus.Entry
	}
	type args struct {
		pl PlatformList
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   PlatformList
	}{
		{
			name: "success",
			fields: fields{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			},
			args: args{
				pl: PlatformList{"new": "test"},
			},
			want: PlatformList{"darwin": "Darwin", "linux": "Linux", "linux-kernel2": "Linux Kernel 2", "new": "test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DynamoServices{
				db:  tt.fields.db,
				log: tt.fields.log,
			}
			if got := svc.Platforms(tt.args.pl); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.Platforms() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProductDownload(t *testing.T) {
	type args struct {
		p *RequestParams
	}
	tests := []struct {
		name     string
		metadata *models.MetaData
		args     args
		want     string
		wantErr  bool
		err      error
	}{
		{
			name: "success",
			metadata: &models.MetaData{
				Architecture:     "amd64",
				FileName:         "automate-cli.zip",
				Platform:         "linux",
				Platform_Version: "",
				SHA1:             "",
				SHA256:           "1234",
			},
			args: args{
				p: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "",
					Architecture:    "amd64",
					Eol:             "false",
					LicenseId:       "",
				},
			},
			want:    "https://packages.chef.io/files/current/latest/chef-automate-cli/automate-cli.zip",
			wantErr: false,
			err:     nil,
		},
		{
			name:     "failure empty response",
			metadata: &models.MetaData{},
			args: args{
				p: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "1.2",
					Platform:        "linux",
					PlatformVersion: "",
					Architecture:    "x86_64",
					Eol:             "false",
					LicenseId:       "",
				},
			},
			want:    "",
			wantErr: false,
			err:     nil,
		},
		{
			name:     "failure db connection error",
			metadata: &models.MetaData{},
			args: args{
				p: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "1.2",
					Platform:        "linux",
					PlatformVersion: "",
					Architecture:    "x86_64",
					Eol:             "false",
					LicenseId:       "",
				},
			},
			want:    "",
			wantErr: true,
			err:     errors.New("ResourceNotFoundException: Requested resource not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture string) (*models.MetaData, error) {
				return tt.metadata, tt.err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}

			got, err := svc.ProductDownload(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("DynamoServices.ProductDownload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DynamoServices.ProductDownload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProductMetadata(t *testing.T) {
	type args struct {
		p *RequestParams
	}
	tests := []struct {
		name     string
		metadata *models.MetaData
		args     args
		want     PackageMetadata
		wantErr  bool
		err      error
	}{
		{
			name: "success",
			metadata: &models.MetaData{
				Architecture:     "amd64",
				FileName:         "automate-cli.zip",
				Platform:         "linux",
				Platform_Version: "",
				SHA1:             "",
				SHA256:           "1234",
			},
			args: args{
				p: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
				},
			},
			want: PackageMetadata{
				Sha1:    "",
				Sha256:  "1234",
				Url:     "",
				Version: "latest",
			},
			wantErr: false,
		},
		{
			name:     "failure empty response",
			metadata: &models.MetaData{},
			args: args{
				p: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "1.2",
					Platform:        "linux",
					PlatformVersion: "1.2",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
			},
			want:    PackageMetadata{},
			wantErr: false,
			err:     nil,
		},
		{
			name:     "failure db connection error",
			metadata: &models.MetaData{},
			args: args{
				p: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "1.2",
					Platform:        "linux",
					PlatformVersion: "1.2",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
			},
			want:    PackageMetadata{},
			wantErr: true,
			err:     errors.New("ResourceNotFoundException: Requested resource not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture string) (*models.MetaData, error) {
				return tt.metadata, tt.err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.ProductMetadata(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("DynamoServices.ProductMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.ProductMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
