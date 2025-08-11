package dbconnection

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/utils/awsutils"
)

type DbConnection interface {
	GetDbConnection() *dynamodb.Client
}

type DbConectionService struct {
	AwsUtil awsutils.AwsUtils
	Config  config.ServiceConfig
	svc     *dynamodb.Client
}

func NewDbConnectionService(awsutils awsutils.AwsUtils, config config.ServiceConfig) *DbConectionService {
	return &DbConectionService{
		AwsUtil: awsutils,
		Config:  config,
	}
}

func (dbc *DbConectionService) GetDbConnection() *dynamodb.Client {
	if dbc.svc == nil {
		// TODO: Implement GetNewConfig in AwsUtils for aws-sdk-go-v2
		cfg, err := dbc.AwsUtil.GetNewConfig(context.TODO(), dbc.Config.AWSConfig)
		if err != nil {
			log.Printf("Error while creating AWS config: %v", err)
			return nil
		}
		dbc.svc = dynamodb.NewFromConfig(cfg)
	}
	return dbc.svc
}
