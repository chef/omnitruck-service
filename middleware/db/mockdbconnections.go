package dbconnection

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type MockDbConnectionService struct {
	GetDbConnectionfunc func() (*dynamodb.DynamoDB )
}

func (mdbc *MockDbConnectionService) GetDbConnection() (*dynamodb.DynamoDB ) {
	return mdbc.GetDbConnectionfunc()
}
