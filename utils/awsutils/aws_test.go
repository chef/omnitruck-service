package awsutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAWSSession(t *testing.T) {
	os.Setenv("ACCESS_KEY", "your_access_key")
	os.Setenv("SECRET_KEY", "your_secret_key")

	dbc := NewAwsUtils()
	sess, err := dbc.GetNewSession()

	os.Unsetenv("ACCESS_KEY")
	os.Unsetenv("SECRET_KEY")
	assert.NoError(t, err)
	assert.NotNil(t, sess)

}
