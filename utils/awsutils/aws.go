package awsutils

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/chef/omnitruck-service/config"
	"go.uber.org/zap"
)

type AwsUtilsImpl struct {
	Log *zap.Logger
}

type AwsUtils interface {
	GetNewSession(config config.AWSConfig) (*session.Session, error)
}

func NewAwsUtils(log *zap.Logger) *AwsUtilsImpl {
	return &AwsUtilsImpl{
		Log: log,
	}
}

func (au *AwsUtilsImpl) GetNewSession(config config.AWSConfig) (*session.Session, error) {
	session, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
			Region:      aws.String(config.Region),
		},
	})
	if err != nil {
		au.Log.Error("Error while creating session: %v" + err.Error())
		return nil, err
	}
	return session, nil
}

var GetSecret = func(secretKey, region string, log *zap.Logger) (secret string) {
	sess, err := session.NewSession()
	if err != nil {
		// Handle session creation error
		log.Error("error while creating a new session: " + err.Error())
		return
	}
	svc := secretsmanager.New(sess,
		aws.NewConfig().WithRegion(region))
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretKey),
	}

	result, err := svc.GetSecretValue(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				log.Error(secretsmanager.ErrCodeDecryptionFailure + aerr.Error())

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				log.Error(secretsmanager.ErrCodeInternalServiceError + aerr.Error())

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				log.Error(secretsmanager.ErrCodeInvalidParameterException + aerr.Error())

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				log.Error(secretsmanager.ErrCodeInvalidRequestException + aerr.Error())

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				log.Error(secretsmanager.ErrCodeResourceNotFoundException + aerr.Error())
			}
		} else {
			log.Error("error while connecting to aws: " + aerr.Error())
		}
		return
	}

	var secretString string
	if result.SecretString != nil {
		secretString = *result.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		_, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			log.Error("Base64 Decode Error:" + err.Error())
			return
		}
	}
	return secretString
}
