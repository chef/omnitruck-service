package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
	helpers "github.com/chef/omnitruck-service/internal/helper"
	"github.com/chef/omnitruck-service/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do"
	log "github.com/sirupsen/logrus"
)

type DownloadsHandler struct {
	Log *log.Entry
}

// NewDownloadsHandler creates a new instance of DownloadsHandler
func NewDownloadsHandler(log *log.Entry) *DownloadsHandler {
	return &DownloadsHandler{
		Log: log,
	}
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

// @Summary Get list of available products
// @description Returns a valid list of valid product keys.
// @description Any of these product keys can be used in the <PRODUCT> value of other endpoints. Please note many of these products are used for internal tools only and many have been EOL'd.
// @Accept      json
// @Produce     json
// @Param       eol query    bool false "EOL Products"
// @Success     200 {object} omnitruck.ItemList
// @Failure     500 {object} ErrorResponse
// @Router      /products [get]
func (h *DownloadsHandler) ProductsHandler(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	params := helpers.GetRequestParams(c)
	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	data, request := downloadService.Products(params)

	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
}

// @Summary Get platform keys
// @description Returns a valid list of valid platform keys along with full friendly names.
// @description Any of these platform keys can be used in the p query string value in various endpoints below.
// @Accept      json
// @Produce     json
// @Success     200 {object} omnitruck.PlatformList
// @Failure     500 {object} ErrorResponse
// @Router      /platforms [get]
func (h *DownloadsHandler) PlatformsHandler(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	var data omnitruck.PlatformList
	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	data, request := downloadService.Platforms()

	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
}

// @Summary Get architecture keys
// @description Returns a valid list of valid platform keys along with friendly names.
// @description Any of these architecture keys can be used in the p query string value in various endpoints below.
// @Accept      json
// @Produce     json
// @Success     200 {object} omnitruck.ItemList
// @Failure     500 {object} ErrorResponse
// @Router      /architectures [get]
func (h *DownloadsHandler) ArchitecturesHandler(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	data, request := downloadService.Architectures()

	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
}

// @Summary Get latest version of a product in a channel
// @description Get the latest version number for a particular channel and product combination.
// @Accept      json
// @Produce     json
// @Param       channel    path     string true  "Channel" Enums(current, stable)
// @Param       product    path     string true  "Product"
// @Param       license_id query    string false "License ID"
// @Success     200        {object} omnitruck.ProductVersion
// @Failure     400        {object} ErrorResponse
// @Failure     403        {object} ErrorResponse
// @Router      /{channel}/{product}/versions/latest [get]
func (h *DownloadsHandler) LatestVersionHandler(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	params := helpers.GetRequestParams(c)
	msg, code, ok := h.ValidateRequest(params, c)
	if !ok {
		return h.SendErrorResponse(c, code, msg)
	}
	params.Version = "latest"

	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	data, request := downloadService.LatestVersion(params)

	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
}

// @Summary Get all versions of a product in a channel
// @description Get a list of all available version numbers for a particular channel and product combination
// @Accept      json
// @Produce     json
// @Param       channel    path     string true  "Channel" Enums(current, stable)
// @Param       product    path     string true  "Product"
// @Param       license_id query    string false "License ID"
// @Param       eol        query    bool   false "EOL Products" Default(false)
// @Success     200        {object} omnitruck.ItemList
// @Failure     400        {object} ErrorResponse
// @Failure     403        {object} ErrorResponse
// @Router      /{channel}/{product}/versions/all [get]
func (h *DownloadsHandler) ProductVersionsHandler(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	params := helpers.GetRequestParams(c)
	msg, code, ok := h.ValidateRequest(params, c)
	if !ok {
		return h.SendErrorResponse(c, code, msg)
	}
	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	data, request := downloadService.ProductVersions(params)
	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
}

// @Summary Get packages for a product version
// @description Get the full list of all packages for a particular channel and product combination.
// @description By default all packages for the latest version are returned. If the v query string parameter is included the packages for the specified version are returned.
// @Accept      json
// @Produce     json
// @Param       channel    path     string true  "Channel" Enums(current, stable)
// @Param       product    path     string true  "Product" Example(chef)
// @Param       v          query    string false "Version"
// @Param       license_id query    string false "License ID"
// @Param       eol        query    bool   false "EOL Products" Default(false)
// @Success     200        {object} omnitruck.PackageList
// @Failure     400        {object} ErrorResponse
// @Failure     403        {object} ErrorResponse
// @Router      /{channel}/{product}/packages [get]
func (h *DownloadsHandler) ProductPackagesHandler(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	params := helpers.GetRequestParams(c)
	msg, code, ok := h.ValidateRequest(params, c)
	if !ok {
		return h.SendErrorResponse(c, code, msg)
	}
	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	data, request := downloadService.ProductPackages(params)
	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
}

// @Summary Get metadata for a product
// @description Get details for a particular package.
// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Accept      json
// @Produce     json
// @Param       channel    path     string true  "Channel"                                                                                                                      Enums(current, stable)
// @Param       product    path     string true  "Product"                                                                                                                      Example(chef)
// @Param       p          query    string true  "Platform, valid values are returned from the `/platforms` endpoint."                                                          Example(ubuntu)
// @Param       pv         query    string true  "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." Example(20.04)
// @Param       m          query    string true  "Machine architecture, valid values are returned by the `/architectures` endpoint."                                            Example(x86_64)
// @Param       pm         query    string true  "Package Manager, valid values depend on the platform (e.g., Linux: deb, tar; Windows: msi)."                                  Example(tar)
// @Param       v          query    string false "Version of the product to be installed. A version always takes the form `x.y.z`"                                              Default(latest)
// @Param       license_id query    string false "License ID"
// @Param       eol        query    bool   false "EOL Products" Default(false)
// @Success     200        {object} omnitruck.PackageMetadata
// @Failure     400        {object} ErrorResponse
// @Failure     403        {object} ErrorResponse
// @Router      /{channel}/{product}/metadata [get]
func (h *DownloadsHandler) ProductMetadataHandler(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	params := helpers.GetRequestParams(c)
	msg, code, ok := h.ValidateRequest(params, c)
	if !ok {
		return h.SendErrorResponse(c, code, msg)
	}
	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	data, request := downloadService.ProductMetadata(params)
	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
}

// @Summary Download a product package
// @description Get details for a particular package.
// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Accept      json
// @Produce     json
// @Param       channel    path   string true  "Channel"                                                                                                                      Enums(current, stable)
// @Param       product    path   string true  "Product"                                                                                                                      Example(chef)
// @Param       p          query  string true  "Platform, valid values are returned from the `/platforms` endpoint."                                                          Example(ubuntu)
// @Param       pv         query  string true  "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." Example(20.04)
// @Param       m          query  string true  "Machine architecture, valid values are returned by the `/architectures` endpoint."                                            Example(x86_64)
// @Param       pm         query  string true  "Package Manager, valid values depend on the platform (e.g., Linux: deb, tar; Windows: msi)."                                  Example(tar)
// @Param       v          query  string false "Version of the product to be installed. A version always takes the form `x.y.z`"                                              Default(latest)
// @Param       license_id query  string false "License ID"
// @Param       eol        query  bool   false "EOL Products" Default(false)
// @Success     302
// @Failure     400 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Router      /{channel}/{product}/download [get]
func (h *DownloadsHandler) ProductDownloadHandler(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	params := helpers.GetRequestParams(c)
	msg, code, ok := h.ValidateRequest(params, c)
	if !ok {
		return h.SendErrorResponse(c, code, msg)
	}
	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	url, downloadResp, header, msg, code, err := downloadService.ProductDownload(params, c)
	if err != nil {
		return h.SendErrorResponse(c, code, msg)
	}
	if downloadResp != nil {
		// If the response is not nil, it means we are returning a file download

		// Set response headers
		for name, values := range header {
			for _, value := range values {
				c.Set(name, value)
			}
		}

		// Set Headers
		c.Set(fiber.HeaderContentType, constants.OCTET_STREAM)
		c.Set(fiber.HeaderContentLength, header.Get(fiber.HeaderContentLength))
		c.Set(fiber.HeaderTransferEncoding, constants.CHUNKED)

		c.Status(200).Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			buf := make([]byte, 32*1024) // 32KB buffer
			for {
				n, err := downloadResp.Read(buf)
				if n > 0 {
					if _, writeErr := w.Write(buf[:n]); writeErr != nil {
						h.Log.Errorf("Error while streaming : %s", writeErr.Error())
						w.Flush()
						break
					}
					if err := w.Flush(); err != nil {
						h.Log.Errorf("Error while streaming : %s", err.Error())
						break
					}
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					h.Log.Errorf("Error while streaming : %s", err.Error())
					break
				}
			}
			defer downloadResp.Close()
		})
		h.Log.Info("Successfully copied response. Returning response")
		return nil
	}
	if url != "" {
		// If the URL is not empty, we redirect to the download URL
		return c.Redirect(url, 302)
	}
	// If both URL and downloadResp are nil, we return an error
	return h.SendErrorResponse(c, http.StatusInternalServerError, "No download URL or response available")
}

