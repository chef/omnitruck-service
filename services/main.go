package services

import (
	"os"
	"strings"
	"time"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
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
	server.App.Get("/relatedProducts", server.relatedProductsHandler)
	server.App.Get("/:channel/:product/fileName", server.fileNameHandler)

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
// @description Any of these product keys can be used in the <PRODUCT> value of other endpoints. Please note many of these products are used for internal tools only and many have been EOL'd.
// @Param       eol query    bool false "EOL Products"
// @Success     200 {object} omnitruck.ItemList
// @Failure     500 {object} services.ErrorResponse
// @Router      /products [get]
func (server *ApiService) productsHandler(c *fiber.Ctx) error {
	params := getRequestParams(c)

	var data omnitruck.ItemList
	request := server.Omnitruck(c).Products(params, &data)

	data = server.DynamoServices(server.DatabaseService, c).Products(data, params.Eol)

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

	data = server.DynamoServices(server.DatabaseService, c).Platforms(data)
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
// @Param       license_id query    string false "License ID"
// @Success     200        {object} omnitruck.ProductVersion
// @Failure     400        {object} services.ErrorResponse
// @Failure     403        {object} services.ErrorResponse
// @Router      /{channel}/{product}/versions/latest [get]
func (server *ApiService) latestVersionHandler(c *fiber.Ctx) error {
	params := getRequestParams(c)
	// Force version to always be latest
	params.Version = "latest"

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	var data omnitruck.ProductVersion
	var request *clients.Request

	if params.Product == "automate" || params.Product == "habitat" {

		if server.Mode == Opensource {
			data, err := server.DynamoServices(server.DatabaseService, c).FetchLatestOsVersion(params)
			if err != nil {
				return server.SendErrorResponse(c, fiber.StatusInternalServerError, "Error while fetching the latest version for the product.")
			}
			return server.SendResponse(c, &data)
		} else {
			data, err = server.DynamoServices(server.DatabaseService, c).VersionLatest(params)

			if err != nil {
				return server.SendErrorResponse(c, fiber.StatusInternalServerError, "Error while fetching the latest version for the product.")
			}
			return server.SendResponse(c, &data)
		}

	}

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
// @Param       license_id query    string false "License ID"
// @Param       eol        query    bool   false "EOL Products" Default(false)
// @Success     200        {object} omnitruck.ItemList
// @Failure     400        {object} services.ErrorResponse
// @Failure     403        {object} services.ErrorResponse
// @Router      /{channel}/{product}/versions/all [get]
func (server *ApiService) productVersionsHandler(c *fiber.Ctx) error {
	params := getRequestParams(c)

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	if params.Product == "automate" || params.Product == "habitat" {
		data, err := server.DynamoServices(server.DatabaseService, c).VersionAll(params)
		if err != nil {
			return server.SendErrorResponse(c, fiber.StatusInternalServerError, "Error while fetching product versions")
		}

		if params.Product == "habitat" && server.Mode == Opensource {
			data = omnitruck.FilterList(data, func(v omnitruck.ProductVersion) bool {
				return !omnitruck.OsProductVersion(params.Product, v)
			})
		}
		if server.Mode == Trial {
			data = []omnitruck.ProductVersion{
				data[len(data)-1],
			}
		}

		return server.SendResponse(c, &data)

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
// @Param       license_id query    string false "License ID"
// @Param       eol        query    bool   false "EOL Products" Default(false)
// @Success     200        {object} omnitruck.PackageList
// @Failure     400        {object} services.ErrorResponse
// @Failure     403        {object} services.ErrorResponse
// @Router      /{channel}/{product}/packages [get]
func (server *ApiService) productPackagesHandler(c *fiber.Ctx) error {
	params := getRequestParams(c)
	var data omnitruck.PackageList
	if params.Product == "automate" || params.Product == "habitat" {

		if server.Mode == Opensource && isLatest(params.Version) {
			v, err := server.DynamoServices(server.DatabaseService, c).FetchLatestOsVersion(params)
			if err != nil {
				return server.SendErrorResponse(c, fiber.StatusInternalServerError, "Error while fetching the information for the product.")
			}
			params.Version = v
		}

		err, ok := server.ValidateRequest(params, c)
		if !ok {
			return err
		}
		data, err = server.DynamoServices(server.DatabaseService, c).ProductPackages(params)
		if err != nil {
			return server.SendErrorResponse(c, fiber.StatusInternalServerError, "Error while fetching the information for the product.")
		} else if len(data) == 0 {
			return server.SendErrorResponse(c, fiber.StatusBadRequest, "Product information not found. Please check the input parameters.")
		}

		data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
			params.Version = m.Version
			params.Platform = platform
			params.Architecture = arch

			m.Url = getDownloadUrl(params, c)

			return m
		})
		return server.SendResponse(c, &data)
	}

	if server.Mode == Opensource && isLatest(params.Version) {
		v, _ := server.fetchLatestOSVersion(params, c)
		params.Version = string(v)
	}

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	request := server.Omnitruck(c).ProductPackages(params).ParseData(&data)
	p := getRequestParams(c)
	data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
		p.Version = m.Version
		p.Platform = platform
		p.PlatformVersion = pv
		p.Architecture = arch

		m.Url = getDownloadUrl(p, c)

		return m
	})

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
// @Param       license_id query    string false "License ID"
// @Param       eol        query    bool   false "EOL Products" Default(false)
// @Success     200        {object} omnitruck.PackageMetadata
// @Failure     400        {object} services.ErrorResponse
// @Failure     403        {object} services.ErrorResponse
// @Router      /{channel}/{product}/metadata [get]
func (server *ApiService) productMetadataHandler(c *fiber.Ctx) error {
	params := getRequestParams(c)
	var data omnitruck.PackageMetadata
	if params.Product == "automate" || params.Product == "habitat" {
		if server.Mode == Opensource && isLatest(params.Version) {
			v, err := server.DynamoServices(server.DatabaseService, c).FetchLatestOsVersion(params)
			if err != nil {
				return server.SendErrorResponse(c, fiber.StatusInternalServerError, "Error while fetching the information for the product.")
			}
			params.Version = v
		}
		err, ok := server.ValidateRequest(params, c)
		if !ok {
			return err
		}
		data, err = server.DynamoServices(server.DatabaseService, c).ProductMetadata(params)

		if err != nil {
			return server.SendErrorResponse(c, fiber.StatusInternalServerError, "Error while fetching the information for the product.")
		} else if data == (omnitruck.PackageMetadata{}) {
			return server.SendErrorResponse(c, fiber.StatusBadRequest, "Product information not found. Please check the input parameters.")
		}
		// Remap the package url to our download URL
		url := getDownloadUrl(params, c)
		data.Url = url
		return server.SendResponse(c, &data)
	}
	if server.Mode == Opensource && isLatest(params.Version) {
		v, _ := server.fetchLatestOSVersion(params, c)
		params.Version = string(v)
	}

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		return err
	}

	request := server.Omnitruck(c).ProductMetadata(params).ParseData(&data)

	// Remap the package url to our download URL
	url := getDownloadUrl(params, c)
	data.Url = url

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}
}

