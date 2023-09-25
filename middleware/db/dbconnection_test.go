package dbconnection

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/utils/awsutils"
	"github.com/stretchr/testify/assert"
)

func TestGetDbConnection(t *testing.T) {
	mockSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	dbc := NewDbConnectionService(&awsutils.MockAwsUtils{
		GetNewSessionfunc: func(config config.AWSConfig) (*session.Session, error) {
			return mockSession, nil
		},
	}, config.ServiceConfig{})

	svc := dbc.GetDbConnection()

	assert.NotNil(t, svc)
}
