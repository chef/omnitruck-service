package strategy

import (
	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
	"github.com/gofiber/fiber/v2"
)

type ProductStrategy interface {
	GetLatestVersion(params *omnitruck.RequestParams) (omnitruck.ProductVersion, *clients.Request)
	GetAllVersions(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, *clients.Request)
	GetPackages(params *omnitruck.RequestParams) (omnitruck.PackageList, error)
	GetMetadata(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, *clients.Request)
	Download(params *omnitruck.RequestParams, c *fiber.Ctx) error
	GetFileName(params *omnitruck.RequestParams) (string, error)
	UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, baseUrl string)
}

type ProductStrategyDeps struct {
	DynamoService    *omnitruck.DynamoServices
	PlatformService  *omnitruck.PlatformServices
	OmnitruckService *omnitruck.Omnitruck
}

// SelectProductStrategy returns the appropriate ProductStrategy based on the product.
func SelectProductStrategy(product string, deps *ProductStrategyDeps) ProductStrategy {
	switch product {
	case constants.AUTOMATE_PRODUCT, constants.HABITAT_PRODUCT:
		return &ProductDynamoStrategy{DynamoService: deps.DynamoService}
	case constants.PLATFORM_SERVICE_PRODUCT:
		return &PlatformServiceStrategy{PlatformService: deps.PlatformService}
	default:
		return &DefaultProductStrategy{OmnitruckService: deps.OmnitruckService}
	}
}
