package handler

//TODO: Remove this file when the new handler is fully implemented and tested.

// import (
// 	"bufio"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"html"
// 	"io"
// 	"net/http"
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/chef/omnitruck-service/clients"
// 	"github.com/chef/omnitruck-service/clients/omnitruck"
// 	"github.com/chef/omnitruck-service/constants"
// 	_ "github.com/chef/omnitruck-service/docs"
// 	"github.com/chef/omnitruck-service/internal/services"
// 	"github.com/chef/omnitruck-service/utils"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/requestid"
// 	"github.com/gofiber/swagger"
// 	"github.com/gomarkdown/markdown"
// )

// // @title        Licensed Omnitruck API
// // @version      1.0
// // @description  Licensed Omnitruck API
// // @license.name Apache 2.0
// // @license.url  http://www.apache.org/licenses/LICENSE-2.0.html
// func (server *ApiService) buildRouter() {
// 	server.App.Get("/swagger/*", swagger.New(swagger.Config{
// 		InstanceName: "OmnitruckApi",
// 	}))

// 	// Add the endpoints that don't require any special handling for various APIs
// 	server.App.Static("/", "./static", fiber.Static{
// 		Compress:      true,
// 		ByteRange:     true,
// 		Browse:        false,
// 		Index:         "index.html",
// 		CacheDuration: 10 * time.Second,
// 		MaxAge:        3600,
// 	})
// 	server.App.Get("/status", requestid.New(), server.HealthCheck)
// 	server.App.Get("/products", requestid.New(), server.productsHandler)
// 	server.App.Get("/platforms", requestid.New(), server.platformsHandler)
// 	server.App.Get("/package-managers", requestid.New(), server.packageManagersHandler)
// 	server.App.Get("/architectures", requestid.New(), server.architecturesHandler)
// 	server.App.Get("/:channel/:product/versions/latest", requestid.New(), server.latestVersionHandler)
// 	server.App.Get("/:channel/:product/versions/all", requestid.New(), server.productVersionsHandler)
// 	server.App.Get("/:channel/:product/packages", requestid.New(), server.productPackagesHandler)
// 	server.App.Get("/:channel/:product/metadata", requestid.New(), server.productMetadataHandler)
// 	server.App.Get("/:channel/:product/download", requestid.New(), server.productDownloadHandler)
// 	server.App.Get("/relatedProducts", requestid.New(), server.relatedProductsHandler)
// 	server.App.Get("/:channel/:product/fileName", requestid.New(), server.fileNameHandler)
// 	server.App.Get("/install.sh", requestid.New(), server.downloadLinuxScript)
// 	server.App.Get("/install.ps1", requestid.New(), server.downloadWindowsScript)
// }

// var JsonUnmarshal = func(data []byte, v any) error {
// 	return json.Unmarshal(data, &v)
// }

// var IoCopy = func(dst io.Writer, src io.Reader) (written int64, err error) {
// 	return io.Copy(dst, src)
// }

// func (server *ApiService) DocsHandler(baseUrl string) func(*fiber.Ctx) error {
// 	content, err := os.ReadFile("docs/index.md")
// 	if err != nil {
// 		content = []byte("Error reading docs/index.md")
// 	}
// 	output := markdown.ToHTML(content, nil, nil)
// 	view_data := fiber.Map{
// 		"Content": string(output),
// 		"baseUrl": baseUrl,
// 	}

// 	return func(c *fiber.Ctx) error {
// 		return c.Render("docs", view_data, "layouts/docs")
// 	}
// }

// // @description Returns a valid list of valid product keys.
// // @description Any of these product keys can be used in the <PRODUCT> value of other endpoints. Please note many of these products are used for internal tools only and many have been EOL'd.
// // @Param       eol query    bool false "EOL Products"
// // @Success     200 {object} omnitruck.ItemList
// // @Failure     500 {object} services.ErrorResponse
// // @Router      /products [get]
// func (server *ApiService) productsHandler(c *fiber.Ctx) error {
// 	params := getRequestParams(c)

// 	var data omnitruck.ItemList
// 	request := server.Omnitruck(c).Products(params, &data)

// 	data = server.DynamoServices(server.DatabaseService, c).Products(data, params.Eol)

// 	if server.Mode == Opensource {
// 		server.logCtx(c).Info("filtering opensource products")
// 		data = omnitruck.SelectList(data, omnitruck.OsProductName)
// 	} else if params.Eol != "true" {
// 		server.logCtx(c).Info("filtering eol products")
// 		data = omnitruck.FilterList(data, omnitruck.EolProductName)
// 	}

// 	if server.Mode == Trial {
// 		data = omnitruck.FilterProductsForFreeTrial(data, omnitruck.ProductsForFreeTrial)
// 		omnitruck.ProductDisplayName(data)
// 	}

// 	if server.Mode == Commercial {
// 		data = append(data, constants.PLATFORM_SERVICE_PRODUCT)
// 	}

