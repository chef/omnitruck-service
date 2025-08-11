package dbconnection

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type MockDbConnectionService struct {
	GetDbConnectionfunc func() *dynamodb.Client
}

func (mdbc *MockDbConnectionService) GetDbConnection() *dynamodb.Client {
	return mdbc.GetDbConnectionfunc()
}