// @Summary Get related products from a BOM
// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Accept      json
// @Produce     json
// @Param       bom    	   query string true  "bom"
// @Param       license_id query string false "License ID"
// @Success     200
// @Failure     400 {object} ErrorResponse
// @Failure     403 {object} ErrorResponse
// @Router      /relatedProducts [get]
func (h *DownloadsHandler) RelatedProductsHandler(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	params := helpers.GetRequestParams(c)
	msg, code, ok := h.ValidateRequest(params, c)
	if !ok {
		return h.SendErrorResponse(c, code, msg)
	}
	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	data, request := downloadService.RelatedProducts(params)
	if request.Ok {
		return h.SendResponse(c, &data)
	} else {
		return h.SendError(c, request)
	}
}

// @Summary Get filename for a product package
// @description The `ACCEPT` HTTP header with a value of `application/json` must be provided in the request for a JSON response to be returned
// @Accept      json
// @Produce     json
// @Param       channel    path     string true  "Channel"                                                                                                                      Enums(current, stable)
// @Param       product    path     string true  "Product"                                                                                                                      Example(chef)
// @Param       p          query    string true  "Platform, valid values are returned from the `/platforms` endpoint."                                                          Example(ubuntu)
// @Param       pv         query    string true  "Platform Version, possible values depend on the platform. For example, Ubuntu: 16.04, or 18.04 or for macOS: 10.14 or 10.15." Example(20.04)
// @Param       m          query    string true  "Machine architecture, valid values are returned by the `/architectures` endpoint."                                            Example(x86_64)
// @Param       pm         query    string true  "Package Manager, valid values depend on the platform (e.g., Linux: deb, tar; Windows: msi)."								    Example(tar)
// @Param       v          query    string false "Version of the product to be installed. A version always takes the form `x.y.z`"                                              Default(latest)
// @Param       license_id query    string false "License ID"
// @Success     200        {object} map[string]interface{}
// @Failure     400        {object} ErrorResponse
// @Failure     403        {object} ErrorResponse
// @Router      /{channel}/{product}/fileName [get]
func (h *DownloadsHandler) FileNameHandler(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	params := helpers.GetRequestParams(c)
	msg, code, ok := h.ValidateRequest(params, c)
	if !ok {
		return h.SendErrorResponse(c, code, msg)
	}
	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	fileName, request := downloadService.GetFileName(params)
	if request.Ok {
		return h.SendResponse(c, map[string]interface{}{
			"fileName": fileName,
		})
	} else {
		return h.SendError(c, request)
	}
}

