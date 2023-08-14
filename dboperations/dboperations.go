package dboperations

import (
	"log"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/chef/omnitruck-service/middleware/db"
	"github.com/chef/omnitruck-service/models"
)

const (
	SKU_PARTITION_KEY     = "sku"
	PRODUCT_PARTITION_KEY = "product"
	PRODUCT_SORT_KEY      = "version"
)

type IDbOperations interface {
	GetPackages(partitionValue string, sortValue string) (models.ProductDetails, error)
	GetVersionAll(partitionValue string) ([]string, error)
	GetMetaData(partitionValue string, sortValue string, platform string, platformVersion string, architecture string) (models.ProductDetails, error)
	GetVersionLatest(partitionValue string) (string, error)
	GetRelatedProducts(partitionValue string) (models.Sku, error)
}

type IDynamoDBOps interface {
	GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error)
}

type DbOperationsService struct {
	db               IDynamoDBOps
	productTableName string
	skuTableName     string
}

func NewDbOperationsService(dbConnection dbconnection.DbConnection) *DbOperationsService {
	return &DbOperationsService{
		db:               dbConnection.GetDbConnection(),
		productTableName: os.Getenv("PRODUCT_TABLE_NAME"),
		skuTableName:     os.Getenv("SKU_TABLE_NAME"),
	}
}

func (dbo *DbOperationsService) GetPackages(partitionValue string, sortValue string) (models.ProductDetails, error) {
	log.Println(dbo.productTableName)
	input := &dynamodb.GetItemInput{
		TableName: aws.String(dbo.productTableName),
		Key: map[string]*dynamodb.AttributeValue{
			PRODUCT_PARTITION_KEY: {S: aws.String(partitionValue)},
			PRODUCT_SORT_KEY:      {S: aws.String(sortValue)},
		},
	}
	res, err := dbo.db.GetItem(input)
	if err != nil {
		log.Println("error while using GetItem:", err)
		return models.ProductDetails{}, err
	}
	var response models.ProductDetails
	if err := dynamodbattribute.UnmarshalMap(res.Item, &response); err != nil {
		return models.ProductDetails{}, err
	}
	return response, nil
}

func (dbo *DbOperationsService) GetVersionAll(partitionValue string) ([]string, error) {
	res, err := dbo.fetchDataValues(partitionValue, dbo.productTableName)
	if err != nil {
		log.Printf("error in getting the Database value: %v", err)
		return nil, err
	}
	var response models.ProductDetails
	versionsArray := []string{}
	for _, i := range res.Items {
		err = dynamodbattribute.UnmarshalMap(i, &response)
		versionsArray = append(versionsArray, response.Version)
		if err != nil {
			log.Printf("Got error unmarshalling: %s", err)
			return nil, err
		}

	}
	return versionsArray, nil
}

func (dbo *DbOperationsService) GetMetaData(partitionValue string, sortValue string, platform string, platformVersion string, architecture string) (models.ProductDetails, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(dbo.productTableName),
		Key: map[string]*dynamodb.AttributeValue{
			PRODUCT_PARTITION_KEY: {S: aws.String(partitionValue)},
			PRODUCT_SORT_KEY:      {S: aws.String(sortValue)},
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

func (dbo *DbOperationsService) GetVersionLatest(partitionValue string) (string, error) {
	versions, err := dbo.GetVersionAll(partitionValue)
	if err != nil {
		log.Printf("Error in getting versions list: %v", err)
		return "", err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	sortValue := versions[0]
	latestVersionDetails, err := dbo.GetPackages(partitionValue, sortValue)
	if err != nil {
		log.Printf("Error in fetching the latest version: %v", err)
		return "", err
	}
	return latestVersionDetails.Version, nil
}

func (dbo *DbOperationsService) GetRelatedProducts(partitionValue string) (models.Sku, error) {
	res, err := dbo.fetchDataValues(partitionValue, dbo.skuTableName)
	if err != nil {
		log.Printf("error in fetching the database values: %v", err)
		return models.Sku{}, err
	}
	var sku models.Sku
	var responseArray []string
	for _, i := range res.Items {
		err = dynamodbattribute.UnmarshalMap(i, &sku)
		responseArray = append(responseArray, sku.Products...)
		if err != nil {
			log.Printf("Got error unmarshalling: %s", err)
			return models.Sku{}, err
		}
	}
	sku.Sku = partitionValue
	sku.Products = responseArray

	return sku, nil
}

func (dbo *DbOperationsService) fetchDataValues(partitionValue string, tableName string) (*dynamodb.ScanOutput, error) {
	filt := expression.Name(SKU_PARTITION_KEY).Equal(expression.Value(partitionValue))

	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		log.Printf("Got error building expression: %v", err)
	}
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(tableName),
	}
	res, err := dbo.db.Scan(params)
	if err != nil {
		log.Printf("Query API call failed: %v", err)
		return nil, err
	}
	return res, nil
}
