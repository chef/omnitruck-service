package awsutils_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/utils/awsutils"
	"github.com/stretchr/testify/assert"
)

func TestCreateAWSSession(t *testing.T) {
	tests := []struct {
		name                          string
		config                        config.AWSConfig
		mockNewSessionWithOptionsfunc func(opts session.Options) (*session.Session, error)
		requiredError                 error
	}{
		{
			name: "successfully created the aws session",
			config: config.AWSConfig{
				AccessKey: "your_access_key",
				SecretKey: "your_secret_key",
				Region:    "your_region",
			},
			mockNewSessionWithOptionsfunc: func(opts session.Options) (*session.Session, error) {
				return &session.Session{}, nil
			},
			requiredError: nil,
		},
		{
			name: "error while creating a session with AWS Config",
			config: config.AWSConfig{
				AccessKey: "your_access_key",
				SecretKey: "your_secret_key",
			},
			mockNewSessionWithOptionsfunc: func(opts session.Options) (*session.Session, error) {
				return nil, errors.New("region not found")
			},
			requiredError: errors.New("region not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := awsutils.MockAWSClient{}
			mockSession.NewSessionWithOptionsfunc = tt.mockNewSessionWithOptionsfunc
			service := awsutils.AwsUtilsImpl{
				AWSClient: &mockSession,
			}
			resp, err := service.GetNewSession(tt.config)
			if err != nil {
				assert.Equal(t, err, tt.requiredError)
			} else {
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestGetSecret(t *testing.T) {
	secretValue := "SECRET STRING"
	type args struct {
		secretKey string
		region    string
	}
	tests := []struct {
		name               string
		config             config.AWSConfig
		arguments          args
		NewSessionfunc     func(cfgs ...*aws.Config) (*session.Session, error)
		GetSecretValuefunc func(*secretsmanager.SecretsManager, *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
		requiredError      error
		response           string
	}{
		{
			name:   "succcessfully got the secret key",
			config: config.AWSConfig{},
			arguments: args{
				secretKey: "YOUR SECRET KEY",
				region:    "YOUR REGION",
			},
			NewSessionfunc: func(cfgs ...*aws.Config) (*session.Session, error) {
				return session.New(aws.NewConfig()), nil
			},
			GetSecretValuefunc: func(sm *secretsmanager.SecretsManager, gsvi *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
				return &secretsmanager.GetSecretValueOutput{}, nil
			},
			requiredError: nil,
			response:      "",
		},
		{
			name:   "error while creating a new session",
			config: config.AWSConfig{},
			arguments: args{
				secretKey: "YOUR SECRET KEY",
				region:    "YOUR REGION",
			},
			NewSessionfunc: func(cfgs ...*aws.Config) (*session.Session, error) {
				return nil, errors.New("region not found")
			},
			response: "",
		},
		{
			name:   "DecryptionFailure error while fetching the secret value",
			config: config.AWSConfig{},
			arguments: args{
				secretKey: "YOUR SECRET KEY",
				region:    "YOUR REGION",
			},
			NewSessionfunc: func(cfgs ...*aws.Config) (*session.Session, error) {
				return session.New(aws.NewConfig()), nil
			},
			GetSecretValuefunc: func(sm *secretsmanager.SecretsManager, gsvi *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
				return nil, &secretsmanager.DecryptionFailure{}
			},
			response: "",
		},
		{
			name:   "InternalServiceError error while fetching the secret value",
			config: config.AWSConfig{},
			arguments: args{
				secretKey: "YOUR SECRET KEY",
				region:    "YOUR REGION",
			},
			NewSessionfunc: func(cfgs ...*aws.Config) (*session.Session, error) {
				return session.New(aws.NewConfig()), nil
			},
			GetSecretValuefunc: func(sm *secretsmanager.SecretsManager, gsvi *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
				return nil, &secretsmanager.InternalServiceError{}
			},
			response: "",
		},
		{
			name:   "InvalidRequestException error while fetching the secret value",
			config: config.AWSConfig{},
			arguments: args{
				secretKey: "YOUR SECRET KEY",
				region:    "YOUR REGION",
			},
			NewSessionfunc: func(cfgs ...*aws.Config) (*session.Session, error) {
				return session.New(aws.NewConfig()), nil
			},
			GetSecretValuefunc: func(sm *secretsmanager.SecretsManager, gsvi *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
				return nil, &secretsmanager.InvalidRequestException{}
			},
			response: "",
		},
		{
			name:   "InvalidParameterException error while fetching the secret value",
			config: config.AWSConfig{},
			arguments: args{
				secretKey: "YOUR SECRET KEY",
				region:    "YOUR REGION",
			},
			NewSessionfunc: func(cfgs ...*aws.Config) (*session.Session, error) {
				return session.New(aws.NewConfig()), nil
			},
			GetSecretValuefunc: func(sm *secretsmanager.SecretsManager, gsvi *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
				return nil, &secretsmanager.InvalidParameterException{}
			},
			response: "",
		},
		{
			name:   "ResourceNotFoundException error while fetching the secret value",
			config: config.AWSConfig{},
			arguments: args{
				secretKey: "YOUR SECRET KEY",
				region:    "YOUR REGION",
			},
			NewSessionfunc: func(cfgs ...*aws.Config) (*session.Session, error) {
				return session.New(aws.NewConfig()), nil
			},
			GetSecretValuefunc: func(sm *secretsmanager.SecretsManager, gsvi *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
				return nil, &secretsmanager.ResourceNotFoundException{}
			},
			response: "",
		},
		{
			name:   "succcessfully got the secret key",
			config: config.AWSConfig{},
			arguments: args{
				secretKey: "YOUR SECRET KEY",
				region:    "YOUR REGION",
			},
			NewSessionfunc: func(cfgs ...*aws.Config) (*session.Session, error) {
				return session.New(aws.NewConfig()), nil
			},
			GetSecretValuefunc: func(sm *secretsmanager.SecretsManager, gsvi *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
				return &secretsmanager.GetSecretValueOutput{
					SecretString: &secretValue,
				}, nil
			},
			requiredError: nil,
			response:      secretValue,
		},
		{
			name:   "error whilen getting the secret",
			config: config.AWSConfig{},
			arguments: args{
				secretKey: "YOUR SECRET KEY",
				region:    "YOUR REGION",
			},
			NewSessionfunc: func(cfgs ...*aws.Config) (*session.Session, error) {
				return nil, errors.New("region not found")
			},
			response: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := awsutils.MockAWSClient{}
			mockSession.NewSessionfunc = tt.NewSessionfunc
			mockSession.GetSecretValuefunc = tt.GetSecretValuefunc
			service := awsutils.AwsUtilsImpl{
				AWSClient: &mockSession,
			}
			resp := service.GetSecret(tt.arguments.secretKey, tt.arguments.region)
			assert.Equal(t, resp, tt.response)
		})
	}
}
