package dboperations

import (
	"reflect"

	"github.com/chef/omnitruck-service/models"
)

type MockIDbOperations struct {
	GetPackagesfunc        func(partitionValue string, sortValue string) (interface{}, error)
	GetVersionAllfunc      func(partitionValue string) ([]string, error)
	GetMetaDatafunc        func(partitionValue, sortValue, platform, platformVersion, architecture, packageManager string) (*models.MetaData, error)
	GetVersionLatestfunc   func(partitionValue string) (string, error)
	GetRelatedProductsfunc func(partitionValue string) (*models.RelatedProducts, error)
	GetPackageManagersfunc func() ([]string, error)
	SetDbInfofunc          func(tableName string, dbModel reflect.Type)
}

func (mdbop *MockIDbOperations) GetPackages(partitionValue string, sortValue string) (interface{}, error) {
	return mdbop.GetPackagesfunc(partitionValue, sortValue)
}

func (mdbop *MockIDbOperations) GetVersionAll(partitionValue string) ([]string, error) {
	return mdbop.GetVersionAllfunc(partitionValue)
}

func (mdbop *MockIDbOperations) GetMetaData(partitionValue, sortValue, platform, platformVersion, architecture, packageManager string) (*models.MetaData, error) {
	return mdbop.GetMetaDatafunc(partitionValue, sortValue, platform, platformVersion, architecture, packageManager)
}

func (mdbop *MockIDbOperations) GetVersionLatest(partitionValue string) (string, error) {
	return mdbop.GetVersionLatestfunc(partitionValue)
}

func (mdbop *MockIDbOperations) GetRelatedProducts(partitionValue string) (*models.RelatedProducts, error) {
	return mdbop.GetRelatedProductsfunc(partitionValue)
}

func (mdbop *MockIDbOperations) GetPackageManagers() ([]string, error) {
	return mdbop.GetPackageManagersfunc()
}

func (mdbop *MockIDbOperations) SetDbInfo(tableName string, dbModel reflect.Type) {
	mdbop.SetDbInfofunc(tableName, dbModel)
}
