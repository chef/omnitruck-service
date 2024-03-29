package awsutils

import (
	"testing"

	"github.com/chef/omnitruck-service/config"
	"github.com/stretchr/testify/assert"
)

func TestCreateAWSSession(t *testing.T) {
	config := config.AWSConfig{
		AccessKey: "your_access_key",
		SecretKey: "your_secret_key",
		Region:    "your_region",
	}

	dbc := NewAwsUtils()
	sess, err := dbc.GetNewSession(config)

	assert.NoError(t, err)
	assert.NotNil(t, sess)

}
