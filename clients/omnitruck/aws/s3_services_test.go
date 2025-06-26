package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chef/omnitruck-service/config"
)

func TestValidateS3Config(t *testing.T) {
	cfg := config.AWSConfig{
		Region: "us-east-1",
		S3Config: config.S3Config{
			Bucket:  "bucket",
			RoleArn: "role",
		},
	}
	if err := ValidateS3Config(cfg); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	cfg.Region = ""
	if err := ValidateS3Config(cfg); err == nil {
		t.Error("expected error for missing region")
	}

	cfg.Region = "us-east-1"
	cfg.S3Config.Bucket = ""
	if err := ValidateS3Config(cfg); err == nil {
		t.Error("expected error for missing bucket")
	}

	cfg.S3Config.Bucket = "bucket"
	cfg.S3Config.RoleArn = ""
	if err := ValidateS3Config(cfg); err == nil {
		t.Error("expected error for missing role arn")
	}
}

func TestNewS3Session(t *testing.T) {
	sess, err := NewS3Session("us-east-1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if sess == nil {
		t.Error("expected session, got nil")
	}
}

func TestNewS3Credentials(t *testing.T) {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	creds := NewS3Credentials(sess, "arn:aws:iam::123456789012:role/test-role")
	if creds == nil {
		t.Error("expected credentials, got nil")
	}
}

func TestGetS3Object_Error(t *testing.T) {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	creds := credentials.NewStaticCredentials("fake", "fake", "")
	_, err := GetS3Object(context.Background(), sess, creds, "fake-bucket", "fake-key")
	if err == nil {
		t.Error("expected error for fake bucket/key, got nil")
	}
}
