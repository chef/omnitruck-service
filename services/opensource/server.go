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

type ErrorResponse struct {
	Code       int    `json:"code"`
	StatusText string `json:"status_text"`
	Message    string `json:"message"`
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

func (server *OpensourceService) productsHandler(c *fiber.Ctx) error {
	code, body, success := server.omnitruck.Products()
	return server.sendResponse(c, success, code, body)
}

func (server *OpensourceService) platformsHandler(c *fiber.Ctx) error {
	code, body, success := server.omnitruck.Platforms()
	return server.sendResponse(c, success, code, body)
}

func (server *OpensourceService) architecturesHandler(c *fiber.Ctx) error {
	code, body, success := server.omnitruck.Architectures()
	return server.sendResponse(c, success, code, body)
}

func (server *OpensourceService) latestVersionHandler(c *fiber.Ctx) error {
	code, body, success := server.omnitruck.LatestVersion(
		c.Params("channel"),
		c.Params("product"),
	)

	return server.sendResponse(c, success, code, body)
}

func (server *OpensourceService) productVersionsHandler(c *fiber.Ctx) error {
	code, body, success := server.omnitruck.ProductVersions(
		c.Params("channel"),
		c.Params("product"),
	)

	return server.sendResponse(c, success, code, body)
}

func (server *OpensourceService) productPackagesHandler(c *fiber.Ctx) error {
	code, body, success := server.omnitruck.ProductPackages(
		c.Params("channel"),
		c.Params("product"),
		c.Query("v"),
	)

	return server.sendResponse(c, success, code, body)
}

func (server *OpensourceService) productMetadataHandler(c *fiber.Ctx) error {
	code, body, success := server.omnitruck.ProductMetadata(
		c.Params("channel"),
		c.Params("product"),
		c.Query("p"),
		c.Query("pv"),
		c.Query("m"),
		c.Query("v"),
	)

	return server.sendResponse(c, success, code, body)
}

func (server *OpensourceService) buildRouter(lw io.Writer) {
	server.f.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
		Output: lw,
	}))

	server.f.Get("/products", server.productsHandler)
	server.f.Get("/platforms", server.platformsHandler)
	server.f.Get("/architectures", server.architecturesHandler)
	server.f.Get("/:channel/:product/versions/latest", server.latestVersionHandler)
	server.f.Get("/:channel/:product/versions/all", server.productVersionsHandler)
	server.f.Get("/:channel/:product/packages", server.productPackagesHandler)
	server.f.Get("/:channel/:product/metadata", server.productMetadataHandler)
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
