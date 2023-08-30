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
		name        string
		metadata    *models.MetaData
		args        args
		want        string
		wantErr     bool
		err         error
		version     string
		version_err error
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
			want:        "https://packages.chef.io/files/current/latest/chef-automate-cli/automate-cli.zip",
			wantErr:     false,
			err:         nil,
			version:     "latest",
			version_err: nil,
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
			want:        "",
			wantErr:     false,
			err:         nil,
			version:     "latest",
			version_err: nil,
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
			want:        "",
			version:     "latest",
			version_err: nil,
			wantErr:     true,
			err:         errors.New("ResourceNotFoundException: Requested resource not found"),
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
			want:        "https://packages.chef.io/files/stable/habitat/1.6.862/hab-x86_64-linux.tar.gz",
			wantErr:     false,
			err:         nil,
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
			version:     "",
			version_err: errors.New("ResourceNotFoundException: Requested resource not found"),
			want:        "",
			wantErr:     true,
			err:         nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture string) (*models.MetaData, error) {
				return tt.metadata, tt.err
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return tt.version, tt.version_err
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
		name        string
		metadata    *models.MetaData
		args        args
		want        PackageMetadata
		wantErr     bool
		err         error
		version     string
		version_err error
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
			version:     "latest",
			version_err: nil,
			want:        PackageMetadata{},
			wantErr:     false,
			err:         nil,
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
			version:     "latest",
			version_err: nil,
			want:        PackageMetadata{},
			wantErr:     true,
			err:         errors.New("ResourceNotFoundException: Requested resource not found"),
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
			wantErr: false,
			err:     nil,
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
			version:     "",
			version_err: errors.New("ResourceNotFoundException: Requested resource not found"),
			want:        PackageMetadata{},
			wantErr:     true,
			err:         nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture string) (*models.MetaData, error) {
				return tt.metadata, tt.err
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return tt.version, tt.version_err
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
			wantErr:     false,
			package_err: nil,
			version_err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetPackagesfunc = func(partitionValue, sortValue string) (*models.ProductDetails, error) {
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
			if (err != nil) != tt.wantErr {
				t.Errorf("DynamoServices.ProductPackages() error = %v, wantErr %v", err, tt.wantErr)
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
		name     string
		args     args
		versions []string
		want     string
		wantErr  bool
		err      error
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
			versions: []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0"},
			want:     "0.9.3",
			wantErr:  false,
			err:      nil,
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
			versions: []string{},
			want:     "",
			wantErr:  true,
			err:      errors.New("ResourceNotFoundException: Requested resource not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return tt.versions, tt.err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.FetchLatestOsVersion(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DynamoServices.ProductPackages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.ProductPackages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDynamoServices_VersionAll(t *testing.T) {
	type args struct {
		p *RequestParams
	}
	tests := []struct {
		name     string
		versions []string
		args     args
		want     []ProductVersion
		wantErr  bool
		err      error
	}{
		// TODO: Add test cases.
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
			want:    []ProductVersion{ProductVersion("0.70.0"), ProductVersion("0.71.0"), ProductVersion("0.72.0"), ProductVersion("0.73.0")},
			wantErr: false,
			err:     nil,
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
			want:    []ProductVersion{},
			wantErr: true,
			err:     errors.New("ResourceNotFoundException: Requested resource not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return tt.versions, tt.err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.VersionAll(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("DynamoServices.VersionAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.VersionAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDynamoServices_VersionLatest(t *testing.T) {
	type args struct {
		p *RequestParams
	}
	tests := []struct {
		name    string
		version string
		args    args
		want    ProductVersion
		wantErr bool
		err     error
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
			want:    ProductVersion("0.70.0"),
			wantErr: false,
			err:     nil,
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
			want:    "",
			wantErr: true,
			err:     errors.New("ResourceNotFoundException: Requested resource not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return tt.version, tt.err
			}
			svc := &DynamoServices{
				db:  mockDbService,
				log: logrus.NewEntry(logrus.New()),
			}
			got, err := svc.VersionLatest(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("DynamoServices.VersionLatest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DynamoServices.VersionLatest() = %v, want %v", got, tt.want)
			}
		})
	}
}
