package awsutils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type MockAWSClient struct {
	NewSessionWithOptionsfunc func(opts session.Options) (*session.Session, error)
	NewSessionfunc            func(cfgs ...*aws.Config) (*session.Session, error)
	GetSecretValuefunc        func(*secretsmanager.SecretsManager, *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
}

func (mac *MockAWSClient) NewSessionWithOptions(opts session.Options) (*session.Session, error) {
	return mac.NewSessionWithOptionsfunc(opts)
}

func (mac *MockAWSClient) NewSession(cfgs ...*aws.Config) (*session.Session, error) {
	return mac.NewSessionfunc(cfgs...)
}

func (mac *MockAWSClient) GetSecretValue(svc *secretsmanager.SecretsManager, input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	return mac.GetSecretValuefunc(svc, input)
}
