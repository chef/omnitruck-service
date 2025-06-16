package dboperations

import (
	"errors"
	"fmt"
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
	GetPackages(partitionValue string, sortValue string) (*models.ProductDetails, error)
	GetVersionAll(partitionValue string) ([]string, error)
	GetMetaData(partitionValue string, sortValue string, platform string, platformVersion string, architecture string) (*models.MetaData, error)
	GetVersionLatest(partitionValue string) (string, error)
	GetRelatedProducts(partitionValue string) (*models.RelatedProducts, error)
	GetPackageManagers() ([]string, error)
	GetPackageManagersVersionsAll(partitionValue string, channel string) ([]string, error)
	GetPackageManagersVersionsLatest(partitionValue string, channel string) (string, error)
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
	}
}

func (dbo *DbOperationsService) GetPackages(partitionValue string, sortValue string) (*models.ProductDetails, error) {
	res, err := dbo.fetchDataValuesWithSortKey(partitionValue, sortValue)
	if err != nil {
		return nil, err
	}
	var response models.ProductDetails
	if err := dynamodbattribute.UnmarshalMap(res.Item, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (dbo *DbOperationsService) GetVersionAll(partitionValue string) ([]string, error) {
	res, err := dbo.fetchDataValues(partitionValue, dbo.productTableName, constants.PRODUCT_PARTITION_KEY)
	if err != nil {
		log.Errorf("error in getting the Database value: %v", err)
		return nil, err
	}
	var response models.ProductDetails
	versionsArray := []string{}
	for _, i := range res.Items {
		err = dynamodbattribute.UnmarshalMap(i, &response)
		versionsArray = append(versionsArray, response.Version)
		if err != nil {
			log.Errorf("Got error unmarshalling: %s", err)
			return nil, err
		}

	}
	return versionsArray, nil
}

func (dbo *DbOperationsService) GetMetaData(partitionValue string, sortValue string, platform string, platformVersion string, architecture string) (*models.MetaData, error) {
	res, err := dbo.fetchDataValuesWithSortKey(partitionValue, sortValue)
	if err != nil {
		return nil, err
	}
	var productDetails models.ProductDetails
	if err := dynamodbattribute.UnmarshalMap(res.Item, &productDetails); err != nil {
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
	return latestVersionDetails.Version, nil
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

func (dbo *DbOperationsService) GetPackageManagersVersionsAll(partitionValue string, channel string) ([]string, error) {
	tableName := dbo.packageDetailsStableTable
	if channel == "current" {
		tableName = dbo.packageDetailsCurrentTable
	}
	res, err := dbo.fetchDataValues(partitionValue, tableName, constants.PRODUCT_PARTITION_KEY)
	if err != nil {
		log.Errorf("error in getting the Database value: %v", err)
		return nil, err
	}
	var response models.PackageDetails
	versionsArray := []string{}
	for _, i := range res.Items {
		err = dynamodbattribute.UnmarshalMap(i, &response)
		versionsArray = append(versionsArray, response.Version)
		if err != nil {
			log.Errorf("Got error unmarshalling: %s", err)
			return nil, err
		}
	}
	return versionsArray, nil
}

func (dbo *DbOperationsService) GetPackageManagersVersionsLatest(partitionValue string, channel string) (string, error) {
    versions, err := dbo.GetPackageManagersVersionsAll(partitionValue, channel)
    if err != nil {
        log.Errorf("Error in getting versions list: %v", err)
        return "", err
    }
    if len(versions) == 0 {
        return "", fmt.Errorf("no versions found for product %s ", partitionValue)
    }
    sort.Sort(sort.Reverse(sort.StringSlice(versions)))
    return versions[0], nil
}
