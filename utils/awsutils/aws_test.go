package awsutils

import (
	"testing"

	"github.com/chef/omnitruck-service/config"
	"github.com/stretchr/testify/assert"
)

func TestGetNewConfig(t *testing.T) {
	awsCfg := config.AWSConfig{
		AccessKey: "your_access_key",
		SecretKey: "your_secret_key",
		Region:    "your_region",
	}

	utils := NewAwsUtils()
	cfg, err := utils.GetNewConfig(nil, awsCfg)

	assert.NoError(t, err)
	assert.Equal(t, "your_region", cfg.Region)
}
