package services

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	_ "github.com/chef/omnitruck-service/docs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/gomarkdown/markdown"
)

// @title        Licensed Omnitruck API
// @version      1.0
// @description  Licensed Omnitruck API
// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html
func (server *ApiService) buildRouter() {
	server.App.Get("/swagger/*", swagger.New(swagger.Config{
		InstanceName: "OmnitruckApi",
	}))

	// Add the endpoints that don't require any special handling for various APIs
	server.App.Static("/", "./static", fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		Index:         "index.html",
		CacheDuration: 10 * time.Second,
		MaxAge:        3600,
	})
	server.App.Get("/status", server.HealthCheck)
	server.App.Get("/products", server.productsHandler)
	server.App.Get("/platforms", server.platformsHandler)
	server.App.Get("/architectures", server.architecturesHandler)
	server.App.Get("/:channel/:product/versions/latest", server.latestVersionHandler)
	server.App.Get("/:channel/:product/versions/all", server.productVersionsHandler)
	server.App.Get("/:channel/:product/packages", server.productPackagesHandler)
	server.App.Get("/:channel/:product/metadata", server.productMetadataHandler)
	server.App.Get("/:channel/:product/download", server.productDownloadHandler)

}

func (server *ApiService) docsHandler(baseUrl string) func(*fiber.Ctx) error {
	content, err := os.ReadFile("docs/index.md")
	if err != nil {
		content = []byte("Error reading docs/index.md")
	}
	output := markdown.ToHTML(content, nil, nil)
	view_data := fiber.Map{
		"Content": string(output),
		"baseUrl": baseUrl,
	}

	return func(c *fiber.Ctx) error {
		return c.Render("docs", view_data, "layouts/docs")
	}
}

// @description Returns a valid list of valid product keys.
// @description Any of these product keys can be used in the <PRODUCT> value of other endpoints. Please note many of these products are used for internal tools only and many have been EOL’d.
// @Param       eol query    bool false "EOL Products"
// @Success     200 {object} omnitruck.ItemList
// @Failure     500 {object} services.ErrorResponse
// @Router      /products [get]
func (server *ApiService) productsHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Eol: c.Query("eol", "false"),
	}

	var data omnitruck.ItemList
	request := server.Omnitruck(c).Products(params, &data)

	if server.Mode == Opensource {
		server.logCtx(c).Info("filtering opensource products")
		data = omnitruck.SelectList(data, omnitruck.OsProductName)
	} else if params.Eol != "true" {
		server.logCtx(c).Info("filtering eol products")
		data = omnitruck.FilterList(data, omnitruck.EolProductName)
	}

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}
}

// @description Returns a valid list of valid platform keys along with full friendly names.
// @description Any of these platform keys can be used in the p query string value in various endpoints below.
// @Success     200 {object} omnitruck.PlatformList
// @Failure     500 {object} services.ErrorResponse
// @Router      /platforms [get]
func (server *ApiService) platformsHandler(c *fiber.Ctx) error {
	var data omnitruck.PlatformList
	request := server.Omnitruck(c).Platforms().ParseData(&data)

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}
}

// @description Returns a valid list of valid platform keys along with friendly names.
// @description Any of these architecture keys can be used in the p query string value in various endpoints below.
// @Success     200 {object} omnitruck.ItemList
// @Failure     500 {object} services.ErrorResponse
// @Router      /architectures [get]
func (server *ApiService) architecturesHandler(c *fiber.Ctx) error {

	var data omnitruck.ItemList
	request := server.Omnitruck(c).Architectures().ParseData(&data)

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}
}

// @description Get the latest version number for a particular channel and product combination.
// @Param       channel    path     string true  "Channel" Enums(current, stable)
// @Param       product    path     string true  "Product"
// @Param       license_id query   string false "License ID"
// @Success     200        {object} omnitruck.ProductVersion
// @Failure     400        {object} services.ErrorResponse
// @Failure     403        {object} services.ErrorResponse
// @Router      /{channel}/{product}/versions/latest [get]
func (server *ApiService) latestVersionHandler(c *fiber.Ctx) error {
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
	var request *clients.Request

	if server.Mode == Opensource {
		data, request = server.fetchLatestOSVersion(params, c)
	} else {
		data, request = server.fetchLatestVersion(params, c)
	}

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}
}

