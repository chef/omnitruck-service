package strategy

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	s3aws "github.com/chef/omnitruck-service/clients/omnitruck/aws"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Patch s3aws functions for testing
func patchS3AWS() func() {
	origValidate := s3aws.ValidateS3Config
	origNewSession := s3aws.NewS3Session
	origNewCreds := s3aws.NewS3Credentials
	origGetObj := s3aws.GetS3Object

	s3aws.ValidateS3Config = func(cfg config.AWSConfig) error {
		return s3aws.MockValidateS3ConfigFunc(cfg)
	}
	s3aws.NewS3Session = func(region string) (*session.Session, error) {
		return s3aws.MockNewS3SessionFunc(region)
	}
	s3aws.NewS3Credentials = func(sess *session.Session, roleArn string) *credentials.Credentials {
		return s3aws.MockNewS3CredentialsFunc(sess, roleArn)
	}
	s3aws.GetS3Object = func(ctx context.Context, sess *session.Session, creds *credentials.Credentials, bucket, key string) (*s3.GetObjectOutput, error) {
		return s3aws.MockGetS3ObjectFunc(ctx, sess, creds, bucket, key)
	}
	return func() {
		s3aws.ValidateS3Config = origValidate
		s3aws.NewS3Session = origNewSession
		s3aws.NewS3Credentials = origNewCreds
		s3aws.GetS3Object = origGetObj
	}
}

func TestInfraProductStrategy_DownloadFromS3(t *testing.T) {
	tests := []struct {
		name                  string
		setupMocks            func()
		params                *omnitruck.RequestParams
		expectedMsg           string
		expectedCode          int
		expectError           bool
		expectResponseNotNil  bool
		expectedContentType   string
		expectedContentLength string
		expectedContentDispo  string
	}{
		{
			name: "Validation error from config",
			setupMocks: func() {
				s3aws.MockValidateS3ConfigFunc = func(cfg config.AWSConfig) error {
					return errors.New("bad config")
				}
			},
			params:       &omnitruck.RequestParams{},
			expectedMsg:  "bad config",
			expectedCode: http.StatusInternalServerError,
			expectError:  false,
		},
		{
			name: "New S3 session fails",
			setupMocks: func() {
				s3aws.MockValidateS3ConfigFunc = func(cfg config.AWSConfig) error { return nil }
				s3aws.MockNewS3SessionFunc = func(region string) (*session.Session, error) {
					return nil, errors.New("session error")
				}
			},
			params: &omnitruck.RequestParams{
				Channel:      constants.CURRENT_CHANNEL,
				Product:      "chef",
				Version:      "1.2.3",
				Platform:     "ubuntu",
				Architecture: "x86_64",
			},
			expectedMsg:  "Failed to create AWS session",
			expectedCode: http.StatusInternalServerError,
			expectError:  true,
		},
		{
			name: "Get S3 object fails",
			setupMocks: func() {
				s3aws.MockValidateS3ConfigFunc = func(cfg config.AWSConfig) error { return nil }
				s3aws.MockNewS3SessionFunc = func(region string) (*session.Session, error) { return session.NewSession() }
				s3aws.MockNewS3CredentialsFunc = func(sess *session.Session, roleArn string) *credentials.Credentials {
					return credentials.NewStaticCredentials("AKIA", "SECRET", "")
				}
				s3aws.MockGetS3ObjectFunc = func(ctx context.Context, sess *session.Session, creds *credentials.Credentials, bucket, key string) (*s3.GetObjectOutput, error) {
					return nil, errors.New("s3 error")
				}
			},
			params: &omnitruck.RequestParams{
				Channel:      constants.CURRENT_CHANNEL,
				Product:      "chef",
				Version:      "1.2.3",
				Platform:     "ubuntu",
				Architecture: "x86_64",
			},
			expectedMsg:  "Failed to get object from S3",
			expectedCode: http.StatusInternalServerError,
			expectError:  true,
		},
		{
			name: "Successful S3 download",
			setupMocks: func() {
				s3aws.MockValidateS3ConfigFunc = func(cfg config.AWSConfig) error { return nil }
				s3aws.MockNewS3SessionFunc = func(region string) (*session.Session, error) { return session.NewSession() }
				s3aws.MockNewS3CredentialsFunc = func(sess *session.Session, roleArn string) *credentials.Credentials {
					return credentials.NewStaticCredentials("AKIA", "SECRET", "")
				}
				contentType := "application/octet-stream"
				contentLength := int64(123)
				contentDisposition := "attachment; filename=file.txt"
				body := io.NopCloser(bytes.NewBufferString("testdata"))
				s3aws.MockGetS3ObjectFunc = func(ctx context.Context, sess *session.Session, creds *credentials.Credentials, bucket, key string) (*s3.GetObjectOutput, error) {
					return &s3.GetObjectOutput{
						Body:               body,
						ContentType:        &contentType,
						ContentLength:      &contentLength,
						ContentDisposition: &contentDisposition,
					}, nil
				}
			},
			params: &omnitruck.RequestParams{
				Channel:      constants.CURRENT_CHANNEL,
				Product:      "chef",
				Version:      "1.2.3",
				Platform:     "ubuntu",
				Architecture: "x86_64",
			},
			expectedMsg:           "",
			expectedCode:          0,
			expectError:           false,
			expectResponseNotNil:  true,
			expectedContentType:   "application/octet-stream",
			expectedContentLength: "123",
			expectedContentDispo:  "attachment; filename=file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer patchS3AWS()()
			tt.setupMocks()

			strategy := &InfraProductStrategy{
				AWSConfig: config.AWSConfig{
					S3Config: config.S3Config{
						Bucket:      "bucket",
						RoleArn:     "arn",
						CurrentPath: "current",
						StablePath:  "stable",
					},
					Region: "us-west-2",
				},
				Log: logrus.NewEntry(logrus.New()),
			}

			url, resp, header, msg, code, err := strategy.downloadFromS3(tt.params, "file.txt")

			assert.Equal(t, "", url)
			assert.Equal(t, tt.expectedMsg, msg)
			assert.Equal(t, tt.expectedCode, code)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectResponseNotNil {
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedContentType, header.Get("Content-Type"))
				assert.Equal(t, tt.expectedContentLength, header.Get("Content-Length"))
				assert.Equal(t, tt.expectedContentDispo, header.Get("Content-Disposition"))
			} else {
				assert.Nil(t, resp)
				assert.Nil(t, header)
			}
		})
	}
}

