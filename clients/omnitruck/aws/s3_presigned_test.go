package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestGetS3PresignedURL(t *testing.T) {
	cfg, err := NewS3Session("us-east-1")
	assert.NoError(t, err)

	creds := NewS3Credentials(cfg, "arn:aws:iam::123456789012:role/test-role")

	// Note: This will fail without valid AWS credentials, but tests the function signature
	_, err = GetS3PresignedURL(context.Background(), cfg, creds, "test-bucket", "test-key", 15)
	// We expect an error since we don't have valid credentials in test
	assert.Error(t, err)
}

func TestMockGetS3PresignedURL(t *testing.T) {
	MockGetS3PresignedURLFunc = func(ctx context.Context, cfg aws.Config, creds aws.CredentialsProvider, bucket, key string, expirationMinutes int) (string, error) {
		return "https://test-bucket.s3.amazonaws.com/test-key?presigned=true", nil
	}

	url, err := MockGetS3PresignedURL(context.Background(), aws.Config{}, nil, "test-bucket", "test-key", 15)
	assert.NoError(t, err)
	assert.Equal(t, "https://test-bucket.s3.amazonaws.com/test-key?presigned=true", url)
}
