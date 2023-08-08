package services

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/chef/omnitruck-service/middleware/db"
	"github.com/chef/omnitruck-service/models"
)

type IDbOperations interface {
	GetPackages(partitionKey string, partitionValue string, sortKey string, sortValue string, tableName string) (models.ProductDetails, error)
	GetVersionAll(partitionKey string, partitionValue string, tableName string) ([]string, error)
	GetMetaData(partitionKey string, partitionValue string, sortKey string, sortValue string, tableName string, platform string, platformVersion string, architecture string) (models.ProductDetails, error)
	GetVersionLatest(partitionKey string, partitionValue string, tableName string) (string, error)
	GetRelatedProducts(partitionKey string, partitionValue string, tableName string) (models.Sku, error)
}

type IDBOps interface {
	GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	Scan(*dynamodb.ScanInput) (*dynamodb.ScanOutput, error)
}

type DbOperationsService struct {
	db IDBOps
}

func NewDbOperationsService(dbConnection dbconnection.DbConnection) *DbOperationsService {
	return &DbOperationsService{
		db: dbConnection.GetDbConnection(),
	}
}

func (dbo *DbOperationsService) GetPackages(partitionKey string, partitionValue string, sortKey string, sortValue string, tableName string) (models.ProductDetails, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			partitionKey: {S: aws.String(partitionValue)},
			sortKey:      {S: aws.String(sortValue)},
		},
	}
	res, err := dbo.db.GetItem(input)
	if err != nil {
		fmt.Println("error while using GetItem:", err)
		return models.ProductDetails{}, err
	}
	var response models.ProductDetails
	if err := dynamodbattribute.UnmarshalMap(res.Item, &response); err != nil {
		return models.ProductDetails{}, err
	}
	return response, nil
}

func (dbo *DbOperationsService) GetVersionAll(partitionKey string, partitionValue string, tableName string) ([]string, error) {
	res, err := dbo.fetchDataValues(partitionKey, partitionValue, tableName)
	if err != nil {
		fmt.Printf("error in getting the Database value: %v", err)
		return nil, err
	}
	var response models.ProductDetails
	versionsArray := []string{}
	for _, i := range res.Items {
		err = dynamodbattribute.UnmarshalMap(i, &response)
		versionsArray = append(versionsArray, response.Version)
		if err != nil {
			fmt.Printf("Got error unmarshalling: %s", err)
			return nil, err
		}

	}
	return versionsArray, nil
}

func (dbo *DbOperationsService) GetMetaData(partitionKey string, partitionValue string, sortKey string, sortValue string, tableName string, platform string, platformVersion string, architecture string) (models.ProductDetails, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			partitionKey: {S: aws.String(partitionValue)},
			sortKey:      {S: aws.String(sortValue)},
		},
	}
	res, err := dbo.db.GetItem(input)
	if err != nil {
		return models.ProductDetails{}, err
	}
	var productDetails models.ProductDetails
	if err := dynamodbattribute.UnmarshalMap(res.Item, &productDetails); err != nil {
		return models.ProductDetails{}, err
	}
	MetaData := productDetails.MetaData
	var response models.MetaData
	var responseArray []models.MetaData
	for _, j := range MetaData {
		if j.Architecture == architecture && j.Platform == platform && j.Platform_Version == platformVersion {
			response.Architecture = architecture
			response.Platform = platform
			response.Platform_Version = platformVersion
			response.SHA1 = j.SHA1
			response.SHA256 = j.SHA256
		}
	}
	responseArray = append(responseArray, response)
	productDetails.MetaData = responseArray
	return productDetails, nil
}

func (dbo *DbOperationsService) GetVersionLatest(partitionKey string, partitionValue string, tableName string) (string, error) {
	versions, err := dbo.GetVersionAll(partitionKey, partitionValue, tableName)
	if err != nil {
		fmt.Printf("Error in getting versions list: %v", err)
		return "", err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	sortValue := versions[0]
	latestVersionDetails, err := dbo.GetPackages(partitionKey, partitionValue, "Version", sortValue, tableName)
	if err != nil {
		fmt.Printf("Error in fetching the latest version: %v", err)
		return "", err
	}
	return latestVersionDetails.Version, nil
}

func (dbo *DbOperationsService) GetRelatedProducts(partitionKey string, partitionValue string, tableName string) (models.Sku, error) {
	res, err := dbo.fetchDataValues(partitionKey, partitionValue, tableName)
	if err != nil {
		fmt.Printf("error in fetching the database values: %v", err)
		return models.Sku{}, err
	}
	var sku models.Sku
	var responseArray []string
	for _, i := range res.Items {
		err = dynamodbattribute.UnmarshalMap(i, &sku)
		responseArray = append(responseArray, sku.Products...)
		if err != nil {
			fmt.Printf("Got error unmarshalling: %s", err)
			return models.Sku{}, err
		}
	}
	sku.Skus = partitionValue
	sku.Products = responseArray

	return sku, nil
}

func (dbo *DbOperationsService) fetchDataValues(partitionKey string, partitionValue string, tableName string) (*dynamodb.ScanOutput, error) {
	filt := expression.Name(partitionKey).Equal(expression.Value(partitionValue))

	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		fmt.Printf("Got error building expression: %v", err)
	}
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(tableName),
	}
	res, err := dbo.db.Scan(params)
	if err != nil {
		fmt.Printf("Query API call failed: %v", err)
		return nil, err
	}
	return res, nil
}
