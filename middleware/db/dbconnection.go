package dbconnection

import (
	"log"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/chef/omnitruck-service/utils/awsutils"
)

var svc *dynamodb.DynamoDB

type DbConnection interface {
	GetDbConnection() *dynamodb.DynamoDB
}

type DbConectionService struct {
	AwsUtil awsutils.AwsUtils
}

func NewDbConnectionService(awsutils awsutils.AwsUtils) *DbConectionService {
	return &DbConectionService{
		AwsUtil: awsutils,
	}
}

func (dbc *DbConectionService) GetDbConnection() *dynamodb.DynamoDB {
	if svc == nil {
		sess, err := dbc.AwsUtil.GetNewSession()
		if err != nil {
			log.Printf("Error while reading the session: %v", err)
			return nil
		}
		svc = dynamodb.New(sess)
	}
	return svc
}
