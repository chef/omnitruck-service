package omnitruck

import (
	"reflect"

	"github.com/chef/omnitruck-service/models"
)

type MockDynamoServices struct {
	ProductMetadataFunc      func(params *RequestParams) (PackageMetadata, error)
	ProductPackagesFunc      func(params *RequestParams) (PackageList, error)
	GetFilenameFunc          func(params *RequestParams) (string, error)
	GetRelatedProductsFunc   func(params *RequestParams) (*models.RelatedProducts, error)
	GetPackageManagersFunc   func() ([]string, error)
	VersionLatestFunc        func(params *RequestParams) (ProductVersion, error)
	VersionAllFunc           func(params *RequestParams) ([]ProductVersion, error)
	ProductDownloadFunc      func(params *RequestParams) (string, error)
	FetchLatestOsVersionFunc func(params *RequestParams) (string, error)
	ProductsFunc             func(products []string, eol string) []string
	PlatformsFunc            func(platforms PlatformList) PlatformList

	SetDbInfoCalledWith []struct {
		Table string
		Model reflect.Type
	}
}

func (m *MockDynamoServices) SetDbInfo(table string, model reflect.Type) {
	m.SetDbInfoCalledWith = append(m.SetDbInfoCalledWith, struct {
		Table string
		Model reflect.Type
	}{
		Table: table,
		Model: model,
	})
}

func (m *MockDynamoServices) ProductMetadata(params *RequestParams) (PackageMetadata, error) {
	if m.ProductMetadataFunc != nil {
		return m.ProductMetadataFunc(params)
	}
	return PackageMetadata{}, nil
}

func (m *MockDynamoServices) ProductPackages(params *RequestParams) (PackageList, error) {
	if m.ProductPackagesFunc != nil {
		return m.ProductPackagesFunc(params)
	}
	return PackageList{}, nil
}

func (m *MockDynamoServices) GetFilename(params *RequestParams) (string, error) {
	if m.GetFilenameFunc != nil {
		return m.GetFilenameFunc(params)
	}
	return "", nil
}

func (m *MockDynamoServices) GetRelatedProducts(params *RequestParams) (*models.RelatedProducts, error) {
	if m.GetRelatedProductsFunc != nil {
		return m.GetRelatedProductsFunc(params)
	}
	return nil, nil
}

func (m *MockDynamoServices) GetPackageManagers() ([]string, error) {
	if m.GetPackageManagersFunc != nil {
		return m.GetPackageManagersFunc()
	}
	return nil, nil
}

func (m *MockDynamoServices) VersionLatest(params *RequestParams) (ProductVersion, error) {
	if m.VersionLatestFunc != nil {
		return m.VersionLatestFunc(params)
	}
	return "", nil
}

func (m *MockDynamoServices) VersionAll(params *RequestParams) ([]ProductVersion, error) {
	if m.VersionAllFunc != nil {
		return m.VersionAllFunc(params)
	}
	return nil, nil
}

func (m *MockDynamoServices) ProductDownload(params *RequestParams) (string, error) {
	if m.ProductDownloadFunc != nil {
		return m.ProductDownloadFunc(params)
	}
	return "", nil
}

func (m *MockDynamoServices) FetchLatestOsVersion(params *RequestParams) (string, error) {
	if m.FetchLatestOsVersionFunc != nil {
		return m.FetchLatestOsVersionFunc(params)
	}
	return "", nil
}

func (m *MockDynamoServices) Products(products []string, eol string) []string {
	if m.ProductsFunc != nil {
		return m.ProductsFunc(products, eol)
	}
	return products
}

func (m *MockDynamoServices) Platforms(platforms PlatformList) PlatformList {
	if m.PlatformsFunc != nil {
		return m.PlatformsFunc(platforms)
	}
	return platforms
}
