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
	"github.com/chef/omnitruck-service/dboperations"
	"github.com/chef/omnitruck-service/models"
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

func TestDownloadFromS3_ValidateS3ConfigError(t *testing.T) {
	defer patchS3AWS()()
	s3aws.MockValidateS3ConfigFunc = func(cfg config.AWSConfig) error {
		return errors.New("bad config")
	}
	strategy := &InfraProductStrategy{
		AWSConfig: config.AWSConfig{},
		Log:       logrus.NewEntry(logrus.New()),
	}
	params := &omnitruck.RequestParams{}
	url, resp, header, msg, code, err := strategy.downloadFromS3(params, "file.txt")
	assert.Empty(t, url)
	assert.Nil(t, resp)
	assert.Nil(t, header)
	assert.Equal(t, "bad config", msg)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Nil(t, err)
}

func TestDownloadFromS3_NewS3SessionError(t *testing.T) {
	defer patchS3AWS()()
	s3aws.MockValidateS3ConfigFunc = func(cfg config.AWSConfig) error { return nil }
	s3aws.MockNewS3SessionFunc = func(region string) (*session.Session, error) {
		return nil, errors.New("session error")
	}
	strategy := &InfraProductStrategy{
		AWSConfig: config.AWSConfig{
			S3Config: config.S3Config{Bucket: "bucket", RoleArn: "arn"},
			Region:   "us-west-2",
		},
		Log: logrus.NewEntry(logrus.New()),
	}
	params := &omnitruck.RequestParams{
		Channel:      constants.CURRENT_CHANNEL,
		Product:      "chef",
		Version:      "1.2.3",
		Platform:     "ubuntu",
		Architecture: "x86_64",
	}
	url, resp, header, msg, code, err := strategy.downloadFromS3(params, "file.txt")
	assert.Empty(t, url)
	assert.Nil(t, resp)
	assert.Nil(t, header)
	assert.Equal(t, "Failed to create AWS session", msg)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Error(t, err)
}

func TestDownloadFromS3_GetS3ObjectError(t *testing.T) {
	defer patchS3AWS()()
	s3aws.MockValidateS3ConfigFunc = func(cfg config.AWSConfig) error { return nil }
	s3aws.MockNewS3SessionFunc = func(region string) (*session.Session, error) { return session.NewSession() }
	s3aws.MockNewS3CredentialsFunc = func(sess *session.Session, roleArn string) *credentials.Credentials {
		return credentials.NewStaticCredentials("AKIA", "SECRET", "")
	}
	s3aws.MockGetS3ObjectFunc = func(ctx context.Context, sess *session.Session, creds *credentials.Credentials, bucket, key string) (*s3.GetObjectOutput, error) {
		return nil, errors.New("s3 error")
	}
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
	params := &omnitruck.RequestParams{
		Channel:      constants.CURRENT_CHANNEL,
		Product:      "chef",
		Version:      "1.2.3",
		Platform:     "ubuntu",
		Architecture: "x86_64",
	}
	url, resp, header, msg, code, err := strategy.downloadFromS3(params, "file.txt")
	assert.Empty(t, url)
	assert.Nil(t, resp)
	assert.Nil(t, header)
	assert.Equal(t, "Failed to get object from S3", msg)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Error(t, err)
}

func TestDownloadFromS3_Success(t *testing.T) {
	defer patchS3AWS()()
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
	params := &omnitruck.RequestParams{
		Channel:      constants.CURRENT_CHANNEL,
		Product:      "chef",
		Version:      "1.2.3",
		Platform:     "ubuntu",
		Architecture: "x86_64",
	}
	url, resp, header, msg, code, err := strategy.downloadFromS3(params, "file.txt")
	assert.Empty(t, url)
	assert.NotNil(t, resp)
	assert.Equal(t, "application/octet-stream", header.Get("Content-Type"))
	assert.Equal(t, "123", header.Get("Content-Length"))
	assert.Equal(t, "attachment; filename=file.txt", header.Get("Content-Disposition"))
	assert.Empty(t, msg)
	assert.Equal(t, 0, code)
	assert.NoError(t, err)
}