func TestInfraProductStrategy_GetLatestVersion(t *testing.T) {
	tests := []struct {
		name           string
		mockFunc       func(params *omnitruck.RequestParams) (omnitruck.ProductVersion, error)
		expectedOK     bool
		expectedCode   int
		expectedOutput string
		expectedMsg    string
	}{
		{
			name: "success",
			mockFunc: func(params *omnitruck.RequestParams) (omnitruck.ProductVersion, error) {
				return "1.2.3", nil
			},
			expectedOK:     true,
			expectedCode:   0,
			expectedOutput: "1.2.3",
		},
		{
			name: "error from dynamo",
			mockFunc: func(params *omnitruck.RequestParams) (omnitruck.ProductVersion, error) {
				return "", errors.New("db error")
			},
			expectedOK:     false,
			expectedCode:   http.StatusInternalServerError,
			expectedOutput: "",
			expectedMsg:    "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamo := &omnitruck.MockDynamoServices{
				VersionLatestFunc: tt.mockFunc,
			}
			strategy := &InfraProductStrategy{
				DynamoService: mockDynamo,
				Log:           logrus.NewEntry(logrus.New()),
			}
			params := &omnitruck.RequestParams{Product: "chef"}

			data, req := strategy.GetLatestVersion(params)

			assert.Equal(t, tt.expectedOK, req.Ok)
			assert.Equal(t, tt.expectedCode, req.Code)
			assert.Equal(t, tt.expectedOutput, string(data))

		})
	}
}
func TestInfraProductStrategy_GetAllVersions(t *testing.T) {
	tests := []struct {
		name           string
		mockFunc       func(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, error)
		expectedOK     bool
		expectedCode   int
		expectedData   []omnitruck.ProductVersion
	}{
		{
			name: "success",
			mockFunc: func(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, error) {
				return []omnitruck.ProductVersion{"1.2.3"}, nil
			},
			expectedOK:   true,
			expectedCode: 0,
			expectedData: []omnitruck.ProductVersion{"1.2.3"},
		},
		{
			name: "error from dynamo",
			mockFunc: func(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, error) {
				return nil, errors.New("db error")
			},
			expectedOK:   false,
			expectedCode: http.StatusInternalServerError,
			expectedData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamo := &omnitruck.MockDynamoServices{
				VersionAllFunc: tt.mockFunc,
			}
			strategy := &InfraProductStrategy{
				DynamoService: mockDynamo,
				Log:           logrus.NewEntry(logrus.New()),
			}
			params := &omnitruck.RequestParams{Product: "chef"}

			data, req := strategy.GetAllVersions(params)

			assert.Equal(t, tt.expectedOK, req.Ok)
			assert.Equal(t, tt.expectedCode, req.Code)
			assert.Equal(t, tt.expectedData, data)
		})
	}
}

