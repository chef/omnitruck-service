package dboperations

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	log "github.com/sirupsen/logrus"

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
	GetPackages(partitionValue string, sortValue string) (interface{}, error)
	GetVersionAll(partitionValue string) ([]string, error)
	GetMetaData(partitionValue string, sortValue string, platform string, platformVersion string, architecture string) (interface{}, error)
	GetVersionLatest(partitionValue string) (string, error)
	GetRelatedProducts(partitionValue string) (*models.RelatedProducts, error)
	GetPackageManagers() ([]string, error)
	SetDbInfo(tableName string, dbModel reflect.Type)
}

type IDynamoDBOps interface {
	GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error)
}

type DbOperationsService struct {
	db                         IDynamoDBOps
	productTableName           string
	skuTableName               string
	packageManagersTable       string
	packageDetailsCurrentTable string
	packageDetailsStableTable  string
	dbModelType                reflect.Type
}

// Expected structure of each item in the DynamoDB "package-manager-dev" table:
//
//	{
//	    "packages": "deb"
//	}
//
// - Each item represents a single package manager.
// - The "packages" attribute is a string containing the name of the package manager (e.g., "deb", "rpm", "msi").
// - One package per item is stored.
type PackageManagerItem struct {
	Packages string `json:"packages"`
}

func NewDbOperationsService(dbConnection dbconnection.DbConnection, config config.ServiceConfig) *DbOperationsService {
	return &DbOperationsService{
		db:                         dbConnection.GetDbConnection(),
		productTableName:           config.MetadataDetailsTable,
		skuTableName:               config.RelatedProductsTable,
		packageManagersTable:       config.PackageManagersTable,
		packageDetailsCurrentTable: config.PackageDetailsCurrentTable,
		packageDetailsStableTable:  config.PackageDetailsStableTable,
		dbModelType:                nil, // This will be set later using SetDbInfo
	}
}

func (dbo *DbOperationsService) SetDbInfo(tableName string, dbModelType reflect.Type) {
	dbo.productTableName = tableName
	dbo.dbModelType = dbModelType
}

func (dbo *DbOperationsService) GetPackages(partitionValue string, sortValue string) (interface{}, error) {
	res, err := dbo.fetchDataValuesWithSortKey(partitionValue, sortValue)
	if err != nil {
		return nil, err
	}
	response := reflect.New(dbo.dbModelType).Interface()
	fmt.Print("Response Item: ", res.Item)
	if err := dynamodbattribute.UnmarshalMap(res.Item, &response); err != nil {
		return nil, err
	}
	if v, ok := response.(*models.ProductDetails); ok {
		return v, nil
	}
	if v, ok := response.(*models.PackageDetails); ok {
		return v, nil
	}
	return nil, nil
}

func (dbo *DbOperationsService) GetVersionAll(partitionValue string) ([]string, error) {
	res, err := dbo.fetchDataValues(partitionValue, dbo.productTableName, constants.PRODUCT_PARTITION_KEY)
	if err != nil {
		log.Errorf("error in getting the Database value: %v", err)
		return nil, err
	}
	versionsArray := []string{}
	for _, i := range res.Items {
		model := reflect.New(dbo.dbModelType).Interface()
		err = dynamodbattribute.UnmarshalMap(i, &model)
		if err != nil {
			log.Errorf("Got error unmarshalling: %s", err)
			return nil, err
		}
		if v, ok := model.(*models.ProductDetails); ok {
			versionsArray = append(versionsArray, v.Version)
		}
		if v, ok := model.(*models.PackageDetails); ok {
			versionsArray = append(versionsArray, v.Version)
		}
	}
	return versionsArray, nil
}

func (dbo *DbOperationsService) GetMetaData(partitionValue string, sortValue string, platform string, platformVersion string, architecture string) (interface{}, error) {
	res, err := dbo.fetchDataValuesWithSortKey(partitionValue, sortValue)
	if err != nil {
		return nil, err
	}
	productDetails := reflect.New(dbo.dbModelType).Interface()
	if err := dynamodbattribute.UnmarshalMap(res.Item, &productDetails); err != nil {
		return nil, err
	}
	if v, ok := productDetails.(*models.ProductDetails); ok {
		MetaData := v.MetaData
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
	return nil, nil
}

func (dbo *DbOperationsService) GetVersionLatest(partitionValue string) (string, error) {
	versions, err := dbo.GetVersionAll(partitionValue)
	if err != nil {
		log.Errorf("Error in getting versions list: %v", err)
		return "", err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versions)))
	sortValue := versions[0]
	latestVersionDetails, err := dbo.GetPackages(partitionValue, sortValue)
	if err != nil {
		log.Errorf("Error in fetching the latest version: %v", err)
		return "", err
	}
	version := ""
	if v, ok := latestVersionDetails.(*models.ProductDetails); ok {
		version = v.Version
	}
	if v, ok := latestVersionDetails.(*models.PackageDetails); ok {
		version = v.Version
	}
	return version, nil
}

func (dbo *DbOperationsService) GetRelatedProducts(partitionValue string) (*models.RelatedProducts, error) {
	var sku models.RelatedProducts
	res, err := dbo.fetchDataValues(partitionValue, dbo.skuTableName, constants.SKU_PARTITION_KEY)
	if err != nil {
		log.Errorf("error in fetching the database values: %v", err)
		return nil, err
	}

	length := len(res.Items)
	if length == 0 {
		//TODO fix all db operation logging
		//need to add error msg logging
		//errors.New("cannot find the specific sku inside the database")
		return &models.RelatedProducts{}, nil
	}

	skuErr := dynamodbattribute.Unmarshal(res.Items[0]["bom"], &sku.Bom)
	if skuErr != nil {
		log.Errorf("Error in unmarshalling the sku name: %v", skuErr)
		return nil, skuErr
	}
	productErr := dynamodbattribute.Unmarshal(res.Items[0]["products"], &sku.Products)
	if productErr != nil {
		log.Errorf("Error in unmarshalling the map of products: %v", skuErr)
		return nil, productErr
	}
	return &sku, nil
}

func (dbo *DbOperationsService) fetchDataValues(partitionValue string, tableName string, partitionKey string) (*dynamodb.ScanOutput, error) {
	filter := expression.Name(partitionKey).Equal(expression.Value(partitionValue))

	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		log.Printf("error while building filter for this request: %v", err)
	}
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(tableName),
	}
	res, err := dbo.db.Scan(params)
	if err != nil {
		log.Errorf("error while using getting the dataBase values: %v", err)
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
		log.Errorf("error while using getting the dataBase values: %v", err)
		return nil, err
	}
	return res, nil
}

func (dbo *DbOperationsService) GetPackageManagers() ([]string, error) {
	tableName := dbo.packageManagersTable

	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	res, err := dbo.db.Scan(input)
	if err != nil {
		log.Errorf("Error scanning table %s: %v", tableName, err)
		return nil, err
	}
	if res == nil || res.Items == nil {
		log.Errorf("Scan result is nil or missing Items from table %s", tableName)
		return nil, errors.New("scan returned no items")
	}

	var results []string
	for _, item := range res.Items {
		var pkgItem PackageManagerItem
		if err := dynamodbattribute.UnmarshalMap(item, &pkgItem); err != nil {
			log.Errorf("Unmarshal error: %v", err)
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}
		if pkgItem.Packages != "" {
			results = append(results, pkgItem.Packages)
		}
	}

	return results, nil
}