// 	if request.Ok {
// 		return server.SendResponse(c, &data)
// 	} else {
// 		return server.SendError(c, request)
// 	}
// }

// // @description Returns a valid list of valid platform keys along with full friendly names.
// // @description Any of these platform keys can be used in the p query string value in various endpoints below.
// // @Success     200 {object} omnitruck.PlatformList
// // @Failure     500 {object} services.ErrorResponse
// // @Router      /platforms [get]
// func (server *ApiService) PlatformsHandler(c *fiber.Ctx) error {
// 	var data omnitruck.PlatformList
// 	request := server.Omnitruck(c).Platforms().ParseData(&data)

// 	data = server.DynamoServices(server.DatabaseService, c).Platforms(data)
// 	if request.Ok {
// 		return server.SendResponse(c, &data)
// 	} else {
// 		return server.SendError(c, request)
// 	}
// }

// // @description Returns a valid list of valid platform keys along with friendly names.
// // @description Any of these architecture keys can be used in the p query string value in various endpoints below.
// // @Success     200 {object} omnitruck.ItemList
// // @Failure     500 {object} services.ErrorResponse
// // @Router      /architectures [get]
// func (server *ApiService) architecturesHandler(c *fiber.Ctx) error {

// 	var data omnitruck.ItemList
// 	request := server.Omnitruck(c).Architectures().ParseData(&data)

// 	if request.Ok {
// 		return server.SendResponse(c, &data)
// 	} else {
// 		return server.SendError(c, request)
// 	}
// }

// // @description Get the latest version number for a particular channel and product combination.
// // @Param       channel    path     string true  "Channel" Enums(current, stable)
// // @Param       product    path     string true  "Product"
// // @Param       license_id query    string false "License ID"
// // @Success     200        {object} omnitruck.ProductVersion
// // @Failure     400        {object} services.ErrorResponse
// // @Failure     403        {object} services.ErrorResponse
// // @Router      /{channel}/{product}/versions/latest [get]
// func (server *ApiService) latestVersionHandler(c *fiber.Ctx) error {
// 	params := getRequestParams(c)
// 	// Force version to always be latest
// 	params.Version = "latest"

// 	err, ok := server.ValidateRequest(params, c)
// 	if !ok {
// 		return err
// 	}

// 	var data omnitruck.ProductVersion
// 	var request *clients.Request

// 	if server.Mode == Opensource {
// 		data, request = server.fetchLatestOSVersion(params, c)
// 	} else {
// 		data, request = server.fetchLatestVersion(params, c)
// 	}

// 	if request.Ok {
// 		return server.SendResponse(c, &data)
// 	} else {
// 		return server.SendError(c, request)
// 	}
// }

// func (server *ApiService) FetchLatestVersion(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.ProductVersion, *clients.Request) {
// 	var data omnitruck.ProductVersion
// 	if params.Product == constants.AUTOMATE_PRODUCT || params.Product == constants.HABITAT_PRODUCT {
// 		request := clients.Request{}
// 		data, err := server.DynamoServices(server.DatabaseService, c).VersionLatest(params)
// 		if err != nil {
// 			code, msg := getErrorCodeAndMsg(err)
// 			server.logCtx(c).WithError(err).Error(utils.ErrorWhileFetchingLatestVersion + params.Product)
// 			request.Failure(code, msg)
// 			return data, &request
// 		} else {
// 			request.Success()
// 			return data, &request
// 		}
// 	} else if params.Product == constants.PLATFORM_SERVICE_PRODUCT {
// 		request := clients.Request{}
// 		data, err := server.PlatformServices(c).PlatformVersionLatest(params, int(server.Mode))
// 		if err != nil {
// 			code, msg := getErrorCodeAndMsg(err)
// 			server.logCtx(c).WithError(err).Error(utils.ErrorWhileFetchingLatestVersion + params.Product)
// 			request.Failure(code, msg)
// 			return data, &request
// 		}
// 		request.Success()
// 		return data, &request
// 	}
// 	request := server.Omnitruck(c).LatestVersion(params).ParseData(&data)

// 	return data, request
// }

// // We need to fetch the full version list and filter out all the non-opensource versions
// // Then we can return the latest OS version
// func (server *ApiService) FetchLatestOSVersion(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.ProductVersion, *clients.Request) {
// 	var data []omnitruck.ProductVersion
// 	if params.Product == constants.PLATFORM_SERVICE_PRODUCT {
// 		request := clients.Request{}
// 		data, err := server.PlatformServices(c).PlatformVersionLatest(params, int(server.Mode))
// 		if err != nil {
// 			code, msg := getErrorCodeAndMsg(err)
// 			server.logCtx(c).WithError(err).Error(utils.ErrorWhileFetchingLatestVersion + params.Product)
// 			request.Failure(code, msg)
// 			return data, &request
// 		}
// 		request.Success()
// 		return data, &request
// 	}
// 	// Need to fetch all versions and filter out to only show the OS versions
// 	if params.Product == constants.AUTOMATE_PRODUCT || params.Product == constants.HABITAT_PRODUCT {
// 		request := clients.Request{}
// 		latestVersion, err := server.DynamoServices(server.DatabaseService, c).FetchLatestOsVersion(params)
// 		if err != nil {
// 			code, msg := getErrorCodeAndMsg(err)
// 			server.logCtx(c).WithError(err).Error("Error while fetching the latest opensource version for the " + params.Product)
// 			request.Failure(code, msg)
// 			return omnitruck.ProductVersion(latestVersion), &request
// 		} else {
// 			request.Success()
// 			return omnitruck.ProductVersion(latestVersion), &request
// 		}
// 	}
// 	request := server.Omnitruck(c).ProductVersions(params).ParseData(&data)

