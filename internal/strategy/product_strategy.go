package strategy

import (
	"io"
	"net/http"
	"reflect"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/models"
	log "github.com/sirupsen/logrus"
)

type ProductStrategy interface {
	GetLatestVersion(params *omnitruck.RequestParams) (omnitruck.ProductVersion, *clients.Request)
	GetAllVersions(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, *clients.Request)
	GetPackages(params *omnitruck.RequestParams) (omnitruck.PackageList, error)
	GetMetadata(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, *clients.Request)
	Download(params *omnitruck.RequestParams) (url string, resp io.ReadCloser, headers http.Header, msg string, code int, err error)
	GetFileName(params *omnitruck.RequestParams) (string, error)
	UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, baseUrl string)
}

type ProductStrategyDeps struct {
	DynamoService     omnitruck.IDynamoServices
	PlatformService   omnitruck.IPlatformServices
	OmnitruckService  *omnitruck.Omnitruck
	Log               *log.Entry
	Replicated        replicated.IReplicated
	LicenseClient     clients.ILicense
	LicenseServiceUrl string
	Mode              constants.ApiType
	Config            config.ServiceConfig
	Locals            map[string]interface{}
}

// SelectProductStrategy returns the appropriate ProductStrategy based on the product.
func SelectProductStrategy(product string, channel string, deps *ProductStrategyDeps) ProductStrategy {
	switch product {
	case constants.AUTOMATE_PRODUCT, constants.HABITAT_PRODUCT:
		deps.DynamoService.SetDbInfo(deps.Config.MetadataDetailsTable, reflect.TypeOf(models.ProductDetails{}))
		return &ProductDynamoStrategy{DynamoService: deps.DynamoService, Log: deps.Log}
	case constants.PLATFORM_SERVICE_PRODUCT:
		return &PlatformServiceStrategy{
			PlatformService:   deps.PlatformService,
			Log:               deps.Log,
			Replicated:        deps.Replicated,
			LicenseClient:     deps.LicenseClient,
			LicenseServiceUrl: deps.LicenseServiceUrl,
			Mode:              deps.Mode,
			Locals:            deps.Locals,
		}
	case constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT, constants.MIGRATION_TOOL:
		if channel == constants.CURRENT_CHANNEL {
			deps.DynamoService.SetDbInfo(deps.Config.PackageDetailsCurrentTable, reflect.TypeOf(models.PackageDetails{}))
		} else {
			deps.DynamoService.SetDbInfo(deps.Config.PackageDetailsStableTable, reflect.TypeOf(models.PackageDetails{}))
		}
		return &InfraProductStrategy{
			DynamoService: deps.DynamoService,
			Log:           deps.Log,
			AWSConfig:     deps.Config.AWSConfig,
		}
	default:
		return &DefaultProductStrategy{OmnitruckService: deps.OmnitruckService}
	}
}
