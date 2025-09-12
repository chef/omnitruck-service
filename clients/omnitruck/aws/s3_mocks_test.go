package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	omnitruckaws "github.com/chef/omnitruck-service/clients/omnitruck/aws"
	omnitruckConfig "github.com/chef/omnitruck-service/config"
	"github.com/stretchr/testify/assert"
)

func TestMockS3Smoke(t *testing.T) {
	called := false

	omnitruckaws.MockValidateS3ConfigFunc = func(cfg omnitruckConfig.AWSConfig) error {
		called = true
		return nil
	}
	err := omnitruckaws.MockValidateS3Config(omnitruckConfig.AWSConfig{})
	assert.NoError(t, err)
	assert.True(t, called)

	omnitruckaws.MockNewS3SessionFunc = func(region string) (aws.Config, error) {
		called = true
		return aws.Config{Region: region}, nil
	}
	cfg, err := omnitruckaws.MockNewS3Session("us-east-1")
	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", cfg.Region)
	assert.True(t, called)

	omnitruckaws.MockNewS3CredentialsFunc = func(cfg aws.Config, roleArn string) aws.CredentialsProvider {
		called = true
		return credentials.NewStaticCredentialsProvider("fake", "fake", "")
	}
	creds := omnitruckaws.MockNewS3Credentials(cfg, "role")
	assert.NotNil(t, creds)
	assert.True(t, called)

	omnitruckaws.MockGetS3ObjectFunc = func(ctx context.Context, cfg aws.Config, creds aws.CredentialsProvider, bucket, key string) (*s3.GetObjectOutput, error) {
		called = true
		return &s3.GetObjectOutput{}, nil
	}
	obj, err := omnitruckaws.MockGetS3Object(context.Background(), cfg, creds, "bucket", "key")
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.True(t, called)
}
