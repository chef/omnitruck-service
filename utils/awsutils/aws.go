package awsutils

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsConfig "github.com/chef/omnitruck-service/models"
)

type AwsUtilsImpl struct {}

type AwsUtils interface {
	GetNewSession() (*session.Session, error)
}

func NewAwsUtils() *AwsUtilsImpl {
	return &AwsUtilsImpl{}
}

func (au *AwsUtilsImpl) GetNewSession() (*session.Session, error) {
	var awsConfig awsConfig.AWSConfig
	awsConfig.AccessKey = os.Getenv("ACCESS_KEY")
	awsConfig.SecretKey = os.Getenv("SECRET_KEY")
	awsConfig.Region = os.Getenv("REGION")
	session, err :=  session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(awsConfig.AccessKey, awsConfig.SecretKey, ""),
			Region:      aws.String(awsConfig.Region),
		},
	})
	if err != nil {
		fmt.Printf("Error while creating session: %v", err)
		return nil, err
	}
	return session, nil
}
