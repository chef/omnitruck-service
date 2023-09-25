package awsutils

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/chef/omnitruck-service/config"
)

type MockAwsUtils struct {
	GetNewSessionfunc func(config config.AWSConfig) (*session.Session, error)
}

func (mau *MockAwsUtils) GetNewSession(config config.AWSConfig) (*session.Session, error) {
	return mau.GetNewSessionfunc(config)
}
