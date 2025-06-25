package omnitruck

import (
	"errors"
	"reflect"
	"testing"

	"github.com/chef/omnitruck-service/dboperations"
	"github.com/chef/omnitruck-service/models"
	"github.com/chef/omnitruck-service/utils"
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
			want: []string{"chef-ice", "habitat", "new"},
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
			want: []string{"automate-1", "chef-ice", "habitat", "new"},
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
		name         string
		metadata     *models.MetaData
		args         args
		want         string
		wantErr      bool
		errMsg       string
		metadata_err error
		version      string
		version_err  error
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
			want:         "https://packages.chef.io/files/current/latest/chef-automate-cli/automate-cli.zip",
			wantErr:      false,
			errMsg:       "",
			metadata_err: nil,
			version:      "latest",
			version_err:  nil,
		},
		{
			name:     "failure validation",
			metadata: &models.MetaData{},
			args: args{
				p: &RequestParams{
					Channel:         "stble",
					Product:         "automate",
					Version:         "1.2",
					Platform:        "linux",
					PlatformVersion: "",
					Architecture:    "x86_64",
					Eol:             "false",
					LicenseId:       "",
				},
			},
			want:         "",
			wantErr:      true,
			errMsg:       "Channel can only be stable or current",
			metadata_err: nil,
			version:      "latest",
			version_err:  nil,
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
			want:         "",
			wantErr:      true,
			errMsg:       "Product information not found. Please check the input parameters.",
			metadata_err: nil,
			version:      "latest",
			version_err:  nil,
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
			want:         "",
			version:      "latest",
			version_err:  nil,
			wantErr:      true,
			errMsg:       utils.DBError,
			metadata_err: errors.New("ResourceNotFoundException: Requested resource not found"),
		},
		{
			name: "success for habitat",
			metadata: &models.MetaData{
				Architecture:     "x86_64",
				FileName:         "hab-x86_64-linux.tar.gz",
				Platform:         "linux",
				Platform_Version: "",
				SHA1:             "",
				SHA256:           "1234",
			},
			args: args{
				p: &RequestParams{
					Channel:         "stable",
					Product:         "habitat",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
			},
			version:      "1.6.862",
			version_err:  nil,
			want:         "https://packages.chef.io/files/stable/habitat/1.6.862/hab-x86_64-linux.tar.gz",
			wantErr:      false,
			errMsg:       "",
			metadata_err: nil,
		},
		{
			name:     "failure for habitat",
			metadata: &models.MetaData{},
			args: args{
				p: &RequestParams{
					Channel:         "stable",
					Product:         "habitat",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
			},
			version:      "",
			version_err:  errors.New("ResourceNotFoundException: Requested resource not found"),
			want:         "",
			wantErr:      true,
			errMsg:       utils.DBError,
			metadata_err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture string) (interface{}, error) {
				return tt.metadata, tt.metadata_err
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return tt.version, tt.version_err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}

			got, err := svc.ProductDownload(tt.args.p)
			if err != nil {
				assert.Equal(t, tt.errMsg, err.Error())
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
		name         string
		metadata     *models.MetaData
		args         args
		want         PackageMetadata
		wantErr      bool
		errMsg       string
		metadata_err error
		version      string
		version_err  error
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
			version:     "latest",
			version_err: nil,
			want: PackageMetadata{
				Sha1:    "",
				Sha256:  "1234",
				Url:     "",
				Version: "latest",
			},
			wantErr:      false,
			metadata_err: nil,
		},
		{
			name:     "failure validation",
			metadata: &models.MetaData{},
			args: args{
				p: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "1.2",
					Platform:        "",
					PlatformVersion: "1.2",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
			},
			version:      "latest",
			version_err:  nil,
			want:         PackageMetadata{},
			wantErr:      true,
			errMsg:       "Platfrom (p) params cannot be empty",
			metadata_err: nil,
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
			version:      "latest",
			version_err:  nil,
			want:         PackageMetadata{},
			wantErr:      true,
			errMsg:       "Product information not found. Please check the input parameters.",
			metadata_err: nil,
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
			version:      "latest",
			version_err:  nil,
			want:         PackageMetadata{},
			metadata_err: errors.New("ResourceNotFoundException: Requested resource not found"),
			wantErr:      true,
			errMsg:       utils.DBError,
		},
		{
			name: "success for habitat",
			metadata: &models.MetaData{
				Architecture:     "x86_64",
				FileName:         "hab-x86_64-linux.tar.gz",
				Platform:         "linux",
				Platform_Version: "",
				SHA1:             "",
				SHA256:           "1234",
			},
			args: args{
				p: &RequestParams{
					Channel:         "stable",
					Product:         "habitat",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
			},
			version:     "1.6.862",
			version_err: nil,
			want: PackageMetadata{
				Sha1:    "",
				Sha256:  "1234",
				Url:     "",
				Version: "1.6.862",
			},
			wantErr:      false,
			metadata_err: nil,
		},
		{
			name:     "failure for habitat",
			metadata: &models.MetaData{},
			args: args{
				p: &RequestParams{
					Channel:         "stable",
					Product:         "habitat",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "",
					Architecture:    "x86_64",
					Eol:             "",
					LicenseId:       "",
				},
			},
			version:      "",
			version_err:  errors.New("ResourceNotFoundException: Requested resource not found"),
			want:         PackageMetadata{},
			wantErr:      true,
			errMsg:       utils.DBError,
			metadata_err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture string) (interface{}, error) {
				return tt.metadata, tt.metadata_err
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return tt.version, tt.version_err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.ProductMetadata(tt.args.p)
			if tt.wantErr {
				assert.Equal(t, tt.errMsg, err.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.ProductMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProductPackages(t *testing.T) {
	type args struct {
		params *RequestParams
	}
	tests := []struct {
		name        string
		args        args
		version     string
		packages    models.ProductDetails
		want        PackageList
		wantErr     bool
		errMsg      string
		package_err error
		version_err error
	}{
		{
			name: "success",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "habitat",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
				},
			},
			version: "1.6.826",
			packages: models.ProductDetails{
				Product: "habitat",
				Version: "1.6.826",
				MetaData: []models.MetaData{
					{
						Architecture:     "aarch64",
						FileName:         "hab-aarch64-darwin.zip",
						Platform:         "darwin",
						Platform_Version: "",
						SHA1:             "abcde",
						SHA256:           "079e5",
					},
				},
			},
			want: map[string]PlatformVersionList{
				"darwin": {
					"pv": ArchList{
						"aarch64": PackageMetadata{
							Sha1:    "abcde",
							Sha256:  "079e5",
							Version: "1.6.826",
						},
					},
				},
			},
			wantErr:     false,
			package_err: nil,
			version_err: nil,
		},
		{
			name: "failure channel validation",
			args: args{
				params: &RequestParams{
					Channel:         "",
					Product:         "habitat",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
				},
			},
			version:     "1.6.826",
			packages:    models.ProductDetails{},
			want:        map[string]PlatformVersionList{},
			wantErr:     true,
			errMsg:      "Channel can only be stable or current",
			package_err: errors.New("ResourceNotFoundException: Requested resource not found"),
			version_err: nil,
		},
		{
			name: "failure not able to fetch latest version",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "habitat",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
				},
			},
			version:     "",
			packages:    models.ProductDetails{},
			want:        map[string]PlatformVersionList{},
			wantErr:     true,
			errMsg:      utils.DBError,
			package_err: nil,
			version_err: errors.New("ResourceNotFoundException: Requested resource not found"),
		},
		{
			name: "failure not able to fetch package details",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "habitat",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
				},
			},
			version:     "1.6.826",
			packages:    models.ProductDetails{},
			want:        map[string]PlatformVersionList{},
			wantErr:     true,
			errMsg:      utils.DBError,
			package_err: errors.New("ResourceNotFoundException: Requested resource not found"),
			version_err: nil,
		},
		{
			name: "failure empty response",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "habitat",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
				},
			},
			version:     "1.6.826",
			packages:    models.ProductDetails{},
			want:        map[string]PlatformVersionList{},
			wantErr:     true,
			errMsg:      "Product information not found. Please check the input parameters.",
			package_err: nil,
			version_err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetPackagesfunc = func(partitionValue, sortValue string) (interface{}, error) {
				return &tt.packages, tt.package_err
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return tt.version, tt.version_err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.ProductPackages(tt.args.params)
			if tt.wantErr {
				assert.Equal(t, tt.errMsg, err.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.ProductPackages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchLatestOsVersion(t *testing.T) {
	type args struct {
		params *RequestParams
	}
	tests := []struct {
		name         string
		args         args
		versions     []string
		want         string
		wantErr      bool
		errMsg       string
		versions_err error
	}{
		{
			name: "success",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "habitat",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
				},
			},
			versions:     []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0"},
			want:         "0.9.3",
			wantErr:      false,
			versions_err: nil,
		},
		{
			name: "failure channel validation",
			args: args{
				params: &RequestParams{
					Channel:         "",
					Product:         "habitat",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
				},
			},
			versions:     []string{},
			want:         "",
			wantErr:      true,
			errMsg:       "Channel can only be stable or current",
			versions_err: nil,
		},
		{
			name: "failure",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "habitat",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
				},
			},
			versions:     []string{},
			want:         "",
			wantErr:      true,
			errMsg:       utils.DBError,
			versions_err: errors.New("ResourceNotFoundException: Requested resource not found"),
		},
		{
			name: "failure no version found",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "habitat",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
				},
			},
			versions:     []string{},
			want:         "",
			wantErr:      true,
			errMsg:       "Product information not found. Please check the input parameters.",
			versions_err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return tt.versions, tt.versions_err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.FetchLatestOsVersion(tt.args.params)
			if tt.wantErr {
				assert.Equal(t, tt.errMsg, err.Error())
				//return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.ProductPackages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionAll(t *testing.T) {
	type args struct {
		p *RequestParams
	}
	tests := []struct {
		name         string
		versions     []string
		args         args
		want         []ProductVersion
		wantErr      bool
		errMsg       string
		versions_err error
	}{
		{
			name:     "Success",
			versions: []string{"0.70.0", "0.71.0", "0.72.0", "0.73.0"},
			args: args{
				p: &RequestParams{
					Channel:   "stable",
					Product:   "habitat",
					Eol:       "",
					LicenseId: "",
				},
			},
			want:         []ProductVersion{ProductVersion("0.70.0"), ProductVersion("0.71.0"), ProductVersion("0.72.0"), ProductVersion("0.73.0")},
			wantErr:      false,
			versions_err: nil,
		},
		{
			name: "Success for chef-ice product",
			versions: []string{"0.70.0", "0.71.0", "0.72.0", "0.73.0"},
			args: args{
				p: &RequestParams{	
					Channel:   "stable",
					Product:   "chef-ice",
					Eol:       "",
					LicenseId: "",
				},
			},
			want:        []ProductVersion{ProductVersion("0.70.0"), ProductVersion("0.71.0"), ProductVersion("0.72.0"), ProductVersion("0.73.0")},
			wantErr:     false,
			versions_err: nil,
		},
		{
			name:     "Failure validation",
			versions: []string{},
			args: args{
				p: &RequestParams{
					Channel:   "",
					Product:   "habitat",
					Eol:       "",
					LicenseId: "",
				},
			},
			want:         []ProductVersion{},
			wantErr:      true,
			errMsg:       "Channel can only be stable or current",
			versions_err: nil,
		},
		{
			name:     "Fail",
			versions: []string{},
			args: args{
				p: &RequestParams{
					Channel:   "stable",
					Product:   "habitat",
					Eol:       "",
					LicenseId: "",
				},
			},
			want:         []ProductVersion{},
			wantErr:      true,
			errMsg:       "Error while fetching product versions",
			versions_err: errors.New("ResourceNotFoundException: Requested resource not found"),
		},
		{
			name:     "Fail",
			versions: []string{},
			args: args{
				p: &RequestParams{
					Channel:   "stable",
					Product:   "habitat",
					Eol:       "",
					LicenseId: "",
				},
			},
			want:         []ProductVersion{},
			wantErr:      true,
			errMsg:       "Product information not found. Please check the input parameters.",
			versions_err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return tt.versions, tt.versions_err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.VersionAll(tt.args.p)
			if tt.wantErr {
				assert.Equal(t, tt.errMsg, err.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.VersionAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionLatest(t *testing.T) {
	type args struct {
		p *RequestParams
	}
	tests := []struct {
		name        string
		version     string
		args        args
		want        ProductVersion
		wantErr     bool
		errMsg      string
		version_err error
	}{
		{
			name:    "Success",
			version: "0.70.0",
			args: args{
				p: &RequestParams{
					Channel:   "stable",
					Product:   "habitat",
					Eol:       "",
					LicenseId: "",
				},
			},
			want:        ProductVersion("0.70.0"),
			wantErr:     false,
			version_err: nil,
		},
		{
			name:    "Failure validation",
			version: "",
			args: args{
				p: &RequestParams{
					Channel:   "",
					Product:   "habitat",
					Eol:       "",
					LicenseId: "",
				},
			},
			want:        "",
			wantErr:     true,
			errMsg:      "Channel can only be stable or current",
			version_err: nil,
		},
		{
			name: "Success for chef-ice product",
			version: "0.70.0",
			args: args{
				p: &RequestParams{	
					Channel:   "stable",
					Product:   "chef-ice",
					Eol:       "",
					LicenseId: "",
				},
			},
			want:        ProductVersion("0.70.0"),
			wantErr:     false,
			version_err: nil,
		},
		{
			name:    "Fail",
			version: "",
			args: args{
				p: &RequestParams{
					Channel:   "stable",
					Product:   "habitat",
					Eol:       "",
					LicenseId: "",
				},
			},
			want:        "",
			wantErr:     true,
			errMsg:      utils.DBError,
			version_err: errors.New("ResourceNotFoundException: Requested resource not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return tt.version, tt.version_err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.VersionLatest(tt.args.p)
			if tt.wantErr {
				assert.Equal(t, tt.errMsg, err.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.VersionLatest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRelatedProducts(t *testing.T) {
	type args struct {
		params *RequestParams
	}
	tests := []struct {
		name                   string
		args                   args
		want                   *models.RelatedProducts
		wantErr                bool
		errMsg                 string
		getRelatedProducts     *models.RelatedProducts
		getRelatedProducts_err error
	}{
		{
			name: "success",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
					BOM:             "Chef Automate",
				},
			},
			want: &models.RelatedProducts{
				Bom:      "Chef Automate",
				Products: map[string]string{"Chef Automate": "automate"},
			},
			wantErr: false,
			errMsg:  "",
			getRelatedProducts: &models.RelatedProducts{
				Bom:      "Chef Automate",
				Products: map[string]string{"Chef Automate": "automate"},
			},
			getRelatedProducts_err: nil,
		},
		{
			name: "failure validation",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
			},
			want:    &models.RelatedProducts{},
			wantErr: true,
			errMsg:  "BOM (bom) params cannot be empty",
			getRelatedProducts: &models.RelatedProducts{
				Bom:      "Chef Automate",
				Products: map[string]string{"Chef Automate": "automate"},
			},
			getRelatedProducts_err: nil,
		},
		{
			name: "failure db connection err",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
					BOM:             "Chef Automate",
				},
			},
			want:                   &models.RelatedProducts{},
			wantErr:                true,
			errMsg:                 utils.DBError,
			getRelatedProducts:     &models.RelatedProducts{},
			getRelatedProducts_err: errors.New("ResourceNotFoundException: Requested resource not found"),
		},
		{
			name: "failure no related products found",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "",
					Version:         "",
					Platform:        "",
					PlatformVersion: "",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
					BOM:             "Chef Automate",
				},
			},
			want:                   &models.RelatedProducts{},
			wantErr:                true,
			errMsg:                 "Product information not found. Please check the input parameters.",
			getRelatedProducts:     &models.RelatedProducts{},
			getRelatedProducts_err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetRelatedProductsfunc = func(partitionValue string) (*models.RelatedProducts, error) {
				return tt.getRelatedProducts, tt.getRelatedProducts_err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.GetRelatedProducts(tt.args.params)
			if tt.wantErr {
				assert.Equal(t, tt.errMsg, err.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.GetRelatedProducts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFilename(t *testing.T) {
	type args struct {
		params *RequestParams
	}
	tests := []struct {
		name         string
		args         args
		want         string
		wantErr      bool
		errMsg       string
		metadata     *models.MetaData
		metadata_err error
		version      string
		version_err  error
	}{
		{
			name: "success",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
			},
			want:    "automate_cli.zip",
			wantErr: false,
			errMsg:  "",
			metadata: &models.MetaData{
				Architecture:     "amd64",
				FileName:         "automate_cli.zip",
				Platform:         "linux",
				Platform_Version: "",
				SHA1:             "abcd",
				SHA256:           "",
			},
			metadata_err: nil,
			version:      "latest",
			version_err:  nil,
		},
		{
			name: "failure validation",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
			},
			want:    "automate_cli.zip",
			wantErr: true,
			errMsg:  "Architecture (m) params cannot be empty",
			metadata: &models.MetaData{
				Architecture:     "amd64",
				FileName:         "automate_cli.zip",
				Platform:         "linux",
				Platform_Version: "",
				SHA1:             "abcd",
				SHA256:           "",
			},
			metadata_err: nil,
			version:      "",
			version_err:  nil,
		},
		{
			name: "failure db connection err for latest version fetch",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
			},
			want:    "",
			wantErr: true,
			errMsg:  utils.DBError,
			metadata: &models.MetaData{
				Architecture:     "amd64",
				FileName:         "automate_cli.zip",
				Platform:         "linux",
				Platform_Version: "",
				SHA1:             "abcd",
				SHA256:           "",
			},
			metadata_err: nil,
			version:      "",
			version_err:  errors.New("ResourceNotFoundException: Requested resource not found"),
		},
		{
			name: "failure db connection err for metadata fetch",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
			},
			want:         "",
			wantErr:      true,
			errMsg:       utils.DBError,
			metadata:     &models.MetaData{},
			metadata_err: errors.New("ResourceNotFoundException: Requested resource not found"),
			version:      "",
			version_err:  nil,
		},
		{
			name: "failure metadata not found",
			args: args{
				params: &RequestParams{
					Channel:         "stable",
					Product:         "automate",
					Version:         "",
					Platform:        "linux",
					PlatformVersion: "pv",
					Architecture:    "amd64",
					Eol:             "",
					LicenseId:       "",
					BOM:             "",
				},
			},
			want:         "",
			wantErr:      true,
			errMsg:       "Product information not found. Please check the input parameters.",
			metadata:     &models.MetaData{},
			metadata_err: nil,
			version:      "",
			version_err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture string) (interface{}, error) {
				return tt.metadata, tt.metadata_err
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return tt.version, tt.version_err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.GetFilename(tt.args.params)
			if tt.wantErr {
				assert.Equal(t, tt.errMsg, err.Error())
				return
			}
			if got != tt.want {
				t.Errorf("DynamoServices.GetFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPackageManagers(t *testing.T) {
	mockDbService := new(dboperations.MockIDbOperations)

	tests := []struct {
		name           string
		mockResult     []string
		mockError      error
		expectedResult []string
		expectError    bool
	}{
		{
			name:           "success - returns package managers",
			mockResult:     []string{"yum", "apt", "zypper"},
			mockError:      nil,
			expectedResult: []string{"yum", "apt", "zypper"},
			expectError:    false,
		},
		{
			name:           "failure - db error",
			mockResult:     nil,
			mockError:      errors.New("database unavailable"),
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService.GetPackageManagersfunc = func() ([]string, error) {
				return tt.mockResult, tt.mockError
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.GetPackageManagers()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, got)
		})
	}
}
