package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	omnitruckConfig "github.com/chef/omnitruck-service/config"
)

type MockS3Session struct{}
type MockS3Creds struct{}

var (
	MockValidateS3ConfigFunc func(cfg omnitruckConfig.AWSConfig) error
	MockNewS3SessionFunc     func(region string) (aws.Config, error)
	MockNewS3CredentialsFunc func(cfg aws.Config, roleArn string) aws.CredentialsProvider
	MockGetS3ObjectFunc      func(ctx context.Context, cfg aws.Config, creds aws.CredentialsProvider, bucket, key string) (*s3.GetObjectOutput, error)
	MockGetS3PresignedURLFunc func(ctx context.Context, cfg aws.Config, creds aws.CredentialsProvider, bucket, key string, expirationMinutes int) (string, error)
)

func MockValidateS3Config(cfg omnitruckConfig.AWSConfig) error {
	return MockValidateS3ConfigFunc(cfg)
}
func MockNewS3Session(region string) (aws.Config, error) {
	return MockNewS3SessionFunc(region)
}
func MockNewS3Credentials(cfg aws.Config, roleArn string) aws.CredentialsProvider {
	return MockNewS3CredentialsFunc(cfg, roleArn)
}
func MockGetS3Object(ctx context.Context, cfg aws.Config, creds aws.CredentialsProvider, bucket, key string) (*s3.GetObjectOutput, error) {
	return MockGetS3ObjectFunc(ctx, cfg, creds, bucket, key)
}
func MockGetS3PresignedURL(ctx context.Context, cfg aws.Config, creds aws.CredentialsProvider, bucket, key string, expirationMinutes int) (string, error) {
	return MockGetS3PresignedURLFunc(ctx, cfg, creds, bucket, key, expirationMinutes)
}
