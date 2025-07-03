package omnitruck

import (
	"reflect"

	"github.com/chef/omnitruck-service/models"
)

type IDynamoServices interface {
	VersionLatest(params *RequestParams) (ProductVersion, error)
	VersionAll(params *RequestParams) ([]ProductVersion, error)
	ProductPackages(params *RequestParams) (PackageList, error)
	ProductMetadata(params *RequestParams) (PackageMetadata, error)
	GetFilename(params *RequestParams) (string, error)
	GetRelatedProducts(params *RequestParams) (*models.RelatedProducts, error)
	GetPackageManagers() ([]string, error)
	SetDbInfo(table string, model reflect.Type)
	ProductDownload(params *RequestParams) (string, error)
	FetchLatestOsVersion(params *RequestParams) (string, error)
	Products(products []string, eol string) []string
	Platforms(platforms PlatformList) PlatformList
}