func TestInfraProductStrategy_DownloadAndGetFileName(t *testing.T) {
	tests := []struct {
		name           string
		mockFilenameFn func(*omnitruck.RequestParams) (string, error)
		expectedFile   string
		expectedMsg    string
		expectedCode   int
		expectErr      bool
		isDownload     bool 
	}{
		{
			name: "download: filename error",
			mockFilenameFn: func(params *omnitruck.RequestParams) (string, error) {
				return "", errors.New("db error")
			},
			expectedMsg:  "", 
			expectedCode: http.StatusInternalServerError,
			expectErr:    true,
			isDownload:   true,
		},
		{
			name: "download: empty filename",
			mockFilenameFn: func(params *omnitruck.RequestParams) (string, error) {
				return "", nil
			},
			expectedMsg:  "Download filename is empty",
			expectedCode: http.StatusInternalServerError,
			expectErr:    false,
			isDownload:   true,
		},
		{
			name: "get filename: success",
			mockFilenameFn: func(params *omnitruck.RequestParams) (string, error) {
				return "file.deb", nil
			},
			expectedFile: "file.deb",
			expectErr:    false,
			isDownload:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamo := &omnitruck.MockDynamoServices{
				GetFilenameFunc: tt.mockFilenameFn,
			}
			strategy := &InfraProductStrategy{
				DynamoService: mockDynamo,
				Log:           logrus.NewEntry(logrus.New()),
			}
			params := &omnitruck.RequestParams{Product: "chef"}

			if tt.isDownload {
				url, rc, hdr, msg, code, err := strategy.Download(params)
				assert.Empty(t, url)
				assert.Nil(t, rc)
				assert.Nil(t, hdr)
				assert.Equal(t, tt.expectedCode, code)
				assert.Equal(t, tt.expectedMsg, msg)
				if tt.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			} else {
				file, err := strategy.GetFileName(params)
				assert.Equal(t, tt.expectedFile, file)
				if tt.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestInfraProductStrategy_GetPackages(t *testing.T) {
	mockDynamo := &omnitruck.MockDynamoServices{
		ProductPackagesFunc: func(params *omnitruck.RequestParams) (omnitruck.PackageList, error) {
			return omnitruck.PackageList{}, nil
		},
	}
	strategy := &InfraProductStrategy{
		DynamoService: mockDynamo,
	}
	params := &omnitruck.RequestParams{Product: "chef"}
	pkgs, err := strategy.GetPackages(params)
	assert.NoError(t, err)
	assert.NotNil(t, pkgs)
}

func TestInfraProductStrategy_GetMetadata(t *testing.T) {
	tests := []struct {
		name              string
		mockMetadataFunc  func(*omnitruck.RequestParams) (omnitruck.PackageMetadata, error)
		expectedVersion   string
		expectedOk        bool
		expectedCode      int
		expectErrorString string
	}{
		{
			name: "success - returns metadata",
			mockMetadataFunc: func(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, error) {
				return omnitruck.PackageMetadata{Version: "1.2.3"}, nil
			},
			expectedVersion: "1.2.3",
			expectedOk:      true,
		},
		{
			name: "error - db failure",
			mockMetadataFunc: func(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, error) {
				return omnitruck.PackageMetadata{}, errors.New("db error")
			},
			expectedVersion:   "",
			expectedOk:        false,
			expectedCode:      http.StatusInternalServerError,
			expectErrorString: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDynamo := &omnitruck.MockDynamoServices{
				ProductMetadataFunc: tt.mockMetadataFunc,
			}
			strategy := &InfraProductStrategy{
				DynamoService: mockDynamo,
				Log:           logrus.NewEntry(logrus.New()),
			}
			params := &omnitruck.RequestParams{Product: "chef"}

			meta, req := strategy.GetMetadata(params)

			assert.Equal(t, tt.expectedOk, req.Ok)

			if tt.expectedOk {
				assert.Equal(t, tt.expectedVersion, meta.Version)
			} else {
				assert.Equal(t, http.StatusInternalServerError, req.Code)
			}
		})
	}
}

func TestInfraProductStrategy_UpdatePackages(t *testing.T) {
	strategy := &InfraProductStrategy{}
	list := omnitruck.PackageList{
		"ubuntu": {
			"20.04": {
				"x86_64": omnitruck.PackageMetadata{Version: "1.2.3"},
			},
		},
	}
	params := &omnitruck.RequestParams{
		Product: "chef", Channel: "stable", Architecture: "x86_64",
	}
	strategy.UpdatePackages(&list, params, "http://myurl")
	pkg := list["ubuntu"]["20.04"]["x86_64"]
	assert.Contains(t, pkg.Url, "http://myurl")
}
