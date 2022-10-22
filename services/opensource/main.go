package opensource

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	_ "github.com/chef/omnitruck-service/docs/opensource"
	omnitruck "github.com/chef/omnitruck-service/omnitruck-client"
	rv "github.com/chef/omnitruck-service/request_validators"
	"github.com/chef/omnitruck-service/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type OpensourceService struct {
	sync.Mutex
	config    services.Config
	validator rv.OpensourceValidator
	omnitruck omnitruck.Omnitruck
	log       *logrus.Entry
	app       *fiber.App
}

type ErrorResponse struct {
	Code       int    `json:"code" example:200`
	StatusText string `json:"status_text" example:OK`
	Message    string `json:"message"`
}

// RequestParams is used to setup validation for the request parameters
type RequestParams struct {
	Channel         string `validate:"required,eq=stable"`
	Product         string `validate:"required"`
	Version         string
	Platform        string `validate:"required_with=PlatformVersion"`
	PlatformVersion string `validate:"required_with=Platform"`
	Architecture    string `validate:"required_with_all=PlatformVersion Platform"`
}

// Because the params objects gets passed back to the omnitruck client as an interface object
// we need to create a getter to fetch the data out of it
//
// TODO: Figure out if there is a better way to implement this so we don't need the getter method
func (rp *RequestParams) Get(name string) string {
	switch name {
	case "channel":
		return rp.Channel
	case "product":
		return rp.Product
	case "version":
		return rp.Version
	case "platform":
		return rp.Platform
	case "platformVersion":
		return rp.PlatformVersion
	case "architecture":
		return rp.Architecture
	default:
		return ""
	}
}

// @title			Licensed Omnitruck API for opensource products
// @version			1.0
// @description 	Licensed Omnitruck API for opensource products
// @license.name	Apache 2.0
// @license.url 	http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:3000
func NewOpensourceServer(c services.Config) *OpensourceService {
	return &OpensourceService{
		validator: rv.NewOpensourceValidator(),
		omnitruck: omnitruck.NewOmnitruckClient(),
		log:       log.WithField("pkg", "OpensourceService"),
		config:    c,
	}
}

func (server *OpensourceService) Start(wg *sync.WaitGroup) error {
	server.Lock()
	defer server.Unlock()

	server.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		EnablePrintRoutes:     false,
		ReadTimeout:           300 * time.Second,
		WriteTimeout:          300 * time.Second,
	})

	server.app.Use(cors.New())
	server.app.Use(recover.New())

	wg.Add(1)
	go server.startOpensourceService()

	return nil
}

func (server *OpensourceService) HealthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	return c.JSON(res)
}

func (server *OpensourceService) Name() string {
	return "OpensourceService"
}

func (server *OpensourceService) sendResponse(c *fiber.Ctx, success bool, code int, body omnitruck.ResponseInterface) error {
	if !success {
		errorMsg := fmt.Sprintf("%v", body)
		// If we aren't given a custom message for the error
		// then pull a standard one for the http code
		if len(errorMsg) == 0 {
			errorMsg = http.StatusText(code)
		}

		return c.Status(code).JSON(ErrorResponse{
			Code:       code,
			StatusText: http.StatusText(code),
			Message:    errorMsg,
		})
	}

	return c.JSON(body)
}

func (server *OpensourceService) ValidateRequest(params *RequestParams, c *fiber.Ctx) (error, bool) {
	errors := server.validator.ValidateParams(params)
	if errors != nil {
		msgs, code := server.validator.ErrorMessages(errors)

		server.log.WithField("errors", msgs).Error("Error validating request")
		return c.Status(code).JSON(ErrorResponse{
			Code:       code,
			StatusText: http.StatusText(code),
			Message:    msgs,
		}), false
	}
	return nil, true
}

// @description Returns a valid list of valid product keys.
// @description Any of these product keys can be used in the <PRODUCT> value of other endpoints. Please note many of these products are used for internal tools only and many have been EOL’d.
// @Success 200 {object} omnitruck.ProductList
// @Failure 500 {object} opensource.ErrorResponse
// @Router /products [get]
func (server *OpensourceService) productsHandler(c *fiber.Ctx) error {
	code, body, success := server.omnitruck.Products()
	return server.sendResponse(c, success, code, body)
}

// @description Returns a valid list of valid platform keys along with full friendly names.
// @description Any of these platform keys can be used in the p query string value in various endpoints below.
// @Success 200 {object} omnitruck.PlatformList
// @Failure 500 {object} opensource.ErrorResponse
// @Router /platforms [get]
func (server *OpensourceService) platformsHandler(c *fiber.Ctx) error {
	code, body, success := server.omnitruck.Platforms()
	return server.sendResponse(c, success, code, body)
}

// @description Returns a valid list of valid platform keys along with friendly names.
// @description Any of these architecture keys can be used in the p query string value in various endpoints below.
// @Success 200 {object} omnitruck.ArchitectureList
// @Failure 500 {object} opensource.ErrorResponse
// @Router /architectures [get]
func (server *OpensourceService) architecturesHandler(c *fiber.Ctx) error {
	code, body, success := server.omnitruck.Architectures()
	return server.sendResponse(c, success, code, body)
}

