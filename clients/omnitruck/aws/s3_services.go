package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/chef/omnitruck-service/config"
)

var NewS3Session = func(region string) (*session.Session, error) {
	return session.NewSession(&aws.Config{Region: aws.String(region)})
}

var NewS3Credentials = func(sess *session.Session, roleArn string) *credentials.Credentials {
	stsSvc := sts.New(sess)
	return stscreds.NewCredentialsWithClient(stsSvc, roleArn)
}

var GetS3Object = func(ctx context.Context, sess *session.Session, creds *credentials.Credentials, bucket, key string) (*s3.GetObjectOutput, error) {
	s3Client := s3.New(sess, &aws.Config{Credentials: creds})
	getObjInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	return s3Client.GetObjectWithContext(ctx, getObjInput)
}

var ValidateS3Config = func(cfg config.AWSConfig) error {
	if cfg.Region == "" || cfg.S3Config.Bucket == "" || cfg.S3Config.RoleArn == "" {
		return fmt.Errorf("AWS configuration is incomplete for S3 download")
	}
	return nil
}