// 	if request.Ok {
// 		data = omnitruck.FilterList(data, func(v omnitruck.ProductVersion) bool {
// 			return !omnitruck.OsProductVersion(params.Product, v)
// 		})
// 	}
// 	if len(data) == 0 {
// 		data = append(data, "")
// 	}

// 	// Return the last opensource version
// 	// This assumes the versions are returned in ascending order
// 	return data[len(data)-1], request
// }

// // @description Get a list of all available version numbers for a particular channel and product combination
// // @Param       channel    path     string true  "Channel" Enums(current, stable)
// // @Param       product    path     string true  "Product"
// // @Param       license_id query    string false "License ID"
// // @Param       eol        query    bool   false "EOL Products" Default(false)
// // @Success     200        {object} omnitruck.ItemList
// // @Failure     400        {object} services.ErrorResponse
// // @Failure     403        {object} services.ErrorResponse
// // @Router      /{channel}/{product}/versions/all [get]
// func (server *ApiService) productVersionsHandler(c *fiber.Ctx) error {
// 	params := getRequestParams(c)
// 	var data []omnitruck.ProductVersion

// 	err, ok := server.ValidateRequest(params, c)
// 	if !ok {
// 		return err
// 	}

// 	if params.Product == constants.AUTOMATE_PRODUCT || params.Product == constants.HABITAT_PRODUCT {
// 		return server.createDynamoServiceResponse(params, c)
// 	} else if params.Product == constants.PLATFORM_SERVICE_PRODUCT {
// 		versions, err := server.PlatformServices(c).PlatformVersionsAll(params, int(server.Mode))
// 		if err != nil {
// 			code, msg := getErrorCodeAndMsg(err)
// 			return server.SendErrorResponse(c, code, msg)
// 		}
// 		return server.SendResponse(c, &versions)
// 	}

// 	request := server.Omnitruck(c).ProductVersions(params).ParseData(&data)

// 	switch server.Mode {
// 	case Commercial:
// 		if params.Eol != "true" {
// 			data = omnitruck.FilterProductList(data, params.Product, omnitruck.EolProductVersion)
// 		}
// 	case Trial:
// 		if params.Eol != "true" {
// 			data = omnitruck.FilterProductList(data, params.Product, omnitruck.EolProductVersion)
// 		}

// 		if len(data) == 0 {
// 			data = append(data, "")
// 		}
// 		data = []omnitruck.ProductVersion{
// 			data[len(data)-1],
// 		}
// 	case Opensource:
// 		data = omnitruck.FilterList(data, func(v omnitruck.ProductVersion) bool {
// 			return !omnitruck.OsProductVersion(params.Product, v)
// 		})
// 	}

// 	if request.Ok {
// 		return server.SendResponse(c, &data)
// 	} else {
// 		return server.SendError(c, request)
// 	}

// }

// func (server *ApiService) createDynamoServiceResponse(params *omnitruck.RequestParams, c *fiber.Ctx) error {
// 	data, err := server.DynamoServices(server.DatabaseService, c).VersionAll(params)
// 	if err != nil {
// 		code, msg := services.GetErrorCodeAndMsg(err)
// 		return server.SendErrorResponse(c, code, msg)
// 	}

// 	if params.Product == constants.HABITAT_PRODUCT && server.Mode == Opensource {
// 		data = omnitruck.FilterList(data, func(v omnitruck.ProductVersion) bool {
// 			return !omnitruck.OsProductVersion(params.Product, v)
// 		})
// 	}
// 	if server.Mode == Trial {
// 		data = []omnitruck.ProductVersion{
// 			data[len(data)-1],
// 		}
// 	}

// 	return server.SendResponse(c, &data)
// }

// // @description Get the full list of all packages for a particular channel and product combination.
// // @description By default all packages for the latest version are returned. If the v query string parameter is included the packages for the specified version are returned.
// // @Param       channel    path     string true  "Channel" Enums(current, stable)
// // @Param       product    path     string true  "Product" Example(chef)
// // @Param       v          query    string false "Version"
// // @Param       license_id query    string false "License ID"
// // @Param       eol        query    bool   false "EOL Products" Default(false)
// // @Success     200        {object} omnitruck.PackageList
// // @Failure     400        {object} services.ErrorResponse
// // @Failure     403        {object} services.ErrorResponse
// // @Router      /{channel}/{product}/packages [get]
// func (server *ApiService) productPackagesHandler(c *fiber.Ctx) error {
// 	params := getRequestParams(c)
// 	var data omnitruck.PackageList
// 	err, ok := server.ValidateRequest(params, c)
// 	if !ok {
// 		return err
// 	}

