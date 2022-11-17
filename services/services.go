package services

import (
	"net/http"
	"sync"
	"time"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
)

type ApiType int

const (
	Trial ApiType = iota
	Opensource
	Commercial
)

type ErrorResponse struct {
	Code       int    `json:"code" example:200`
	StatusText string `json:"status_text" example:OK`
	Message    string `json:"message"`
}

type Config struct {
	Name   string
	Listen string
	Log    *log.Entry
	Mode   ApiType
}

type Service interface {
	Name() string
	Start(*sync.WaitGroup) error
	Stop() error
}

type ApiService struct {
	sync.Mutex
	Config    Config
	Omnitruck omnitruck.Omnitruck
	Log       *log.Entry
	App       *fiber.App
	Validator omnitruck.RequestValidator
	Mode      ApiType
}

func (server *ApiService) Initialize(c Config) *ApiService {
	server.Omnitruck = omnitruck.NewOmnitruckClient()
	server.Log = c.Log
	server.Config = c
	server.Validator = omnitruck.NewValidator()
	server.Mode = c.Mode

	server.App = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		EnablePrintRoutes:     false,
		ReadTimeout:           300 * time.Second,
		WriteTimeout:          300 * time.Second,
	})

	server.App.Use(cors.New())
	// This will catch panics in the app and prevent it from crashing the server
	// TODO: Figure out if we can better handle logging these, currently it just returns a panic message to the user
	server.App.Use(recover.New())

	// Add the endpoints that don't require any special handling for various APIs
	server.App.Get("/", server.HealthCheck)
	server.App.Get("/status", server.HealthCheck)
	server.App.Get("/products", server.productsHandler)
	server.App.Get("/platforms", server.platformsHandler)
	server.App.Get("/architectures", server.architecturesHandler)

	return server
}

func (server *ApiService) Start(wg *sync.WaitGroup) error {
	wg.Add(1)
	go server.StartService()

	return nil
}

func (server *ApiService) StartService() {
	// Setup io writer for the logger
	// Needs to be in the method where we start the service
	// So the io writer will be closed when the service ends
	lw := server.Log.Writer()
	defer lw.Close()

	server.App.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
		Output: lw,
	}))

	server.Log.Infof("Starting %s server at: %s", server.Config.Name, server.Config.Listen)
	err := server.App.Listen(server.Config.Listen)
	if err != nil {
		if err == http.ErrServerClosed {
			server.Log.WithError(err).Error("Unable to start service")
		} else {
			server.Log.WithError(err).Fatal("Service stopped")
		}
	}
}

func (server *ApiService) ValidateRequest(params *omnitruck.RequestParams, c *fiber.Ctx) (error, bool) {
	server.Log.Debugf("Validating request %+v", params)
	errors := server.Validator.Params(params)
	if errors != nil {
		msgs, code := server.Validator.ErrorMessages(errors)

		server.Log.WithField("errors", msgs).Error("Error validating request")
		return c.Status(code).JSON(ErrorResponse{
			Code:       code,
			StatusText: http.StatusText(code),
			Message:    msgs,
		}), false
	}

	return nil, true
}

func (server *ApiService) SendResponse(c *fiber.Ctx, data clients.RequestDataInterface) error {
	return c.JSON(data)
}

func (server *ApiService) SendError(c *fiber.Ctx, request *clients.Request) error {
	return c.Status(request.Code).JSON(ErrorResponse{
		Code:       request.Code,
		StatusText: http.StatusText(request.Code),
		Message:    request.Message,
	})
}

func (server *ApiService) HealthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	return c.JSON(res)
}

// @description Returns a valid list of valid product keys.
// @description Any of these product keys can be used in the <PRODUCT> value of other endpoints. Please note many of these products are used for internal tools only and many have been EOL’d.
// @Param eol			query 	bool 	false 	"EOL Products"
// @Success 200 {object} omnitruck.ItemList
// @Failure 500 {object} services.ErrorResponse
// @Router /products [get]
func (server *ApiService) productsHandler(c *fiber.Ctx) error {
	params := &omnitruck.RequestParams{
		Eol: c.Query("eol", "false"),
	}

	var data omnitruck.ItemList
	request := server.Omnitruck.Products(params, &data)

	if params.Eol != "true" {
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
// @Success 200 {object} omnitruck.PlatformList
// @Failure 500 {object} services.ErrorResponse
// @Router /platforms [get]
func (server *ApiService) platformsHandler(c *fiber.Ctx) error {
	var data omnitruck.PlatformList
	request := server.Omnitruck.Platforms().ParseData(&data)

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}
}

// @description Returns a valid list of valid platform keys along with friendly names.
// @description Any of these architecture keys can be used in the p query string value in various endpoints below.
// @Success 200 {object} omnitruck.ItemList
// @Failure 500 {object} services.ErrorResponse
// @Router /architectures [get]
func (server *ApiService) architecturesHandler(c *fiber.Ctx) error {

	var data omnitruck.ItemList
	request := server.Omnitruck.Architectures().ParseData(&data)

	if request.Ok {
		return server.SendResponse(c, &data)
	} else {
		return server.SendError(c, request)
	}
}
