package awsutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/chef/omnitruck-service/config"
)

type MockAwsUtils struct {
	GetNewConfigFunc func(ctx context.Context, awsConfig config.AWSConfig) (aws.Config, error)
}

func (mau *MockAwsUtils) GetNewConfig(ctx context.Context, awsConfig config.AWSConfig) (aws.Config, error) {
	return mau.GetNewConfigFunc(ctx, awsConfig)
}
