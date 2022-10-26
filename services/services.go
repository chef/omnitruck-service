package services

import (
	"net/http"
	"sync"

	omnitruck "github.com/chef/omnitruck-service/omnitruck-client"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
}

func (server *ApiService) New(c Config) *ApiService {
	server.Omnitruck = omnitruck.NewOmnitruckClient()
	server.Log = c.Log.WithField("pkg", c.Name)
	server.Config = c

	return server
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
