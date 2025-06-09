package awsutils

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type IAWSClient interface {
	NewSessionWithOptions(opts session.Options) (*session.Session, error)
	NewSession(cfgs ...*aws.Config) (*session.Session, error)
	GetSecretValue(svc *secretsmanager.SecretsManager, input *secretsmanager.GetSecretValueInput) (result *secretsmanager.GetSecretValueOutput, err error)
}

type AWSClient struct{}

func NewAWSClient() IAWSClient {
	return &AWSClient{}
}

func (awsc *AWSClient) NewSessionWithOptions(opts session.Options) (*session.Session, error) {
	return session.NewSessionWithOptions(opts)
}

func (awsc *AWSClient) NewSession(cfgs ...*aws.Config) (*session.Session, error) {
	return session.NewSession()
}

func (awsc *AWSClient) GetSecretValue(svc *secretsmanager.SecretsManager, input *secretsmanager.GetSecretValueInput) (result *secretsmanager.GetSecretValueOutput, err error) {
	return svc.GetSecretValue(input)
}
