package services

import (
	"net/http"
	"sync"
	"time"

	omnitruck "github.com/chef/omnitruck-service/omnitruck-client"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
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
}

func (server *ApiService) Initialize(c Config) *ApiService {
	server.Omnitruck = omnitruck.NewOmnitruckClient()
	server.Log = c.Log
	server.Config = c
	server.Validator = omnitruck.NewValidator()

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

	return server
}

func (server *ApiService) Start(wg *sync.WaitGroup) error {
	wg.Add(1)
	go server.StartService()

	return nil
}

func (server *ApiService) StartService() {
	// Setup io writer for the logger
	lw := server.Log.Writer()
	defer lw.Close()

	server.Log.Infof("Starting %s server at: %s", server.Config.Name, server.Config.Listen)

	server.App.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
		Output: lw,
	}))

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

func (server *ApiService) SendResponse(c *fiber.Ctx, data omnitruck.RequestDataInterface) error {
	return c.JSON(data)
}

func (server *ApiService) SendError(c *fiber.Ctx, request *omnitruck.Request) error {
	return c.Status(request.Code).JSON(ErrorResponse{
		Code:       request.Code,
		StatusText: http.StatusText(request.Code),
		Message:    request.Message,
	})
}
