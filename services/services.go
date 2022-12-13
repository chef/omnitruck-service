package services

import (
	"net/http"
	"sync"
	"time"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/middleware/license"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
} //@name ErrorResponse

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
	Log       *log.Entry
	App       *fiber.App
	Validator omnitruck.RequestValidator
	Mode      ApiType
}

func New(c Config) *ApiService {
	service := ApiService{}
	service.Initialize(c)

	service.buildRouter()

	return &service
}

func (server *ApiService) Initialize(c Config) *ApiService {
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
	// server.App.Use(recover.New())

	if c.Mode == Trial || c.Mode == Opensource {
		channel := omnitruck.ContainsValidator{
			Field:      "Channel",
			Values:     []string{"stable"},
			Code:       400,
			AllowEmpty: true,
		}
		server.Validator.Add(&channel)
	}

	if c.Mode == Trial {
		version := omnitruck.ContainsValidator{
			Field:      "Version",
			Values:     []string{"latest"},
			Code:       400,
			AllowEmpty: true,
			Skip: func(c omnitruck.Context) bool {
				return c.License
			},
		}
		server.Validator.Add(&version)
	}

	if c.Mode == Trial || c.Mode == Commercial {
		server.Log.Info("Adding EOL Validator")
		eolversion := omnitruck.EolVersionValidator{}
		server.Validator.Add(&eolversion)
	}

	server.App.Use(license.New(license.Config{
		Required: c.Mode == Commercial,
		Next: func(license_id string, c *fiber.Ctx) bool {
			switch c.Path() {
			case "/status":
				return true
			case "/":
				return true
			}

			return false
		},
	}))

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

func (server *ApiService) Omnitruck(c *fiber.Ctx) *omnitruck.Omnitruck {
	client := omnitruck.New(server.logCtx(c))

	return &client
}

func (server *ApiService) logCtx(c *fiber.Ctx) *log.Entry {
	return server.Log.WithField("license_id", c.Locals("license_id"))
}

func (server *ApiService) validLicense(c *fiber.Ctx) bool {
	v := c.Locals("valid_license")
	return v != nil && v.(bool)
}

func (server *ApiService) ValidateRequest(params *omnitruck.RequestParams, c *fiber.Ctx) (error, bool) {
	server.logCtx(c).Debugf("Validating request %+v", params)
	context := omnitruck.Context{
		License: server.validLicense(c),
	}

	errors := server.Validator.Params(params, context)
	if errors != nil {
		msgs, code := server.Validator.ErrorMessages(errors)

		server.logCtx(c).WithField("errors", msgs).Error("Error validating request")
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

func isLatest(v string) bool {
	return len(v) == 0 || v == "latest"
}
