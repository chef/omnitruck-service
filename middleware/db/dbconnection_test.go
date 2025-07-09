package dbconnection

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/utils/awsutils"
	"github.com/stretchr/testify/assert"
)

func TestGetDbConnection_Success(t *testing.T) {
	svc = nil

	mockSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	dbc := NewDbConnectionService(&awsutils.MockAwsUtils{
		GetNewSessionfunc: func(config config.AWSConfig) (*session.Session, error) {
			return mockSession, nil
		},
	}, config.ServiceConfig{})

	conn := dbc.GetDbConnection()
	assert.NotNil(t, conn)
}

func TestGetDbConnection_ErrorCase(t *testing.T) {
	svc = nil

	dbc := NewDbConnectionService(&awsutils.MockAwsUtils{
		GetNewSessionfunc: func(config config.AWSConfig) (*session.Session, error) {
			return nil, errors.New("simulated error")
		},
	}, config.ServiceConfig{})

	conn := dbc.GetDbConnection()
	assert.Nil(t, conn)
}
