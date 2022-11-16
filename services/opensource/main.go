package opensource

import (
	"github.com/chef/omnitruck-service/clients/omnitruck"
	_ "github.com/chef/omnitruck-service/docs/opensource"
	"github.com/chef/omnitruck-service/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

type OpensourceService struct {
	services.ApiService
}

// @title			Licensed Omnitruck API for opensource products
// @version			1.0
// @description 	Licensed Omnitruck API for opensource products
// @license.name	Apache 2.0
// @license.url 	http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:3000
func NewServer(c services.Config) *OpensourceService {
	service := OpensourceService{}
	service.Initialize(c)

	channel := omnitruck.ContainsValidator{
		Field:      "Channel",
		Values:     []string{"stable"},
		Code:       400,
		AllowEmpty: true,
	}
	service.Validator.Add(&channel)

	service.buildRouter()
	return &service
}

func (server *OpensourceService) buildRouter() {
	server.App.Get("/swagger/*", swagger.New(swagger.Config{
		InstanceName: "Opensource",
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
// @Param eol			query 	bool 	false 	"EOL Products"
// @Success 200 {object} omnitruck.ProductVersion
// @Failure 400 {object} services.ErrorResponse
// @Failure 403 {object} services.ErrorResponse
// @Router /{channel}/{product}/versions/latest [get]
func (server *OpensourceService) latestVersionHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel: c.Params("channel"),
		Product: c.Params("product"),
	}
	err, ok := server.ValidateRequest(params, c)
	if err != nil || !ok {
		return err
	}

	var data []omnitruck.ProductVersion
	// Need to fetch all versions and filter out to only show the OS versions
	request := server.Omnitruck.ProductVersions(params).ParseData(&data)

	data = omnitruck.FilterList(data, func(v omnitruck.ProductVersion) bool {
		return !omnitruck.OsProductVersion(params.Product, v)
	})

	// Return the last opensource version
	// This assumes the versions are returned in ascending order
	latest_os_version := data[len(data)-1]

	if request.Ok {
		return server.SendResponse(c, &latest_os_version)
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
func (server *OpensourceService) productVersionsHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel: c.Params("channel"),
		Product: c.Params("product"),
	}
	err, ok := server.ValidateRequest(params, c)
	if err != nil || !ok {
		return err
	}

	var data []omnitruck.ProductVersion
	request := server.Omnitruck.ProductVersions(params).ParseData(&data)

	data = omnitruck.FilterList(data, func(v omnitruck.ProductVersion) bool {
		return !omnitruck.OsProductVersion(params.Product, v)
	})

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
func (server *OpensourceService) productPackagesHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel: c.Params("channel"),
		Product: c.Params("product"),
		Version: c.Query("v"),
	}
	err, ok := server.ValidateRequest(params, c)
	if err != nil || !ok {
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
func (server *OpensourceService) productMetadataHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel:         c.Params("channel"),
		Product:         c.Params("product"),
		Version:         c.Query("v"),
		Platform:        c.Query("p"),
		PlatformVersion: c.Query("pv"),
		Architecture:    c.Query("m"),
	}
	err, ok := server.ValidateRequest(params, c)
	if err != nil || !ok {
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
func (server *OpensourceService) productDownloadHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel:         c.Params("channel"),
		Product:         c.Params("product"),
		Version:         c.Query("v"),
		Platform:        c.Query("p"),
		PlatformVersion: c.Query("pv"),
		Architecture:    c.Query("m"),
	}
	err, ok := server.ValidateRequest(params, c)
	if err != nil || !ok {
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