// @Summary Download install script for Linux
// @description The `ACCEPT` HTTP header with a value of `application/x-sh` must be provided in the request for a shell script response to be returned
// @Accept      json
// @Produce     json
// @Param       license_id query    string false "License ID"
// @Success     200        {object} map[string]interface{}
// @Failure     403        {object} ErrorResponse
// @Failure     500        {object} ErrorResponse
// @Router      /install.sh [get]
func (h *DownloadsHandler) DownloadLinuxScript(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	params := helpers.GetRequestParams(c)
	msg, code, ok := h.ValidateRequest(params, c)
	if !ok {
		return h.SendErrorResponse(c, code, msg)
	}
	c.Set("Content-Type", "application/x-sh")
	c.Set("Content-Disposition", "attachment;filename=install.sh")
	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	resp, request := downloadService.GetLinuxScript(params)
	if !request.Ok {
		return h.SendError(c, request)
	} else {
		return c.SendString(resp)
	}
}

// @Summary Download install script for Windows
// @description The `ACCEPT` HTTP header with a value of `text/plain` must be provided in the request for a text response to be returned
// @Accept      json
// @Produce     json
// @Param       license_id query    string false "License ID"
// @Success     200        {object} map[string]interface{}
// @Failure     403        {object} ErrorResponse
// @Failure     500        {object} ErrorResponse
// @Router      /install.ps1 [get]
func (h *DownloadsHandler) DownloadWindowsScript(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	params := helpers.GetRequestParams(c)
	msg, code, ok := h.ValidateRequest(params, c)
	if !ok {
		return h.SendErrorResponse(c, code, msg)
	}
	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	c.Set("Content-Disposition", "attachment;filename=install.ps1")

	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	resp, request := downloadService.GetWindowsScript(params)
	if !request.Ok {
		return h.SendError(c, request)
	} else {
		return c.SendString(resp)
	}
}