// @description Get the latest version number for a particular channel and product combination.
// @Param channel 		path 	string 	true 	"Channel" Enums(current, stable)
// @Param product   	path 	string 	true 	"Product"
// @Param license_id 	header 	string 	false 	"License ID"
// @Param eol			query 	bool 	false 	"EOL Products"
// @Success 200 {object} omnitruck.ProductVersion
// @Failure 400 {object} opensource.ErrorResponse
// @Failure 403 {object} opensource.ErrorResponse
// @Router /{channel}/{product}/versions/latest [get]
func (server *OpensourceService) latestVersionHandler(c *fiber.Ctx) error {
	params := &RequestParams{
		Channel: c.Params("channel"),
		Product: c.Params("product"),
	}
	err, ok := server.ValidateRequest(params, c)
	if err != nil || !ok {
		return err
	}

	code, body, success := server.omnitruck.LatestVersion(params)

	return server.sendResponse(c, success, code, body)
}

// @description Get a list of all available version numbers for a particular channel and product combination
// @Param channel 		path 	string 	true 	"Channel" Enums(current, stable)
// @Param product   	path 	string 	true 	"Product"
// @Param license_id 	header 	string 	false 	"License ID"
// @Param eol			query 	bool 	false 	"EOL Products" Default(false)
// @Success 200 {object} omnitruck.VersionList
// @Failure 400 {object} opensource.ErrorResponse
// @Failure 403 {object} opensource.ErrorResponse
// @Router /{channel}/{product}/versions/all [get]
func (server *OpensourceService) productVersionsHandler(c *fiber.Ctx) error {
	params := &RequestParams{
		Channel: c.Params("channel"),
		Product: c.Params("product"),
	}
	err, ok := server.ValidateRequest(params, c)
	if err != nil || !ok {
		return err
	}

	code, body, success := server.omnitruck.ProductVersions(params)

	return server.sendResponse(c, success, code, body)
}

// @description Get the full list of all packages for a particular channel and product combination.
// @description By default all packages for the latest version are returned. If the v query string parameter is included the packages for the specified version are returned.
// @Param channel 		path 	string 	true 	"Channel" Enums(current, stable)
// @Param product   	path 	string 	true 	"Product" Example(chef)
// @Param v				query	string	false	"Version"
// @Param license_id 	header 	string 	false 	"License ID"
// @Param eol			query 	bool 	false 	"EOL Products" Default(false)
// @Success 200 {object} omnitruck.PackageList
// @Failure 400 {object} opensource.ErrorResponse
// @Failure 403 {object} opensource.ErrorResponse
// @Router /{channel}/{product}/packages [get]
func (server *OpensourceService) productPackagesHandler(c *fiber.Ctx) error {
	params := &RequestParams{
		Channel: c.Params("channel"),
		Product: c.Params("product"),
		Version: c.Query("v"),
	}
	err, ok := server.ValidateRequest(params, c)
	if err != nil || !ok {
		return err
	}

	code, body, success := server.omnitruck.ProductPackages(params)

	return server.sendResponse(c, success, code, body)
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
// @Failure 400 {object} opensource.ErrorResponse
// @Failure 403 {object} opensource.ErrorResponse
// @Router /{channel}/{product}/metadata [get]
func (server *OpensourceService) productMetadataHandler(c *fiber.Ctx) error {
	params := &RequestParams{
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

	code, body, success := server.omnitruck.ProductMetadata(params)

	return server.sendResponse(c, success, code, body)
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
// @Failure 400 {object} opensource.ErrorResponse
// @Failure 403 {object} opensource.ErrorResponse
// @Router /{channel}/{product}/download [get]
func (server *OpensourceService) productDownloadHandler(c *fiber.Ctx) error {
	params := &RequestParams{
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

	code, body, success := server.omnitruck.ProductDownload(params)

	switch v := body.(type) {
	case omnitruck.PackageMetadata:
		if success {
			server.log.Infof("Redirecting user to %s", v.Url)
			return c.Redirect(v.Url, 302)
		}
	}

	return server.sendResponse(c, success, code, body)
}

func (server *OpensourceService) buildRouter(lw io.Writer) {
	server.app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
		Output: lw,
	}))

	server.app.Get("/swagger/*", swagger.HandlerDefault)

	server.app.Get("/", server.HealthCheck)
	server.app.Get("/products", server.productsHandler)
	server.app.Get("/platforms", server.platformsHandler)
	server.app.Get("/architectures", server.architecturesHandler)
	server.app.Get("/:channel/:product/versions/latest", server.latestVersionHandler)
	server.app.Get("/:channel/:product/versions/all", server.productVersionsHandler)
	server.app.Get("/:channel/:product/packages", server.productPackagesHandler)
	server.app.Get("/:channel/:product/metadata", server.productMetadataHandler)
	server.app.Get("/:channel/:product/download", server.productDownloadHandler)
}

func (server *OpensourceService) startOpensourceService() {
	// Setup io writer for the logger
	lw := server.log.Writer()
	defer lw.Close()

	server.buildRouter(lw)

	server.log.Infof("Starting %s server at: %s", server.Name(), server.config.Listen)
	err := server.app.Listen(server.config.Listen)
	if err != nil {
		if err == http.ErrServerClosed {
			server.log.WithError(err).Error("unable to start OpensourceService")
		} else {
			server.log.WithError(err).Fatal("OpensourceService stopped")
		}
	}
}
