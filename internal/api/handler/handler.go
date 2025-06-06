package handler

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	helpers "github.com/chef/omnitruck-service/internal/helper"
	"github.com/chef/omnitruck-service/internal/services"
	"github.com/gofiber/fiber/v2"
)

type DownloadsHandler struct {
}

// NewDownloadsHandler creates a new instance of DownloadsHandler
func NewDownloadsHandler() *DownloadsHandler {
	return &DownloadsHandler{}
}

type ErrorResponse struct {
	Code       int    `json:"code"`
	StatusText string `json:"status_text"`
	Message    string `json:"message"`
} //@name ErrorResponse

func (h *DownloadsHandler) JSON(c *fiber.Ctx, data interface{}) error {
	var resultBytes bytes.Buffer
	enc := json.NewEncoder(&resultBytes)
	enc.SetEscapeHTML(false)
	err := enc.Encode(data)
	c.Context().Response.SetBodyRaw(resultBytes.Bytes())
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return err
}

func (h *DownloadsHandler) SendResponse(c *fiber.Ctx, data interface{}) error {
	return h.JSON(c, data)
}

func (h *DownloadsHandler) SendError(c *fiber.Ctx, request *clients.Request) error {

	return c.Status(request.Code).JSON(ErrorResponse{
		Code:       request.Code,
		StatusText: http.StatusText(request.Code),
		Message:    request.Message,
	})
}

func (h *DownloadsHandler) SendErrorResponse(c *fiber.Ctx, code int, msg string) error {
	return c.Status(code).JSON(ErrorResponse{
		Code:       code,
		StatusText: http.StatusText(code),
		Message:    msg,
	})
}

func (h *DownloadsHandler) ProductsHandler(c *fiber.Ctx) error {
	params := helpers.GetRequestParams(c)
	downloadService := services.NewDownloadService(c)
	data, request := downloadService.Products(params)

	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
}

// @description Returns a valid list of valid platform keys along with full friendly names.
// @description Any of these platform keys can be used in the p query string value in various endpoints below.
// @Success     200 {object} omnitruck.PlatformList
// @Failure     500 {object} services.ErrorResponse
// @Router      /platforms [get]
func (h *DownloadsHandler) PlatformsHandler(c *fiber.Ctx) error {
	var data omnitruck.PlatformList
	downloadService := services.NewDownloadService(c)
	data, request := downloadService.Platforms()

	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
}

// @description Returns a valid list of valid platform keys along with friendly names.
// @description Any of these architecture keys can be used in the p query string value in various endpoints below.
// @Success     200 {object} omnitruck.ItemList
// @Failure     500 {object} services.ErrorResponse
// @Router      /architectures [get]
func (h *DownloadsHandler) ArchitecturesHandler(c *fiber.Ctx) error {

	downloadService := services.NewDownloadService(c)
	data, request := downloadService.Architectures()

	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
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
func (h *DownloadsHandler) LatestVersionHandler(c *fiber.Ctx) error {
	params := helpers.GetRequestParams(c)
	params.Version = "latest"

	downloadService := services.NewDownloadService(c)
	data, request := downloadService.LatestVersion(params)

	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
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
func (h *DownloadsHandler) ProductVersionsHandler(c *fiber.Ctx) error {
	params := helpers.GetRequestParams(c)
	downloadService := services.NewDownloadService(c)

	data, request := downloadService.ProductVersions(params)
	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
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
func (h *DownloadsHandler) ProductPackagesHandler(c *fiber.Ctx) error {
	params := helpers.GetRequestParams(c)
	downloadService := services.NewDownloadService(c)

	data, request := downloadService.ProductPackages(params)
	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
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
func (h *DownloadsHandler) ProductMetadataHandler(c *fiber.Ctx) error {
	params := helpers.GetRequestParams(c)
	downloadService := services.NewDownloadService(c)

	data, request := downloadService.ProductMetadata(params)
	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
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
func (h *DownloadsHandler) ProductDownloadHandler(c *fiber.Ctx) error {
	// params := helpers.GetRequestParams(c)
	// downloadService := services.NewDownloadService(c)
	// // request := downloadService.ProductDownload(params)
	// // if request.Ok {
	// // 	return c.Redirect(request.Url, http.StatusFound)
	// // } else {
	// // 	return h.SendError(c, request)
	// // }

	return nil
}

// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Param       bom    	   query string true  "bom"
// @Param       license_id query string false "License ID"
// @Success     200
// @Failure     400 {object} services.ErrorResponse
// @Failure     403 {object} services.ErrorResponse
// @Router      /relatedProducts [get]
func (h *DownloadsHandler) RelatedProductsHandler(c *fiber.Ctx) error {
	params := helpers.GetRequestParams(c)
	downloadService := services.NewDownloadService(c)
	data, request := downloadService.RelatedProducts(params)
	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
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
func (h *DownloadsHandler) FileNameHandler(c *fiber.Ctx) error {
	params := helpers.GetRequestParams(c)
	downloadService := services.NewDownloadService(c)
	fileName, request := downloadService.GetFileName(params)
	if request.Ok {
		return h.SendResponse(c, map[string]interface{}{
			"file_name": fileName,
		})
	} else {
		return h.SendError(c, request)
	}
}

// @description The `ACCEPT` HTTP header with a value of `application/x-sh` must be provided in the request for a shell script response to be returned
// @Param       license_id query    string false "License ID"
// @Success     200        {object} map[string]interface{}
// @Failure     403        {object} services.ErrorResponse
// @Failure     500        {object} services.ErrorResponse
// @Router      /install.sh [get]
func (h *DownloadsHandler) DownloadLinuxScript(c *fiber.Ctx) error {
	params := helpers.GetRequestParams(c)
	c.Set("Content-Type", "application/x-sh")
	c.Set("Content-Disposition", "attachment;filename=install.sh")
	downloadService := services.NewDownloadService(c)
	resp, request := downloadService.GetLinuxScript(params)
	if !request.Ok {
		return h.SendError(c, request)
	} else {
		return c.SendString(resp)
	}
}

// @description The `ACCEPT` HTTP header with a value of `text/plain` must be provided in the request for a text response to be returned
// @Param       license_id query    string false "License ID"
// @Success     200        {object} map[string]interface{}
// @Failure     403        {object} services.ErrorResponse
// @Failure     500        {object} services.ErrorResponse
// @Router      /install.ps1 [get]
func (h *DownloadsHandler) DownloadWindowsScript(c *fiber.Ctx) error {
	params := helpers.GetRequestParams(c)
	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	c.Set("Content-Disposition", "attachment;filename=install.ps1")

	downloadService := services.NewDownloadService(c)
	resp, request := downloadService.GetWindowsScript(params)
	if !request.Ok {
		return h.SendError(c, request)
	} else {
		return c.SendString(resp)
	}
}