// 	if params.Product == constants.PLATFORM_SERVICE_PRODUCT {
// 		data, err = server.PlatformServices(c).PlatformPackages(params, int(server.Mode))
// 		if err != nil {
// 			code, msg := getErrorCodeAndMsg(err)
// 			return server.SendErrorResponse(c, code, msg)
// 		}

// 		data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
// 			params.Version = m.Version
// 			params.Platform = platform
// 			params.Architecture = arch

// 			m.Url = getDownloadUrl(params, c)

// 			return m
// 		})
// 		return server.SendResponse(c, &data)
// 	}

// 	err = server.versionCheckForTrialAndOsServer(params, c)
// 	if err != nil {
// 		code, msg := services.GetErrorCodeAndMsg(err)
// 		return server.SendErrorResponse(c, code, msg)
// 	}

// 	if params.Product == constants.AUTOMATE_PRODUCT || params.Product == constants.HABITAT_PRODUCT {
// 		data, err = server.DynamoServices(server.DatabaseService, c).ProductPackages(params)
// 		if err != nil {
// 			code, msg := getErrorCodeAndMsg(err)
// 			return server.SendErrorResponse(c, code, msg)
// 		}

// 		data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
// 			params.Version = m.Version
// 			params.Platform = platform
// 			params.Architecture = arch

// 			m.Url = getDownloadUrl(params, c)

// 			return m
// 		})
// 		return server.SendResponse(c, &data)
// 	}

// 	request := server.Omnitruck(c).ProductPackages(params).ParseData(&data)
// 	p := getRequestParams(c)
// 	data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
// 		p.Version = m.Version
// 		p.Platform = platform
// 		p.PlatformVersion = pv
// 		p.Architecture = arch

// 		m.Url = getDownloadUrl(p, c)

// 		return m
// 	})

// 	if request.Ok {
// 		return server.SendResponse(c, &data)
// 	} else {
// 		return server.SendError(c, request)
// 	}
// }

// // @description Get details for a particular package.
// // @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// // @Param       channel    path     string true  "Channel"                                                                                                                      Enums(current, stable)
// // @Param       product    path     string true  "Product"                                                                                                                      Example(chef)
// // @Param       p          query    string true  "Platform, valid values are returned from the `/platforms` endpoint."                                                          Example(ubuntu)
// // @Param       pv         query    string true  "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." Example(20.04)
// // @Param       m          query    string true  "Machine architecture, valid values are returned by the `/architectures` endpoint."                                            Example(x86_64)
// // @Param       v          query    string false "Version of the product to be installed. A version always takes the form `x.y.z`"                                              Default(latest)
// // @Param       license_id query    string false "License ID"
// // @Param       eol        query    bool   false "EOL Products" Default(false)
// // @Success     200        {object} omnitruck.PackageMetadata
// // @Failure     400        {object} services.ErrorResponse
// // @Failure     403        {object} services.ErrorResponse
// // @Router      /{channel}/{product}/metadata [get]
// func (server *ApiService) productMetadataHandler(c *fiber.Ctx) error {
// 	params := getRequestParams(c)
// 	var data omnitruck.PackageMetadata
// 	var request *clients.Request
// 	err, ok := server.ValidateRequest(params, c)
// 	if !ok {
// 		return err
// 	}

// 	if params.Product == constants.PLATFORM_SERVICE_PRODUCT {
// 		request = &clients.Request{}
// 		data, err = server.PlatformServices(c).PlatformMetadata(params, int(server.Mode))
// 		if err != nil {
// 			code, msg := getErrorCodeAndMsg(err)
// 			return server.SendErrorResponse(c, code, msg)
// 		}
// 		return server.getChefPlatformMetaData(params, data, c)

// 	}

// 	err = server.versionCheckForTrialAndOsServer(params, c)
// 	if err != nil {
// 		code, msg := getErrorCodeAndMsg(err)
// 		return server.SendErrorResponse(c, code, msg)
// 	}

// 	if params.Product == constants.AUTOMATE_PRODUCT || params.Product == constants.HABITAT_PRODUCT {
// 		request = &clients.Request{}
// 		data, err = server.DynamoServices(server.DatabaseService, c).ProductMetadata(params)
// 		if err != nil {
// 			code, msg := getErrorCodeAndMsg(err)
// 			request.Failure(code, msg)
// 		} else {
// 			request.Success()
// 		}

// 	} else {
// 		request = server.Omnitruck(c).ProductMetadata(params).ParseData(&data)
// 	}

// 	// Remap the package url to our download URL
// 	url := getDownloadUrl(params, c)
// 	data.Url = url

