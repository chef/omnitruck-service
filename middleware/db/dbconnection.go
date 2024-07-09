package dbconnection

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/utils/awsutils"
	"go.uber.org/zap"
)

var svc *dynamodb.DynamoDB

type DbConnection interface {
	GetDbConnection() *dynamodb.DynamoDB
}

type DbConectionService struct {
	AwsUtil awsutils.AwsUtils
	Config  config.ServiceConfig
	Logger  *zap.Logger
}

func NewDbConnectionService(awsutils awsutils.AwsUtils, config config.ServiceConfig, log *zap.Logger) *DbConectionService {
	return &DbConectionService{
		AwsUtil: awsutils,
		Config:  config,
		Logger:  log,
	}
}

func (dbc *DbConectionService) GetDbConnection() *dynamodb.DynamoDB {
	if svc == nil {
		sess, err := dbc.AwsUtil.GetNewSession(dbc.Config.AWSConfig)
		if err != nil {
			dbc.Logger.Error("Error while reading the session: " + err.Error())
			return nil
		}
		svc = dynamodb.New(sess)
	}
	return svc
}
