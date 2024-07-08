package dboperations

import (
	"errors"
	"sort"

	"github.com/progress-platform-services/platform-common/plogger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	dbconnection "github.com/chef/omnitruck-service/middleware/db"
	"github.com/chef/omnitruck-service/models"
)

type IDbOperations interface {
	GetPackages(partitionValue string, sortValue string) (*models.ProductDetails, error)
	GetVersionAll(partitionValue string) ([]string, error)
	GetMetaData(partitionValue string, sortValue string, platform string, platformVersion string, architecture string) (*models.MetaData, error)
	GetVersionLatest(partitionValue string) (string, error)
	GetRelatedProducts(partitionValue string) (*models.RelatedProducts, error)
}

type IDynamoDBOps interface {
	GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error)
}

type DbOperationsService struct {
	db               IDynamoDBOps
	productTableName string
	skuTableName     string
	log              plogger.ILogger
}

func NewDbOperationsService(dbConnection dbconnection.DbConnection, config config.ServiceConfig, log plogger.ILogger) *DbOperationsService {
	return &DbOperationsService{
		db:               dbConnection.GetDbConnection(),
		productTableName: config.MetadataDetailsTable,
		skuTableName:     config.RelatedProductsTable,
		log:              log,
	}
}

func (dbo *DbOperationsService) GetPackages(partitionValue string, sortValue string) (*models.ProductDetails, error) {
	res, err := dbo.fetchDataValuesWithSortKey(partitionValue, sortValue)
	if err != nil {
		dbo.log.Error("error while fetching the values using sortKey: ", err)
		return nil, err
	}
	var response models.ProductDetails
	if err := dynamodbattribute.UnmarshalMap(res.Item, &response); err != nil {
		dbo.log.Error("error while unmarshing the responseMap: ", err)
		return nil, err
	}
	return &response, nil
}

func (dbo *DbOperationsService) GetVersionAll(partitionValue string) ([]string, error) {
	res, err := dbo.fetchDataValues(partitionValue, dbo.productTableName, constants.PRODUCT_PARTITION_KEY)
	if err != nil {
		dbo.log.Error("error in getting the Database value: ", err)
		return nil, err
	}
	var response models.ProductDetails
	versionsArray := []string{}
	for _, i := range res.Items {
		err = dynamodbattribute.UnmarshalMap(i, &response)
		versionsArray = append(versionsArray, response.Version)
		if err != nil {
			dbo.log.Error("Got error unmarshalling: ", err)
			return nil, err
		}

	}
	return versionsArray, nil
}

func (dbo *DbOperationsService) GetMetaData(partitionValue string, sortValue string, platform string, platformVersion string, architecture string) (*models.MetaData, error) {
	res, err := dbo.fetchDataValuesWithSortKey(partitionValue, sortValue)
	if err != nil {
		dbo.log.Error("error while fetching the values using sortKey: ", err)
		return nil, err
	}
	var productDetails models.ProductDetails
	if err := dynamodbattribute.UnmarshalMap(res.Item, &productDetails); err != nil {
		dbo.log.Error("error while unmarshing the responseMap: ", err)
		return nil, err
	}
	MetaData := productDetails.MetaData
	var response models.MetaData
	for _, j := range MetaData {
		if j.Architecture == architecture && j.Platform == platform {
			response.Architecture = architecture
			response.Platform = platform
			response.Platform_Version = platformVersion
			response.SHA1 = j.SHA1
			response.SHA256 = j.SHA256
			response.FileName = j.FileName
		}
	}
	return &response, nil
}

func (dbo *DbOperationsService) GetVersionLatest(partitionValue string) (string, error) {
	versions, err := dbo.GetVersionAll(partitionValue)
	if err != nil {
		dbo.log.Error("Error in getting versions list: ", err)
		return "", err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	sortValue := versions[0]
	latestVersionDetails, err := dbo.GetPackages(partitionValue, sortValue)
	if err != nil {
		dbo.log.Error("Error in fetching the latest version: ", err)
		return "", err
	}
	return latestVersionDetails.Version, nil
}

func (dbo *DbOperationsService) GetRelatedProducts(partitionValue string) (*models.RelatedProducts, error) {
	var sku models.RelatedProducts
	res, err := dbo.fetchDataValues(partitionValue, dbo.skuTableName, constants.SKU_PARTITION_KEY)
	if err != nil {
		dbo.log.Error("error in fetching the database values: ", err)
		return nil, err
	}

	length := len(res.Items)
	if length == 0 {
		//TODO fix all db operation logging
		//need to add error msg logging
		dbo.log.Error("error while getting the sku information: ", errors.New("cannot find the specific sku inside the database"))
		return &models.RelatedProducts{}, nil
	}

	skuErr := dynamodbattribute.Unmarshal(res.Items[0]["bom"], &sku.Bom)
	if skuErr != nil {
		dbo.log.Error("Error in unmarshalling the sku name: ", skuErr)
		return nil, skuErr
	}
	productErr := dynamodbattribute.Unmarshal(res.Items[0]["products"], &sku.Products)
	if productErr != nil {
		dbo.log.Error("Error in unmarshalling the map of products: ", skuErr)
		return nil, productErr
	}
	return &sku, nil
}

func (dbo *DbOperationsService) fetchDataValues(partitionValue string, tableName string, partitionKey string) (*dynamodb.ScanOutput, error) {
	filter := expression.Name(partitionKey).Equal(expression.Value(partitionValue))

	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		dbo.log.Error("error while building filter for this request: ", err)
	}
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(tableName),
	}
	res, err := dbo.db.Scan(params)
	if err != nil {
		dbo.log.Error("error while using getting the dataBase values: ", err)
		return nil, err
	}
	return res, nil
}

func (dbo *DbOperationsService) fetchDataValuesWithSortKey(partitionValue string, sortValue string) (*dynamodb.GetItemOutput, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(dbo.productTableName),
		Key: map[string]*dynamodb.AttributeValue{
			constants.PRODUCT_PARTITION_KEY: {S: aws.String(partitionValue)},
			constants.PRODUCT_SORT_KEY:      {S: aws.String(sortValue)},
		},
	}
	res, err := dbo.db.GetItem(input)
	if err != nil {
		dbo.log.Error("error while using getting the dataBase values: ", err)
		return nil, err
	}
	return res, nil
}