// 	if request.Ok {
// 		return server.SendResponse(c, &data)
// 	} else {
// 		return server.SendError(c, request)
// 	}
// }

// func (server *ApiService) getChefPlatformMetaData(params *omnitruck.RequestParams, data omnitruck.PackageMetadata, c *fiber.Ctx) error {
// 	url := getDownloadUrl(params, c)
// 	data.Url = url
// 	if params.Version == "" {
// 		data.Version = "latest"
// 	}
// 	return server.SendResponse(c, &data)
// }

// // @description Get details for a particular package.
// // @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// // @Param       channel    path   string true  "Channel"                                                                                                                      Enums(current, stable)
// // @Param       product    path   string true  "Product"                                                                                                                      Example(chef)
// // @Param       p          query  string true  "Platform, valid values are returned from the `/platforms` endpoint."                                                          Example(ubuntu)
// // @Param       pv         query  string true  "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." Example(20.04)
// // @Param       m          query  string true  "Machine architecture, valid values are returned by the `/architectures` endpoint."                                            Example(x86_64)
// // @Param       v          query  string false "Version of the product to be installed. A version always takes the form `x.y.z`"                                              Default(latest)
// // @Param       license_id query  string false "License ID"
// // @Param       eol        query  bool   false "EOL Products" Default(false)
// // @Success     302
// // @Failure     400 {object} services.ErrorResponse
// // @Failure     403 {object} services.ErrorResponse
// // @Router      /{channel}/{product}/download [get]
// func (server *ApiService) productDownloadHandler(c *fiber.Ctx) error {
// 	params := getRequestParams(c)
// 	flag := verifyRequestType(params)
// 	err, ok := server.ValidateRequest(params, c)
// 	if !ok {
// 		return err
// 	}
// 	server.logCtx(c).Infof("Received product download request for %s", params.Product)

// 	err = server.versionCheckForTrialAndOsServer(params, c)
// 	if err != nil {
// 		code, msg := getErrorCodeAndMsg(err)
// 		return server.SendErrorResponse(c, code, msg)
// 	}

// 	if params.Product == constants.AUTOMATE_PRODUCT || params.Product == constants.HABITAT_PRODUCT {
// 		return server.downloadAutomateOrHabitat(params, c)
// 	}

// 	if params.Product == constants.PLATFORM_SERVICE {
// 		if server.Mode == Commercial {
// 			return server.downloadChefPlatform(params, c)
// 		}
// 		return server.SendErrorResponse(c, http.StatusBadRequest, constants.PLATFORM_ERROR)
// 	}

// 	var data omnitruck.PackageMetadata
// 	request := server.Omnitruck(c).ProductDownload(params).ParseData(&data)

// 	if request.Ok {
// 		if flag {
// 			data.Url = data.Url + substring
// 		}
// 		server.logCtx(c).Infof("Redirecting user to %s", data.Url)
// 		return c.Redirect(data.Url, 302)
// 	} else {
// 		return server.SendError(c, request)
// 	}
// }

// func (server *ApiService) DownloadChefPlatform(params *omnitruck.RequestParams, c *fiber.Ctx) error {
// 	//1. Get the replicated customer email associated to licenseId
// 	resp := clients.Response{}
// 	request := server.LicenseClient.GetReplicatedCustomerEmail(params.LicenseId, server.Config.ServiceConfig.LicenseServiceUrl, &resp)

// 	if !request.Ok {
// 		server.logCtx(c).Errorf("Received error response from getReplicatedCustomer")
// 		return server.SendErrorResponse(c, request.Code, request.Message)
// 	}

// 	var replicatedEmailResp clients.GetReplicatedCustomerResponse
// 	err := JsonUnmarshal(request.Body, &replicatedEmailResp)

// 	if err != nil {
// 		server.logCtx(c).Errorf("Error while unmarshalling getReplicatedCustomer response : %s", err.Error())
// 		return server.SendErrorResponse(c, http.StatusInternalServerError, constants.UNMARSHAL_ERR_MSG+err.Error())
// 	}

// 	if replicatedEmailResp.StatusCode != http.StatusOK {
// 		server.logCtx(c).Errorf("Received error response from getReplicatedCustomer")
// 		return server.SendErrorResponse(c, replicatedEmailResp.StatusCode, replicatedEmailResp.Message)

// 	}
// 	server.logCtx(c).Debug("Successfully fetched replicated customer email")

// 	//2. Run a search customer on replicated with email
// 	requestId := fmt.Sprint(c.Locals("requestid"))
// 	customers, err := server.Replicated.SearchCustomersByEmail(replicatedEmailResp.ReplicatedEmail, requestId)

// 	if err != nil {
// 		server.logCtx(c).Errorf("Error while fetching replicated customers with Email : %s", err.Error())
// 		return server.SendErrorResponse(c, http.StatusInternalServerError, constants.REPLICATED_CUSTOMER_ERROR)
// 	}

