package dbconnection

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/utils/awsutils"
	"github.com/stretchr/testify/assert"
)

func TestGetDbConnection_Success(t *testing.T) {
	dbc := NewDbConnectionService(&awsutils.MockAwsUtils{
		GetNewConfigFunc: func(ctx context.Context, awsConfig config.AWSConfig) (aws.Config, error) {
			return aws.Config{Region: "us-east-1"}, nil
		},
	}, config.ServiceConfig{})

	conn := dbc.GetDbConnection()
	assert.NotNil(t, conn)
}

func TestGetDbConnection_ErrorCase(t *testing.T) {
	dbc := NewDbConnectionService(&awsutils.MockAwsUtils{
		GetNewConfigFunc: func(ctx context.Context, awsConfig config.AWSConfig) (aws.Config, error) {
			return aws.Config{}, errors.New("simulated error")
		},
	}, config.ServiceConfig{})

	conn := dbc.GetDbConnection()
	assert.Nil(t, conn)
}
