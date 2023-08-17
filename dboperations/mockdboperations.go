package dboperations

import (
	"github.com/chef/omnitruck-service/models"
)

type MockIDbOperations struct {
	GetPackagesfunc            func(partitionValue string, sortValue string) (*models.ProductDetails, error)
	GetVersionAllfunc          func(partitionValue string) ([]string, error)
	GetMetaDatafunc            func(partitionValue string, sortValue string, platform string, platformVersion string, architecture string) (*models.MetaData, error)
	GetMetaDataWithoutSortfunc func(partitionValue string, platform string, platformVersion string, architecture string) (*models.MetaData, error)
	GetVersionLatestfunc       func(partitionValue string) (string, error)
	GetRelatedProductsfunc     func(partitionValue string) (*models.RelatedProducts, error)
}

func (mdbop *MockIDbOperations) GetPackages(partitionValue string, sortValue string) (*models.ProductDetails, error) {
	return mdbop.GetPackagesfunc(partitionValue, sortValue)
}

func (mdbop *MockIDbOperations) GetVersionAll(partitionValue string) ([]string, error) {
	return mdbop.GetVersionAllfunc(partitionValue)
}

func (mdbop *MockIDbOperations) GetMetaData(partitionValue string, sortValue string, platform string, platformVersion string, architecture string) (*models.MetaData, error) {
	return mdbop.GetMetaDatafunc(partitionValue, sortValue, platform, platformVersion, architecture)
}

func (mdbop *MockIDbOperations) GetVersionLatest(partitionValue string) (string, error) {
	return mdbop.GetVersionLatestfunc(partitionValue)
}

func (mdbop *MockIDbOperations) GetRelatedProducts(partitionValue string) (*models.RelatedProducts, error) {
	return mdbop.GetRelatedProductsfunc(partitionValue)
}

func (mdop *MockIDbOperations) GetMetaDataWithoutSort(partitionValue string, platform string, platformVersion string, architecture string) (*models.MetaData, error) {
	return mdop.GetMetaDataWithoutSortfunc(partitionValue, platform, platformVersion, architecture)
}