// 	if len(customers) == 0 {
// 		server.logCtx(c).Errorf("No replicated customers found with Email : %s", replicatedEmailResp.ReplicatedEmail)
// 		return server.SendErrorResponse(c, http.StatusInternalServerError, constants.REPLICATED_CUSTOMER_ERROR)
// 	}
// 	customer := customers[0]
// 	server.logCtx(c).Debug("Successfully fetched replicated customer details")

// 	//3. Based on Airgap flag, formulate the download URL
// 	url, err := server.Replicated.GetDowloadUrl(customer, requestId)
// 	if err != nil {
// 		return server.SendErrorResponse(c, http.StatusInternalServerError, constants.REPLICATED_DOWNLOAD_ERROR)
// 	}
// 	server.logCtx(c).Info("Successfully formulated download url")

// 	server.logCtx(c).Info("Performing download from replicated")
// 	downloadResp, err := server.Replicated.DownloadFromReplicated(url, requestId, customer.InstallationId)
// 	if err != nil {
// 		server.logCtx(c).Errorf("Error while downloading from replicated : %s", err.Error())
// 		return server.SendErrorResponse(c, http.StatusInternalServerError, constants.REPLICATED_DOWNLOAD_ERROR)
// 	}
// 	//	defer downloadResp.Body.Close()

// 	server.logCtx(c).Info("Successfully downloaded from replicated")

// 	// Set response headers
// 	for name, values := range downloadResp.Header {
// 		for _, value := range values {
// 			c.Set(name, value)
// 		}
// 	}

// 	// Set Headers
// 	c.Set(fiber.HeaderContentType, constants.OCTET_STREAM)
// 	c.Set(fiber.HeaderContentLength, downloadResp.Header.Get(fiber.HeaderContentLength))
// 	c.Set(fiber.HeaderContentDisposition, constants.PLATFORM_SERVICE_CONTENT_DISPOSITION)
// 	c.Set(fiber.HeaderTransferEncoding, constants.CHUNKED)

// 	c.Status(200).Context().SetBodyStreamWriter(func(w *bufio.Writer) {
// 		buf := make([]byte, 32*1024) // 32KB buffer
// 		for {
// 			n, err := downloadResp.Body.Read(buf)
// 			if err == io.EOF {
// 				break
// 			}
// 			if err != nil {
// 				server.Log.Errorf("Error while streaming : %s", err.Error())
// 				return
// 			}
// 			if n > 0 {
// 				if _, writeErr := w.Write(buf[:n]); writeErr != nil {
// 					server.Log.Errorf("Error while streaming : %s", writeErr.Error())
// 					if err := w.Flush(); err != nil {
// 						server.Log.Errorf("Error while streaming : %s", err.Error())
// 						break
// 					}
// 					return

// 				}
// 				// Explicitly flush the response
// 				if err := w.Flush(); err != nil {
// 					server.Log.Errorf("Error while streaming : %s", err.Error())
// 					break
// 				}
// 			}
// 		}
// 		defer downloadResp.Body.Close()
// 	})
// 	server.logCtx(c).Info("Successfully copied response. Returning response")
// 	return nil
// }

// func (server *ApiService) downloadAutomateOrHabitat(params *omnitruck.RequestParams, c *fiber.Ctx) error {
// 	url, err := server.DynamoServices(server.DatabaseService, c).ProductDownload(params)
// 	if err != nil {
// 		code, msg := services.GetErrorCodeAndMsg(err)
// 		return server.SendErrorResponse(c, code, msg)
// 	}
// 	server.logCtx(c).Infof("Redirecting user to %s", url)
// 	return c.Redirect(url, 302)
// }

// // @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// // @Param       bom    	   query string true  "bom"
// // @Param       license_id query string false "License ID"
// // @Success     200
// // @Failure     400 {object} services.ErrorResponse
// // @Failure     403 {object} services.ErrorResponse
// // @Router      /relatedProducts [get]
// func (server *ApiService) relatedProductsHandler(c *fiber.Ctx) error {
// 	params := getRequestParams(c)

// 	server.logCtx(c).Info("Validating related products API for " + params.BOM)

// 	err, ok := server.ValidateRequest(params, c)
// 	if !ok {
// 		server.logCtx(c).Error("Validation of related products API for "+params.BOM+"failed", err.Error())
// 		return err
// 	}

// 	relatedProducts, err := server.DynamoServices(server.DatabaseService, c).GetRelatedProducts(params)

// 	if err != nil {
// 		code, msg := getErrorCodeAndMsg(err)
// 		server.logCtx(c).Error("Error while fetching related products for "+params.BOM, err.Error())
// 		return server.SendErrorResponse(c, code, msg)
// 	}

// 	response := map[string]interface{}{
// 		"relatedProducts": relatedProducts.Products,
// 	}
// 	server.logCtx(c).Info("Returning success response from related products API for " + params.BOM)
// 	return server.SendResponse(c, response)
// }

