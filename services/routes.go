package services

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
)

// routes sets up all HTTP routes for the ApiService
func (server *ApiService) routes() {
	server.App.Get("/swagger/*", swagger.New(swagger.Config{
		InstanceName: "OmnitruckApi",
	}))

	server.App.Static("/", "./static", fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		Index:         "index.html",
		CacheDuration: 10 * time.Second,
		MaxAge:        3600,
	})
	server.App.Get("/status", requestid.New(), server.HealthCheck)
	server.App.Get("/products", requestid.New(), server.productsHandler)
	server.App.Get("/platforms", requestid.New(), server.platformsHandler)
	server.App.Get("/architectures", requestid.New(), server.architecturesHandler)
	server.App.Get("/:channel/:product/versions/latest", requestid.New(), server.latestVersionHandler)
	server.App.Get("/:channel/:product/versions/all", requestid.New(), server.productVersionsHandler)
	server.App.Get("/:channel/:product/packages", requestid.New(), server.productPackagesHandler)
	server.App.Get("/:channel/:product/metadata", requestid.New(), server.productMetadataHandler)
	server.App.Get("/:channel/:product/download", requestid.New(), server.productDownloadHandler)
	server.App.Get("/relatedProducts", requestid.New(), server.relatedProductsHandler)
	server.App.Get("/:channel/:product/fileName", requestid.New(), server.fileNameHandler)
	server.App.Get("/install.sh", requestid.New(), server.downloadLinuxScript)
	server.App.Get("/install.ps1", requestid.New(), server.downloadWindowsScript)
}