func (server *ApiService) fetchLatestVersion(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.ProductVersion, *clients.Request) {
	var data omnitruck.ProductVersion
	request := server.Omnitruck(c).LatestVersion(params).ParseData(&data)

	return data, request
}

// We need to fetch the full version list and filter out all the non-opensource versions
// Then we can return the latest OS version
func (server *ApiService) fetchLatestOSVersion(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.ProductVersion, *clients.Request) {
	var data []omnitruck.ProductVersion
	// Need to fetch all versions and filter out to only show the OS versions
	request := server.Omnitruck(c).ProductVersions(params).ParseData(&data)

	if request.Ok {
		data = omnitruck.FilterList(data, func(v omnitruck.ProductVersion) bool {
			return !omnitruck.OsProductVersion(params.Product, v)
		})
	}

	// Return the last opensource version
	// This assumes the versions are returned in ascending order
	return data[len(data)-1], request
}

// @description Get a list of all available version numbers for a particular channel and product combination
// @Param       channel    path     string true  "Channel" Enums(current, stable)
// @Param       product    path     string true  "Product"
// @Param       license_id query   string false "License ID"
// @Param       eol        query    bool   false "EOL Products" Default(false)
// @Success     200        {object} omnitruck.ItemList
// @Failure     400        {object} services.ErrorResponse
// @Failure     403        {object} services.ErrorResponse
// @Router      /{channel}/{product}/versions/all [get]
func (server *ApiService) productVersionsHandler(c *fiber.Ctx) error {
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
	request := server.Omnitruck(c).ProductVersions(params).ParseData(&data)

	switch server.Mode {
	case Commercial:
		if params.Eol != "true" {
			data = omnitruck.FilterProductList(data, params.Product, omnitruck.EolProductVersion)
		}
	case Trial:
		if params.Eol != "true" {
			data = omnitruck.FilterProductList(data, params.Product, omnitruck.EolProductVersion)
		}

		// Only return the latest version if no license is present
		if !c.Locals("valid_license").(bool) {
			data = []omnitruck.ProductVersion{
				data[len(data)-1],
			}
		}
	case Opensource:
		data = omnitruck.FilterList(data, func(v omnitruck.ProductVersion) bool {
			return !omnitruck.OsProductVersion(params.Product, v)
		})
	}

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}

}

// @description Get the full list of all packages for a particular channel and product combination.
// @description By default all packages for the latest version are returned. If the v query string parameter is included the packages for the specified version are returned.
// @Param       channel    path     string true  "Channel" Enums(current, stable)
// @Param       product    path     string true  "Product" Example(chef)
// @Param       v          query    string false "Version"
// @Param       license_id query   string false "License ID"
// @Param       eol        query    bool   false "EOL Products" Default(false)
// @Success     200        {object} omnitruck.PackageList
// @Failure     400        {object} services.ErrorResponse
// @Failure     403        {object} services.ErrorResponse
// @Router      /{channel}/{product}/packages [get]
func (server *ApiService) productPackagesHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel: c.Params("channel"),
		Product: c.Params("product"),
		Version: c.Query("v"),
		Eol:     c.Query("eol"),
	}

	if server.Mode == Opensource && isLatest(params.Version) {
		v, _ := server.fetchLatestOSVersion(params, c)
		params.Version = string(v)
	}

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	var data omnitruck.PackageList
	request := server.Omnitruck(c).ProductPackages(params).ParseData(&data)

	for platform, versions := range data {
		for version, arches := range versions {
			for arch, metadata := range arches {
				fmt.Printf("%v\n", metadata)
				params := &omnitruck.RequestParams{
					Channel:         c.Params("channel"),
					Product:         c.Params("product"),
					Version:         metadata.Version,
					Platform:        platform,
					PlatformVersion: version,
					Architecture:    arch,
				}

				metadata.Url = server.getDownloadUrl(params, metadata, c)
				data[platform][version][arch] = metadata
			}
		}
	}

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}

}