// // @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// // @Param       channel    path     string true  "Channel"                                                                                                                      Enums(current, stable)
// // @Param       product    path     string true  "Product"                                                                                                                      Example(chef)
// // @Param       p          query    string true  "Platform, valid values are returned from the `/platforms` endpoint."                                                          Example(ubuntu)
// // @Param       pv         query    string true  "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." Example(20.04)
// // @Param       m          query    string true  "Machine architecture, valid values are returned by the `/architectures` endpoint."                                            Example(x86_64)
// // @Param       v          query    string false "Version of the product to be installed. A version always takes the form `x.y.z`"                                              Default(latest)
// // @Param       license_id query    string false "License ID"
// // @Success     200        {object} map[string]interface{}
// // @Failure     400        {object} services.ErrorResponse
// // @Failure     403        {object} services.ErrorResponse
// // @Router      /{channel}/{product}/fileName [get]
// func (server *ApiService) fileNameHandler(c *fiber.Ctx) error {
// 	params := getRequestParams(c)
// 	err, ok := server.ValidateRequest(params, c)
// 	if !ok {
// 		server.logCtx(c).Error("Validation of file name API for " + params.Product + " failed")
// 		return err
// 	}

// 	if params.Product == constants.PLATFORM_SERVICE_PRODUCT {
// 		fileName, err := server.PlatformServices(c).PlatformFilename(params, int(server.Mode))
// 		if err != nil {
// 			code, msg := getErrorCodeAndMsg(err)
// 			server.logCtx(c).Error("Error while fetching fileName for "+params.Product, err.Error())
// 			return server.SendErrorResponse(c, code, msg)
// 		}
// 		response := map[string]interface{}{
// 			"fileName": fileName,
// 		}
// 		server.logCtx(c).Info(constants.SUCCESS_RESPONSE_FROM_FILENAME_MSG + params.Product)
// 		return server.SendResponse(c, response)
// 	}

// 	server.logCtx(c).Info("Validating download file name for " + params.Product + " in channel " + params.Channel)
// 	err = server.versionCheckForTrialAndOsServer(params, c)
// 	if err != nil {
// 		code, msg := getErrorCodeAndMsg(err)
// 		return server.SendErrorResponse(c, code, msg)
// 	}

// 	//assuming that the metadata table will always have only the latest version record for automate, querying db without sortkey
// 	if params.Product == constants.AUTOMATE_PRODUCT || params.Product == constants.HABITAT_PRODUCT {
// 		fileName, err := server.DynamoServices(server.DatabaseService, c).GetFilename(params)
// 		if err != nil {
// 			code, msg := getErrorCodeAndMsg(err)
// 			server.logCtx(c).Error("Error while fetching fileName for "+params.Product, err.Error())
// 			return server.SendErrorResponse(c, code, msg)
// 		}

// 		response := map[string]interface{}{
// 			"fileName": fileName,
// 		}
// 		server.logCtx(c).Info(constants.SUCCESS_RESPONSE_FROM_FILENAME_MSG + params.Product)
// 		return server.SendResponse(c, response)

// 	} else {
// 		var data omnitruck.PackageMetadata
// 		request := server.Omnitruck(c).ProductMetadata(params).ParseData(&data)

// 		if request.Ok {
// 			url := data.Url
// 			fileName := getFileNameFromURL(url)
// 			response := map[string]interface{}{
// 				"fileName": fileName,
// 			}
// 			server.logCtx(c).Info(constants.SUCCESS_RESPONSE_FROM_FILENAME_MSG + params.Product)
// 			return server.SendResponse(c, response)
// 		} else {
// 			return server.SendError(c, request)
// 		}

// 	}
// }

// func getFileNameFromURL(url string) string {
// 	segments := strings.Split(url, "/")
// 	return segments[len(segments)-1]
// }

// func (server *ApiService) IsLatestForTrial(params *omnitruck.RequestParams, c *fiber.Ctx) *clients.Request {
// 	latestVersion, request := server.FetchLatestVersion(params, c)
// 	if params.Version == "latest" || params.Version == "" || params.Version == string(latestVersion) {
// 		request.Success()
// 		return request
// 	}
// 	if !request.Ok {
// 		return request
// 	}
// 	request.Failure(fiber.StatusBadRequest, "Version is not latest.")
// 	return request
// }

// func getErrorCodeAndMsg(err error) (code int, msg string) {
// 	var fiberErr *fiber.Error

// 	if errors.As(err, &fiberErr) {
// 		code = fiberErr.Code
// 		msg = fiberErr.Message
// 		return code, msg
// 	}
// 	return fiber.StatusInternalServerError, ""
// }

