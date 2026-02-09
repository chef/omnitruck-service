package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	omnitruckConfig "github.com/chef/omnitruck-service/config"
)

// NewS3Session creates a new AWS session using aws-sdk-go-v2
var NewS3Session = func(region string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	return cfg, err
}

// NewS3Credentials returns credentials using STS AssumeRole with aws-sdk-go-v2
var NewS3Credentials = func(cfg aws.Config, roleArn string) aws.CredentialsProvider {
	stsClient := sts.NewFromConfig(cfg)
	return stscreds.NewAssumeRoleProvider(stsClient, roleArn)
}

// GetS3Object fetches an object from S3 using aws-sdk-go-v2
var GetS3Object = func(ctx context.Context, cfg aws.Config, creds aws.CredentialsProvider, bucket, key string) (*s3.GetObjectOutput, error) {
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Credentials = creds
	})
	getObjInput := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	return s3Client.GetObject(ctx, getObjInput)
}

// GetS3PresignedURL generates a presigned URL for S3 object download
var GetS3PresignedURL = func(ctx context.Context, cfg aws.Config, creds aws.CredentialsProvider, bucket, key string, expirationMinutes int) (string, error) {
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.Credentials = creds
	})
	presignClient := s3.NewPresignClient(s3Client)
	
	getObjInput := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	
	presignedReq, err := presignClient.PresignGetObject(ctx, getObjInput, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expirationMinutes) * time.Minute
	})
	if err != nil {
		return "", err
	}
	
	return presignedReq.URL, nil
}

var ValidateS3Config = func(cfg omnitruckConfig.AWSConfig) error {
	if cfg.Region == "" || cfg.S3Config.Bucket == "" || cfg.S3Config.RoleArn == "" {
		return fmt.Errorf("AWS configuration is incomplete for S3 download")
	}
	return nil
}
