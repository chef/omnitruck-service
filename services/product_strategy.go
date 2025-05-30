package services

import (
	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
	"github.com/gofiber/fiber/v2"
)

type ProductStrategy interface {
	GetLatestVersion(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.ProductVersion, *clients.Request)
	GetAllVersions(params *omnitruck.RequestParams, c *fiber.Ctx) ([]omnitruck.ProductVersion, *clients.Request)
	GetPackages(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.PackageList, error)
	GetMetadata(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.PackageMetadata, *clients.Request)
	Download(params *omnitruck.RequestParams, c *fiber.Ctx) error
	GetFileName(params *omnitruck.RequestParams, c *fiber.Ctx) (string, error)
	UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, c *fiber.Ctx)
}

// SelectProductStrategy returns the appropriate ProductStrategy based on the product.
func SelectProductStrategy(product string, server *ApiService) ProductStrategy {
	switch product {
	case constants.AUTOMATE_PRODUCT, constants.HABITAT_PRODUCT:
		return &ProductDynamoStrategy{Server: server}
	case constants.PLATFORM_SERVICE_PRODUCT:
		return &PlatformServiceStrategy{Server: server}
	default:
		return &DefaultProductStrategy{Server: server}
	}
}