// func (server *ApiService) versionCheckForTrialAndOsServer(params *omnitruck.RequestParams, c *fiber.Ctx) error {
// 	if server.Mode == Trial {
// 		err := server.IsLatestForTrial(params, c)
// 		if !err.Ok {
// 			return fiber.NewError(err.Code, err.Message)
// 		}
// 	} else if server.Mode == Opensource {
// 		if isLatest(params.Version) {
// 			v, err := server.FetchLatestOSVersion(params, c)
// 			if !err.Ok {
// 				server.logCtx(c).Error("Error while fetching latest opensource version for the product ", params.Product, " error :- ", err.Message)
// 				return fiber.NewError(err.Code, err.Message)
// 			}
// 			params.Version = string(v)
// 		} else {
// 			err := server.IsOsVersion(params, c)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func (server *ApiService) isOsVersion(params *omnitruck.RequestParams, c *fiber.Ctx) error {
// 	var err error
// 	version := params.Version
// 	allversions := []omnitruck.ProductVersion{}

// 	errMsg := fmt.Sprintf(`Version %s not support on this persona.`, version)
// 	errLog := fmt.Sprintf(`Error while fetching all versions for the product %s. error :- `, params.Product)
// 	if params.Product == constants.HABITAT_PRODUCT || params.Product == constants.AUTOMATE_PRODUCT {
// 		allversions, err = server.DynamoServices(server.DatabaseService, c).VersionAll(params)
// 		if err != nil {
// 			server.logCtx(c).Error(errLog, err.Error())
// 			return fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
// 		}
// 	} else {
// 		request := server.Omnitruck(c).ProductVersions(params).ParseData(&allversions)
// 		if !request.Ok {
// 			server.logCtx(c).Error(errLog, request.Message)
// 			return fiber.NewError(request.Code, request.Message)
// 		}
// 	}

// 	if params.Product == constants.AUTOMATE_PRODUCT {
// 		if version == string(allversions[0]) {
// 			return nil
// 		}
// 		return fiber.NewError(fiber.StatusBadRequest, errMsg)
// 	}
// 	allversions = omnitruck.FilterList(allversions, func(v omnitruck.ProductVersion) bool {
// 		return !omnitruck.OsProductVersion(params.Product, v)
// 	})

// 	for _, val := range allversions {
// 		if val == omnitruck.ProductVersion(version) {
// 			return nil
// 		}
// 	}
// 	return fiber.NewError(fiber.StatusBadRequest, errMsg)
// }

// // @description The `ACCEPT` HTTP header with a value of `application/x-sh` must be provided in the request for a shell script response to be returned
// // @Param       license_id query    string false "License ID"
// // @Success     200        {object} map[string]interface{}
// // @Failure     403        {object} services.ErrorResponse
// // @Failure     500        {object} services.ErrorResponse
// // @Router      /install.sh [get]
// func (server *ApiService) downloadLinuxScript(c *fiber.Ctx) error {
// 	params := getRequestParams(c)
// 	c.Set("Content-Type", "application/x-sh")
// 	err, ok := server.ValidateRequest(params, c)
// 	if !ok {
// 		server.logCtx(c).Error("Validation of download linux script API failed: ", err)
// 		return err
// 	}
// 	if server.Mode == Opensource {
// 		params.LicenseId = ""
// 	}
// 	filePath := "../templates/install.sh.tmpl"
// 	resp, err := server.TemplateRenderer.GetScript(c.BaseURL(), params, filePath)
// 	if err != nil {
// 		return err
// 	}
// 	c.Set("Content-Disposition", "attachment;filename=install.sh")
// 	return c.SendString(resp)
// }

// // @description The `ACCEPT` HTTP header with a value of `text/plain` must be provided in the request for a text response to be returned
// // @Param       license_id query    string false "License ID"
// // @Success     200        {object} map[string]interface{}
// // @Failure     403        {object} services.ErrorResponse
// // @Failure     500        {object} services.ErrorResponse
// // @Router      /install.ps1 [get]
// func (server *ApiService) downloadWindowsScript(c *fiber.Ctx) error {
// 	params := getRequestParams(c)
// 	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
// 	err, ok := server.ValidateRequest(params, c)
// 	if !ok {
// 		server.logCtx(c).Error("Validation of download windows script API failed: ", err)
// 		return err
// 	}
// 	if server.Mode == Opensource {
// 		params.LicenseId = ""
// 	}
// 	filePath := "../templates/install.ps1.tmpl"
// 	scriptResp, err := server.TemplateRenderer.GetScript(c.BaseURL(), params, filePath)
// 	if err != nil {
// 		return err
// 	}
// 	c.Set("Content-Disposition", "attachment;filename=install.ps1")
// 	resp := html.UnescapeString(scriptResp)
// 	return c.SendString(resp)
// }

// // @description Get the list of available package managers
// // @Success 200 {object} map[string]interface{}
// // @Failure     500 {object} services.ErrorResponse
// // @Router /package-managers [get]
// func (server *ApiService) packageManagersHandler(c *fiber.Ctx) error {
// 	server.logCtx(c).Info("Fetching package managers")

// 	packageManagers, err := server.DynamoServices(server.DatabaseService, c).GetPackageManagers()
// 	if err != nil {
// 		code, msg := getErrorCodeAndMsg(err)
// 		return server.SendErrorResponse(c, code, msg)
// 	}

// 	return server.SendResponse(c, packageManagers)
// }
