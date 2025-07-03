package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/chef/omnitruck-service/clients/omnitruck/aws"
	"github.com/chef/omnitruck-service/config"
	"github.com/stretchr/testify/assert"
)

func TestMockS3Smoke(t *testing.T) {
	called := false

	aws.MockValidateS3ConfigFunc = func(cfg config.AWSConfig) error {
		called = true
		return nil
	}
	err := aws.MockValidateS3Config(config.AWSConfig{})
	assert.NoError(t, err)
	assert.True(t, called)

	aws.MockNewS3SessionFunc = func(region string) (*session.Session, error) {
		called = true
		return &session.Session{}, nil
	}
	sess, err := aws.MockNewS3Session("us-east-1")
	assert.NoError(t, err)
	assert.NotNil(t, sess)
	assert.True(t, called)

	aws.MockNewS3CredentialsFunc = func(sess *session.Session, roleArn string) *credentials.Credentials {
		called = true
		return credentials.AnonymousCredentials
	}
	creds := aws.MockNewS3Credentials(sess, "role")
	assert.NotNil(t, creds)
	assert.True(t, called)

	aws.MockGetS3ObjectFunc = func(ctx context.Context, sess *session.Session, creds *credentials.Credentials, bucket, key string) (*s3.GetObjectOutput, error) {
		called = true
		return &s3.GetObjectOutput{}, nil
	}
	obj, err := aws.MockGetS3Object(context.Background(), sess, creds, "bucket", "key")
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.True(t, called)
}
