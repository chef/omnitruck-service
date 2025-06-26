package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/chef/omnitruck-service/config"
)

type MockS3Session struct{}
type MockS3Creds struct{}

var (
	MockValidateS3ConfigFunc func(cfg config.AWSConfig) error
	MockNewS3SessionFunc     func(region string) (*session.Session, error)
	MockNewS3CredentialsFunc func(sess *session.Session, roleArn string) *credentials.Credentials
	MockGetS3ObjectFunc      func(ctx context.Context, sess *session.Session, creds *credentials.Credentials, bucket, key string) (*s3.GetObjectOutput, error)
)

func MockValidateS3Config(cfg config.AWSConfig) error {
	return MockValidateS3ConfigFunc(cfg)
}
func MockNewS3Session(region string) (*session.Session, error) {
	return MockNewS3SessionFunc(region)
}
func MockNewS3Credentials(sess *session.Session, roleArn string) *credentials.Credentials {
	return MockNewS3CredentialsFunc(sess, roleArn)
}
func MockGetS3Object(ctx context.Context, sess *session.Session, creds *credentials.Credentials, bucket, key string) (*s3.GetObjectOutput, error) {
	return MockGetS3ObjectFunc(ctx, sess, creds, bucket, key)
}
