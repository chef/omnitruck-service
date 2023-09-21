package dbconnection

import (
	"log"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/utils/awsutils"
)

var svc *dynamodb.DynamoDB

type DbConnection interface {
	GetDbConnection() *dynamodb.DynamoDB
}

type DbConectionService struct {
	AwsUtil awsutils.AwsUtils
	Config  config.DbConfig
}

func NewDbConnectionService(awsutils awsutils.AwsUtils, config config.DbConfig) *DbConectionService {
	return &DbConectionService{
		AwsUtil: awsutils,
		Config:  config,
	}
}

func (dbc *DbConectionService) GetDbConnection() *dynamodb.DynamoDB {
	if svc == nil {
		sess, err := dbc.AwsUtil.GetNewSession(dbc.Config.AWSConfig)
		if err != nil {
			log.Printf("Error while reading the session: %v", err)
			return nil
		}
		svc = dynamodb.New(sess)
	}
	return svc
}