// @description Get details for a particular package.
// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Param       channel    path     string true  "Channel"                                                                                                                      Enums(current, stable)
// @Param       product    path     string true  "Product"                                                                                                                      Example(chef)
// @Param       p          query    string true  "Platform, valid values are returned from the `/platforms` endpoint."                                                          Example(ubuntu)
// @Param       pv         query    string true  "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." Example(20.04)
// @Param       m          query    string true  "Machine architecture, valid values are returned by the `/architectures` endpoint."                                            Example(x86_64)
// @Param       v          query    string false "Version of the product to be installed. A version always takes the form `x.y.z`"                                              Default(latest)
// @Param       license_id query   string false "License ID"
// @Param       eol        query    bool   false "EOL Products" Default(false)
// @Success     200        {object} omnitruck.PackageMetadata
// @Failure     400        {object} services.ErrorResponse
// @Failure     403        {object} services.ErrorResponse
// @Router      /{channel}/{product}/metadata [get]
func (server *ApiService) productMetadataHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel:         c.Params("channel"),
		Product:         c.Params("product"),
		Version:         c.Query("v"),
		Platform:        c.Query("p"),
		PlatformVersion: c.Query("pv"),
		Architecture:    c.Query("m"),
	}

	if server.Mode == Opensource && isLatest(params.Version) {
		v, _ := server.fetchLatestOSVersion(params, c)
		params.Version = string(v)
	}

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	var data omnitruck.PackageMetadata
	request := server.Omnitruck(c).ProductMetadata(params).ParseData(&data)

	// Remap the package url to our download URL
	url := server.getDownloadUrl(params, data, c)
	data.Url = url

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}
}

func (server *ApiService) getDownloadUrl(params *omnitruck.RequestParams, m omnitruck.PackageMetadata, c *fiber.Ctx) string {
	u, _ := url.Parse(c.BaseURL())
	path, _ := url.JoinPath(params.Channel, params.Product, "download")
	u.Path = path
	p := params.UrlParams()
	p.Add("license_id", c.Locals("license_id").(string))
	u.RawQuery = p.Encode()

	return u.String()
}

// @description Get details for a particular package.
// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Param       channel    path   string true  "Channel"                                                                                                                      Enums(current, stable)
// @Param       product    path   string true  "Product"                                                                                                                      Example(chef)
// @Param       p          query  string true  "Platform, valid values are returned from the `/platforms` endpoint."                                                          Example(ubuntu)
// @Param       pv         query  string true  "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." Example(20.04)
// @Param       m          query  string true  "Machine architecture, valid values are returned by the `/architectures` endpoint."                                            Example(x86_64)
// @Param       v          query  string false "Version of the product to be installed. A version always takes the form `x.y.z`"                                              Default(latest)
// @Param       license_id query string false "License ID"
// @Param       eol        query  bool   false "EOL Products" Default(false)
// @Success     302
// @Failure     400 {object} services.ErrorResponse
// @Failure     403 {object} services.ErrorResponse
// @Router      /{channel}/{product}/download [get]
func (server *ApiService) productDownloadHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Channel:         c.Params("channel"),
		Product:         c.Params("product"),
		Version:         c.Query("v"),
		Platform:        c.Query("p"),
		PlatformVersion: c.Query("pv"),
		Architecture:    c.Query("m"),
	}

	if server.Mode == Opensource && isLatest(params.Version) {
		v, _ := server.fetchLatestOSVersion(params, c)
		params.Version = string(v)
	}

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	var data omnitruck.PackageMetadata
	request := server.Omnitruck(c).ProductDownload(params).ParseData(&data)

	if request.Ok {
		server.logCtx(c).Infof("Redirecting user to %s", data.Url)
		return c.Redirect(data.Url, 302)
	} else {
		return server.SendError(c, request)
	}
}