// @Summary Get available package managers
// @description Get the list of available package managers
// @Accept      json
// @Produce     json
// @Success 200 {object} map[string]interface{}
// @Failure     500 {object} ErrorResponse
// @Router /package-managers [get]
func (h *DownloadsHandler) PackageManagersHandler(c *fiber.Ctx) error {
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Not able to process the request.")
	}
	locals := setLocals(c)
	downloadService, err := services.NewDownloadService(reqInjector, h.Log, locals)
	if err != nil {
		return h.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create download service")
	}
	packageManagers, request := downloadService.GetPackageManagers()
	if !request.Ok {
		return h.SendError(c, request)
	} else {
		return h.SendResponse(c, packageManagers)
	}
}

func (h *DownloadsHandler) ValidateRequest(params *omnitruck.RequestParams, c *fiber.Ctx) (string, int, bool) {
	context := omnitruck.Context{
		License: h.validLicense(c),
	}

	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return "Failed to retrieve request injector", fiber.StatusInternalServerError, false
	}

	validator := do.MustInvokeNamed[omnitruck.IRequestValidator](reqInjector, "validator")

	errors := validator.Params(params, context)
	if errors != nil {
		msgs, code := validator.ErrorMessages(errors)

		return msgs, code, false
	}

	return "", 0, true
}

func (h *DownloadsHandler) validLicense(c *fiber.Ctx) bool {
	v := c.Locals("valid_license")
	return v != nil && v.(bool)
}

func setLocals(c *fiber.Ctx) map[string]interface{} {
	locals := map[string]interface{}{}
	if c.Locals("valid_license") != nil {
		requestId := c.Locals("valid_license").(bool)
		locals["valid_license"] = requestId

	} else {
		locals["valid_license"] = false
	}

	if c.Locals("requestid") != nil {
		requestId := c.Locals("requestid").(string)
		locals["requestid"] = requestId

	} else {
		locals["requestid"] = ""
	}
	if c.Locals("base_url") != nil {
		baseUrl := c.Locals("base_url").(string)
		locals["base_url"] = baseUrl
	} else {
		locals["base_url"] = c.BaseURL()
	}
	if c.Locals("license_id") != nil {
		licenseId := c.Locals("license_id").(string)
		locals["license_id"] = licenseId
	} else {
		locals["license_id"] = ""
	}
	return locals
}
