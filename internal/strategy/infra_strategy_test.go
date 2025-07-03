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

func TestInfraProductStrategy_GetLatestVersion(t *testing.T) {
	mockDynamo := &omnitruck.MockDynamoServices{
		VersionLatestFunc: func(params *omnitruck.RequestParams) (omnitruck.ProductVersion, error) {
			return omnitruck.ProductVersion("1.2.3"), nil
		},
	}
	strategy := &InfraProductStrategy{
		DynamoService: mockDynamo,
		Log:           logrus.NewEntry(logrus.New()),
	}
	params := &omnitruck.RequestParams{Product: "chef"}
	data, req := strategy.GetLatestVersion(params)
	assert.True(t, req.Ok)
	assert.Equal(t, "1.2.3", string(data))
}

func TestInfraProductStrategy_GetLatestVersion_Error(t *testing.T) {
	mockDynamo := &omnitruck.MockDynamoServices{
		VersionLatestFunc: func(params *omnitruck.RequestParams) (omnitruck.ProductVersion, error) {
			return "", errors.New("db error")
		},
	}
	strategy := &InfraProductStrategy{
		DynamoService: mockDynamo,
		Log:           logrus.NewEntry(logrus.New()),
	}
	params := &omnitruck.RequestParams{Product: "chef"}
	_, req := strategy.GetLatestVersion(params)
	assert.False(t, req.Ok)
	assert.Equal(t, http.StatusInternalServerError, req.Code)
}

func TestInfraProductStrategy_GetAllVersions(t *testing.T) {
	mockDynamo := &omnitruck.MockDynamoServices{
		VersionAllFunc: func(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, error) {
			return []omnitruck.ProductVersion{"1.2.3"}, nil
		},
	}
	strategy := &InfraProductStrategy{
		DynamoService: mockDynamo,
		Log:           logrus.NewEntry(logrus.New()),
	}
	params := &omnitruck.RequestParams{Product: "chef"}
	data, req := strategy.GetAllVersions(params)
	assert.True(t, req.Ok)
	assert.Len(t, data, 1)
	assert.Equal(t, "1.2.3", string(data[0]))
}

func TestInfraProductStrategy_GetAllVersions_Error(t *testing.T) {
	mockDynamo := &omnitruck.MockDynamoServices{
		VersionAllFunc: func(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, error) {
			return nil, errors.New("db error")
		},
	}
	strategy := &InfraProductStrategy{
		DynamoService: mockDynamo,
		Log:           logrus.NewEntry(logrus.New()),
	}
	params := &omnitruck.RequestParams{Product: "chef"}
	data, req := strategy.GetAllVersions(params)
	assert.False(t, req.Ok)
	assert.Nil(t, data)
	assert.Equal(t, http.StatusInternalServerError, req.Code)
	t.Logf("Error message: %s", req.Message)
}

func TestInfraProductStrategy_Download_FilenameError(t *testing.T) {
	mockDynamo := &omnitruck.MockDynamoServices{
		GetFilenameFunc: func(params *omnitruck.RequestParams) (string, error) {
			return "", errors.New("db error")
		},
	}
	strategy := &InfraProductStrategy{
		DynamoService: mockDynamo,
		Log:           logrus.NewEntry(logrus.New()),
	}
	params := &omnitruck.RequestParams{Product: "chef"}
	url, rc, hdr, msg, code, err := strategy.Download(params)
	assert.Empty(t, url)
	assert.Nil(t, rc)
	assert.Nil(t, hdr)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Error(t, err)
	t.Logf("Download error message (may be empty): '%s'", msg)
}

func TestInfraProductStrategy_Download_EmptyFilename(t *testing.T) {
	mockDynamo := &omnitruck.MockDynamoServices{
		GetFilenameFunc: func(params *omnitruck.RequestParams) (string, error) {
			return "", nil
		},
	}
	strategy := &InfraProductStrategy{
		DynamoService: mockDynamo,
		Log:           logrus.NewEntry(logrus.New()),
	}
	params := &omnitruck.RequestParams{Product: "chef"}
	url, rc, hdr, msg, code, err := strategy.Download(params)
	assert.Empty(t, url)
	assert.Nil(t, rc)
	assert.Nil(t, hdr)
	assert.Equal(t, "Download filename is empty", msg)
	assert.Equal(t, 500, code)
	assert.Nil(t, err)
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
	mockDynamo := &omnitruck.MockDynamoServices{
		ProductMetadataFunc: func(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, error) {
			return omnitruck.PackageMetadata{Version: "1.2.3"}, nil
		},
	}
	strategy := &InfraProductStrategy{
		DynamoService: mockDynamo,
		Log:           logrus.NewEntry(logrus.New()),
	}
	params := &omnitruck.RequestParams{Product: "chef"}
	data, req := strategy.GetMetadata(params)
	assert.True(t, req.Ok)
	assert.Equal(t, "1.2.3", data.Version)
}

func TestInfraProductStrategy_GetMetadata_Error(t *testing.T) {
	mockDynamo := &omnitruck.MockDynamoServices{
		ProductMetadataFunc: func(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, error) {
			return omnitruck.PackageMetadata{}, errors.New("db error")
		},
	}
	strategy := &InfraProductStrategy{
		DynamoService: mockDynamo,
		Log:           logrus.NewEntry(logrus.New()),
	}
	params := &omnitruck.RequestParams{Product: "chef"}
	_, req := strategy.GetMetadata(params)
	assert.False(t, req.Ok)
	assert.Equal(t, http.StatusInternalServerError, req.Code)
	t.Logf("Error message was: %s", req.Message)
}

func TestInfraProductStrategy_GetFileName(t *testing.T) {
	mockDynamo := &omnitruck.MockDynamoServices{
		GetFilenameFunc: func(params *omnitruck.RequestParams) (string, error) {
			return "file.deb", nil
		},
	}
	strategy := &InfraProductStrategy{
		DynamoService: mockDynamo,
	}
	params := &omnitruck.RequestParams{Product: "chef"}
	file, err := strategy.GetFileName(params)
	assert.NoError(t, err)
	assert.Equal(t, "file.deb", file)
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
