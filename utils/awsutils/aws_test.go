package awsutils

import (
	"testing"

	"github.com/chef/omnitruck-service/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCreateAWSSession(t *testing.T) {
	config := config.AWSConfig{
		AccessKey: "your_access_key",
		SecretKey: "your_secret_key",
		Region:    "your_region",
	}

	log := zap.NewNop()

	dbc := NewAwsUtils(log)
	sess, err := dbc.GetNewSession(config)

	assert.NoError(t, err)
	assert.NotNil(t, sess)

}
