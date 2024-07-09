package awsutils

import (
	"testing"

	"github.com/chef/omnitruck-service/config"
	"github.com/progress-platform-services/platform-common/plogger"
	"github.com/stretchr/testify/assert"
)

func TestCreateAWSSession(t *testing.T) {
	config := config.AWSConfig{
		AccessKey: "your_access_key",
		SecretKey: "your_secret_key",
		Region:    "your_region",
	}

	plog, _ := plogger.NewLogger(plogger.LoggerConfig{
		LogToStdout: true,
		LogLevel:    "DEBUG",
	})
	
	dbc := NewAwsUtils(plog)
	sess, err := dbc.GetNewSession(config)

	assert.NoError(t, err)
	assert.NotNil(t, sess)

}
