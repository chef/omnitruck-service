package awsutils

import (
	"context"
	"encoding/base64"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	omnitruckConfig "github.com/chef/omnitruck-service/config"
)

type AwsUtilsImpl struct{}
type AwsUtils interface {
	GetNewConfig(ctx context.Context, awsConfig omnitruckConfig.AWSConfig) (aws.Config, error)
}

// GetNewConfig returns an aws.Config for aws-sdk-go-v2
func (au *AwsUtilsImpl) GetNewConfig(ctx context.Context, awsConfig omnitruckConfig.AWSConfig) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx,
		config.WithRegion(awsConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsConfig.AccessKey, awsConfig.SecretKey, "")),
	)
}

func NewAwsUtils() *AwsUtilsImpl {
	return &AwsUtilsImpl{}
}

var GetSecret = func(secretKey, region string) (secret string) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		// Handle session creation error
		log.Println(err.Error())
		return
	}
	svc := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretKey,
	}
	result, err := svc.GetSecretValue(ctx, input)
	if err != nil {
		log.Println("Error getting secret value:", err)
		return
	}
	if result.SecretString != nil {
		secret = *result.SecretString
	} else if result.SecretBinary != nil {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		_, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			log.Println("Base64 Decode Error:", err)
			return
		}
		secret = string(decodedBinarySecretBytes)
	}
	return
}