func TestInfraProductStrategy_GetFileName(t *testing.T) {
	tests := []struct {
		name                 string
		params               *omnitruck.RequestParams
		want                 string
		mockGetMetaData      func(partitionValue, sortValue, platform, platformVersion, architecture, packageManager string) (*models.MetaData, error)
		mockGetVersionLatest func(partitionValue string) (string, error)
		wantErr              bool
		errorMsg             string
	}{
		{
			name: "chef-ice Success",
			params: &omnitruck.RequestParams{
				Channel:         constants.CURRENT_CHANNEL,
				Product:         constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT,
				Version:         "19.1.01",
				Platform:        "linux",
				PlatformVersion: "pv",
				Architecture:    "x86_64",
				PackageManager:  "deb",
			},
			want: "chef_19.1.01-1_amd64.deb",
			mockGetMetaData: func(partitionValue, sortValue, platform, platformVersion, architecture, packageManager string) (*models.MetaData, error) {
				return &models.MetaData{
					FileName: "chef_19.1.01-1_amd64.deb",
				}, nil
			},
			mockGetVersionLatest: func(partitionValue string) (string, error) {
				return "19.1.01", nil
			},
			wantErr: false,
		},
		{
			name: "automate Success",
			params: &omnitruck.RequestParams{
				Channel:         constants.STABLE_CHANNEL,
				Product:         constants.AUTOMATE_PRODUCT,
				Version:         "latest",
				Platform:        "linux",
				PlatformVersion: "pv",
				Architecture:    "amd64",
				PackageManager:  "pm",
			},
			want: "automate_cli.zip",
			mockGetMetaData: func(partitionValue, sortValue, platform, platformVersion, architecture, packageManager string) (*models.MetaData, error) {
				return &models.MetaData{
					FileName: "automate_cli.zip",
				}, nil
			},
			mockGetVersionLatest: func(partitionValue string) (string, error) {
				return "latest", nil
			},
			wantErr: false,
		},
		{
			name: "parameter validation failure",
			params: &omnitruck.RequestParams{
				Channel:         constants.CURRENT_CHANNEL,
				Product:         constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT,
				Version:         "19.1.01",
				Platform:        "linux",
				PlatformVersion: "pv",
				Architecture:    "x86_64",
			},
			want: "",
			mockGetMetaData: func(partitionValue, sortValue, platform, platformVersion, architecture, packageManager string) (*models.MetaData, error) {
				return nil, nil
			},
			wantErr:  true,
			errorMsg: "Package Manager (pm) params cannot be empty",
		},
		{
			name: "data not found for given params",
			params: &omnitruck.RequestParams{
				Channel:         constants.CURRENT_CHANNEL,
				Product:         constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT,
				Version:         "latest",
				Platform:        "linux",
				PlatformVersion: "pv",
				Architecture:    "x86_64",
				PackageManager:  "msi",
			},
			want: "",
			mockGetMetaData: func(partitionValue, sortValue, platform, platformVersion, architecture, packageManager string) (*models.MetaData, error) {
				return &models.MetaData{}, nil
			},
			mockGetVersionLatest: func(partitionValue string) (string, error) {
				return "19.1.01", nil
			},
			wantErr:  true,
			errorMsg: "Product information not found. Please check the input parameters.",
		},
		{
			name: "error in fetching latest version",
			params: &omnitruck.RequestParams{
				Channel:         constants.CURRENT_CHANNEL,
				Product:         constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT,
				Platform:        "linux",
				PlatformVersion: "pv",
				Architecture:    "x86_64",
				PackageManager:  "deb",
			},
			want: "",
			mockGetMetaData: func(partitionValue, sortValue, platform, platformVersion, architecture, packageManager string) (*models.MetaData, error) {
				return &models.MetaData{}, nil
			},
			mockGetVersionLatest: func(partitionValue string) (string, error) {
				return "", errors.New("db error")
			},
			wantErr:  true,
			errorMsg: "Error while fetching the information for the product from DB.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := dboperations.MockIDbOperations{}
			db.GetMetaDatafunc = tt.mockGetMetaData
			db.GetVersionLatestfunc = tt.mockGetVersionLatest
			log := logrus.NewEntry(logrus.New())
			dynamoService := omnitruck.NewDynamoServices(&db, log)
			s := &InfraProductStrategy{
				DynamoService: &dynamoService,
				Log:           log,
			}
			got, err := s.GetFileName(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
