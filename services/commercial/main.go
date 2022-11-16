package commercial

import (
	"github.com/chef/omnitruck-service/clients/omnitruck"
	_ "github.com/chef/omnitruck-service/docs/commercial"
	"github.com/chef/omnitruck-service/middleware/license"
	"github.com/chef/omnitruck-service/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

type CommercialService struct {
	services.ApiService
}

// @title			Licensed Commercial Omnitruck API
// @version			1.0
// @description 	Licensed Commercial Omnitruck API
// @license.name	Apache 2.0
// @license.url 	http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:3002
func NewServer(c services.Config) *CommercialService {
	service := CommercialService{}
	service.Initialize(c)

	service.App.Use(license.New(license.Config{
		Next: func(id string, c *fiber.Ctx) bool {
			// Allow empty licenses
			if len(id) == 0 {
				return true
			}
			return false
		},
	}))

	service.Log.Info("Adding EOL Validator")
	eolversion := omnitruck.EolVersionValidator{}
	service.Validator.Add(&eolversion)

	service.buildRouter()

	return &service
}

func (server *CommercialService) buildRouter() {
	server.App.Get("/swagger/*", swagger.New(swagger.Config{
		InstanceName: "Commercial",
	}))

	server.App.Get("/:channel/:product/versions/latest", server.latestVersionHandler)
	server.App.Get("/:channel/:product/versions/all", server.productVersionsHandler)
	server.App.Get("/:channel/:product/packages", server.productPackagesHandler)
	server.App.Get("/:channel/:product/metadata", server.productMetadataHandler)
	server.App.Get("/:channel/:product/download", server.productDownloadHandler)

}

// @description Get the latest version number for a particular channel and product combination.
// @Param channel 		path 	string 	true 	"Channel" Enums(current, stable)
// @Param product   	path 	string 	true 	"Product"
// @Param license_id 	header 	string 	false 	"License ID"
// @Success 200 {object} omnitruck.ProductVersion
// @Failure 400 {object} services.ErrorResponse
// @Failure 403 {object} services.ErrorResponse
// @Router /{channel}/{product}/versions/latest [get]
func (server *CommercialService) latestVersionHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel: c.Params("channel"),
		Product: c.Params("product"),
		Version: "latest",
	}
	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	var data omnitruck.ProductVersion
	request := server.Omnitruck.LatestVersion(params).ParseData(&data)

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}

}

// @description Get a list of all available version numbers for a particular channel and product combination
// @Param channel 		path 	string 	true 	"Channel" Enums(current, stable)
// @Param product   	path 	string 	true 	"Product"
// @Param license_id 	header 	string 	false 	"License ID"
// @Param eol			query 	bool 	false 	"EOL Products" Default(false)
// @Success 200 {object} omnitruck.ItemList
// @Failure 400 {object} services.ErrorResponse
// @Failure 403 {object} services.ErrorResponse
// @Router /{channel}/{product}/versions/all [get]
func (server *CommercialService) productVersionsHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel: c.Params("channel"),
		Product: c.Params("product"),
		Eol:     c.Query("eol", "false"),
	}

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	var data []omnitruck.ProductVersion
	request := server.Omnitruck.ProductVersions(params).ParseData(&data)

	if params.Eol != "true" {
		data = omnitruck.FilterProductList(data, params.Product, omnitruck.EolProductVersion)
	}

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}

}

// @description Get the full list of all packages for a particular channel and product combination.
// @description By default all packages for the latest version are returned. If the v query string parameter is included the packages for the specified version are returned.
// @Param channel 		path 	string 	true 	"Channel" Enums(current, stable)
// @Param product   	path 	string 	true 	"Product" Example(chef)
// @Param v				query	string	false	"Version"
// @Param license_id 	header 	string 	false 	"License ID"
// @Param eol			query 	bool 	false 	"EOL Products" Default(false)
// @Success 200 {object} omnitruck.PackageList
// @Failure 400 {object} services.ErrorResponse
// @Failure 403 {object} services.ErrorResponse
// @Router /{channel}/{product}/packages [get]
func (server *CommercialService) productPackagesHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel: c.Params("channel"),
		Product: c.Params("product"),
		Version: c.Query("v"),
		Eol:     c.Query("eol"),
	}

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	var data omnitruck.PackageList
	request := server.Omnitruck.ProductPackages(params).ParseData(&data)

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}

}

// @description Get details for a particular package.
// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Param channel 		path 	string 	true 	"Channel" 			Enums(current, stable)
// @Param product   	path 	string 	true 	"Product" 			Example(chef)
// @Param p				query	string	true	"Platform, valid values are returned from the `/platforms` endpoint." 				Example(ubuntu)
// @Param pv			query	string	true	"Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." 	Example(20.04)
// @Param m				query	string	true	"Machine architecture, valid values are returned by the `/architectures` endpoint."	Example(x86_64)
// @Param v				query	string	false	"Version of the product to be installed. A version always takes the form `x.y.z`"			Default(latest)
// @Param license_id 	header 	string 	false 	"License ID"
// @Param eol			query 	bool 	false 	"EOL Products" Default(false)
// @Success 200 {object} omnitruck.PackageMetadata
// @Failure 400 {object} services.ErrorResponse
// @Failure 403 {object} services.ErrorResponse
// @Router /{channel}/{product}/metadata [get]
func (server *CommercialService) productMetadataHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel:         c.Params("channel"),
		Product:         c.Params("product"),
		Version:         c.Query("v"),
		Platform:        c.Query("p"),
		PlatformVersion: c.Query("pv"),
		Architecture:    c.Query("m"),
	}

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	var data omnitruck.PackageMetadata
	request := server.Omnitruck.ProductMetadata(params).ParseData(&data)

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}

}

// @description Get details for a particular package.
// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Param channel 		path 	string 	true 	"Channel" 			Enums(current, stable)
// @Param product   	path 	string 	true 	"Product" 			Example(chef)
// @Param p				query	string	true	"Platform, valid values are returned from the `/platforms` endpoint." 				Example(ubuntu)
// @Param pv			query	string	true	"Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." 	Example(20.04)
// @Param m				query	string	true	"Machine architecture, valid values are returned by the `/architectures` endpoint."	Example(x86_64)
// @Param v				query	string	false	"Version of the product to be installed. A version always takes the form `x.y.z`"			Default(latest)
// @Param license_id 	header 	string 	false 	"License ID"
// @Param eol			query 	bool 	false 	"EOL Products" Default(false)
// @Success 302
// @Failure 400 {object} services.ErrorResponse
// @Failure 403 {object} services.ErrorResponse
// @Router /{channel}/{product}/download [get]
func (server *CommercialService) productDownloadHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel:         c.Params("channel"),
		Product:         c.Params("product"),
		Version:         c.Query("v"),
		Platform:        c.Query("p"),
		PlatformVersion: c.Query("pv"),
		Architecture:    c.Query("m"),
	}

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	var data omnitruck.PackageMetadata
	request := server.Omnitruck.ProductDownload(params).ParseData(&data)

	if request.Ok {
		server.Log.Infof("Redirecting user to %s", data.Url)
		return c.Redirect(data.Url, 302)
	} else {
		return server.SendError(c, request)
	}
}
