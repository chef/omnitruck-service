package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/credentials"
	omnitruckConfig "github.com/chef/omnitruck-service/config"
)

func TestValidateS3Config(t *testing.T) {
	cfg := omnitruckConfig.AWSConfig{
		Region: "us-east-1",
		S3Config: omnitruckConfig.S3Config{
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
	cfg, err := NewS3Session("us-east-1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// aws.Config is a struct, not a pointer, so just check region
	if cfg.Region == "" {
		t.Error("expected region to be set in config")
	}
}

func TestNewS3Credentials(t *testing.T) {
	cfg, err := NewS3Session("us-east-1")
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	creds := NewS3Credentials(cfg, "arn:aws:iam::123456789012:role/test-role")
	if creds == nil {
		t.Error("expected credentials, got nil")
	}
}

func TestGetS3Object_Error(t *testing.T) {
	cfg, err := NewS3Session("us-east-1")
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	creds := credentials.NewStaticCredentialsProvider("fake", "fake", "")
	_, err = GetS3Object(context.Background(), cfg, creds, "fake-bucket", "fake-key")
	if err == nil {
		t.Error("expected error for fake bucket/key, got nil")
	}
}
