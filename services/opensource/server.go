package opensource

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	omnitruck "github.com/chef/omnitruck-service/omnitruck-client"
	"github.com/chef/omnitruck-service/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type OpensourceService struct {
	sync.Mutex
	config    services.Config
	omnitruck omnitruck.Omnitruck
	log       *logrus.Entry
	f         *fiber.App
}

func NewOpensourceServer(c services.Config) *OpensourceService {
	return &OpensourceService{
		omnitruck: omnitruck.NewOmnitruckClient(),
		log:       log.WithField("pkg", "OpensourceService"),
		config:    c,
	}
}

func (server *OpensourceService) Start(wg *sync.WaitGroup) error {
	server.Lock()
	defer server.Unlock()

	server.f = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		EnablePrintRoutes:     false,
		ReadTimeout:           300 * time.Second,
		WriteTimeout:          300 * time.Second,
	})

	wg.Add(1)
	go server.startOpensourceService()

	return nil
}

func (server *OpensourceService) Name() string {
	return "OpensourceService"
}

func (server *OpensourceService) productHandler(c *fiber.Ctx) error {
	code, body, err := server.omnitruck.Products()
	if err != nil {
		server.log.WithError(err).Error("Unable to fetch data from Omnitruck API")
	}
	if server.omnitruck.IsSuccess(code) {
		return c.SendString("Unable to fetch data from Omnitruck API")
	}

	return c.JSON(body)
}

func (server *OpensourceService) platformHandler(c *fiber.Ctx) error {
	code, body, err := server.omnitruck.Platforms()
	if err != nil {
		server.log.WithError(err).Error("Unable to fetch data from Omnitruck API")
	}
	if code != 200 {
		return c.SendString("Unable to fetch data from Omnitruck API")
	}

	return c.JSON(body)
}

func (server *OpensourceService) architectureHandler(c *fiber.Ctx) error {
	code, body, err := server.omnitruck.Architectures()
	if err != nil {
		server.log.WithError(err).Error("Unable to fetch data from Omnitruck API")
	}
	if code != 200 {
		return c.SendString("Unable to fetch data from Omnitruck API")
	}

	return c.JSON(body)
}

func (server *OpensourceService) latestVersionHandler(c *fiber.Ctx) error {
	code, body, err := server.omnitruck.LatestVersion(
		c.Params("channel"),
		c.Params("product"),
	)

	if err != nil {
		server.log.WithError(err).Error("Error fetching data from Omnitruck API")
	}
	if code != 200 {
		msg := fmt.Sprintf("%s", err)
		return c.SendString(msg)
	}

	return c.JSON(body)
}

func (server *OpensourceService) buildRouter(lw io.Writer) {
	server.f.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
		Output: lw,
	}))

	server.f.Get("/products", server.productHandler)
	server.f.Get("/platforms", server.platformHandler)
	server.f.Get("/architectures", server.architectureHandler)
	server.f.Get("/:channel/:product/versions/latest", server.latestVersionHandler)
}

func (server *OpensourceService) startOpensourceService() {
	// Setup io writer for the logger
	lw := server.log.Writer()
	defer lw.Close()

	server.buildRouter(lw)

	server.log.Infof("Starting %s server at: %s", server.Name(), server.config.Listen)
	err := server.f.Listen(server.config.Listen)
	if err != nil {
		if err == http.ErrServerClosed {
			server.log.WithError(err).Error("unable to start OpensourceService")
		} else {
			server.log.WithError(err).Fatal("OpensourceService stopped")
		}
	}
}