// @description Get details for a particular package.
// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Param       channel    path   string true  "Channel"                                                                                                                      Enums(current, stable)
// @Param       product    path   string true  "Product"                                                                                                                      Example(chef)
// @Param       p          query  string true  "Platform, valid values are returned from the `/platforms` endpoint."                                                          Example(ubuntu)
// @Param       pv         query  string true  "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." Example(20.04)
// @Param       m          query  string true  "Machine architecture, valid values are returned by the `/architectures` endpoint."                                            Example(x86_64)
// @Param       v          query  string false "Version of the product to be installed. A version always takes the form `x.y.z`"                                              Default(latest)
// @Param       license_id query  string false "License ID"
// @Param       eol        query  bool   false "EOL Products" Default(false)
// @Success     302
// @Failure     400 {object} services.ErrorResponse
// @Failure     403 {object} services.ErrorResponse
// @Router      /{channel}/{product}/download [get]
func (server *ApiService) productDownloadHandler(c *fiber.Ctx) error {
	params := getRequestParams(c)
	flag := verifyRequestType(params)

	if params.Product == "automate" || params.Product == "habitat" {
		if server.Mode == Opensource && isLatest(params.Version) {
			v, err := server.DynamoServices(server.DatabaseService, c).FetchLatestOsVersion(params)
			if err != nil {
				return server.SendErrorResponse(c, fiber.StatusInternalServerError, "Error while fetching the information for the product.")
			}
			params.Version = v
		}
		err, ok := server.ValidateRequest(params, c)
		if !ok {
			return err
		}
		url, err := server.DynamoServices(server.DatabaseService, c).ProductDownload(params)
		if err != nil {
			return server.SendErrorResponse(c, fiber.StatusInternalServerError, "Error while fetching the information for the product.")
		} else if url == "" {
			return server.SendErrorResponse(c, fiber.StatusBadRequest, "Product information not found. Please check the input parameters.")
		}
		server.logCtx(c).Infof("Redirecting user to %s", url)
		return c.Redirect(url, 302)
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
		if flag {
			data.Url = data.Url + substring
		}
		server.logCtx(c).Infof("Redirecting user to %s", data.Url)
		return c.Redirect(data.Url, 302)
	} else {
		return server.SendError(c, request)
	}
}

// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Param       bom    	   query string true  "bom"
// @Param       license_id query string false "License ID"
// @Success     200
// @Failure     400 {object} services.ErrorResponse
// @Failure     403 {object} services.ErrorResponse
// @Router      /relatedProducts [get]
func (server *ApiService) relatedProductsHandler(c *fiber.Ctx) error {
	params := getRequestParams(c)

	server.Log.Info("Validating related products API for " + params.BOM)

	err, ok := server.ValidateRequest(params, c)
	if !ok {
		server.Log.Error("Validation of related products API for "+params.BOM+"failed", err.Error())
		return err
	}

	relatedProducts, err := server.DatabaseService.GetRelatedProducts(params.BOM)

	if err != nil {
		request := clients.Request{}
		server.Log.Error("Error while fetching related products for "+params.BOM, err.Error())
		return server.SendError(c, request.Failure(fiber.StatusInternalServerError, "Unable to retrieve related products for "+params.BOM))
	}

	if len(relatedProducts.Products) == 0 {
		request := clients.Request{}
		server.Log.Error("No related products found for " + params.BOM)
		return server.SendError(c, request.Failure(fiber.StatusBadRequest, "No related products found for BOM"))
	}

	response := map[string]interface{}{
		"relatedProducts": relatedProducts.Products,
	}
	server.Log.Info("Returning success response from related products API for " + params.BOM)
	return server.SendResponse(c, response)

}

// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Param       channel    path     string true  "Channel"                                                                                                                      Enums(current, stable)
// @Param       product    path     string true  "Product"                                                                                                                      Example(chef)
// @Param       p          query    string true  "Platform, valid values are returned from the `/platforms` endpoint."                                                          Example(ubuntu)
// @Param       pv         query    string true  "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." Example(20.04)
// @Param       m          query    string true  "Machine architecture, valid values are returned by the `/architectures` endpoint."                                            Example(x86_64)
// @Param       v          query    string false "Version of the product to be installed. A version always takes the form `x.y.z`"                                              Default(latest)
// @Param       license_id query    string false "License ID"
// @Success     200        {object} map[string]interface{}
// @Failure     400        {object} services.ErrorResponse
// @Failure     403        {object} services.ErrorResponse
// @Router      /{channel}/{product}/fileName [get]
func (server *ApiService) fileNameHandler(c *fiber.Ctx) error {
	params := getRequestParams(c)
	server.Log.Info("Validating download file name for " + params.Product + "in channel " + params.Channel)
	err, ok := server.ValidateRequest(params, c)
	if !ok {
		server.Log.Error("Validation of file name API for "+params.Product+"failed", err.Error())
		return err
	}

	//assuming that the metadata table will always have only the latest version record for automate, querying db without sortkey
	if params.Product == constants.AUTOMATE_PRODUCT || params.Product == constants.HABITAT_PRODUCT {
		version := params.Version
		if params.Version == constants.LATEST || params.Version == "" {
			latestVersion, err := server.DatabaseService.GetVersionLatest(params.Product)
			if err != nil {
				request := clients.Request{}
				server.Log.Error("Error while getting latest version for fetching fileName for "+params.Product, err.Error())
				return server.SendError(c, request.Failure(500, "Unable to get latest version of "+params.Product))
			} else {
				version = latestVersion
			}
		}
		metadata, err := server.DatabaseService.GetMetaData(params.Product, version, params.Platform, params.PlatformVersion, params.Architecture)

		if err != nil {
			request := clients.Request{}
			server.Log.Error("Error while fetching fileName for "+params.Product, err.Error())
			return server.SendError(c, request.Failure(500, "Error while fetching file name"))
		}
		if metadata == nil || metadata.FileName == "" {
			request := clients.Request{}
			server.Log.Error("Error while fetching fileName for "+params.Product, "Unable to find the product information for given parameters")
			return server.SendError(c, request.Failure(400, "Unable to find the product information for given parameters"))
		}
		response := map[string]interface{}{
			"fileName": metadata.FileName,
		}
		server.Log.Info("Returning success response from fileName API for " + params.Product)
		return server.SendResponse(c, response)

	} else {
		var data omnitruck.PackageMetadata
		request := server.Omnitruck(c).ProductMetadata(params).ParseData(&data)

		if request.Ok {
			url := data.Url
			fileName := getFileNameFromURL(url)
			response := map[string]interface{}{
				"fileName": fileName,
			}
			server.Log.Info("Returning success response from fileName API for " + params.Product)
			return server.SendResponse(c, response)
		} else {
			return server.SendError(c, request)
		}

	}
}

func getFileNameFromURL(url string) string {
	segments := strings.Split(url, "/")
	return segments[len(segments)-1]
}
