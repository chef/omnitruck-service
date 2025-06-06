package httpserver

import (
	"time"

	"github.com/chef/omnitruck-service/internal/api/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
)

// routes sets up all HTTP routes for the ApiService
func (server *ApiServer) buildRouter() {
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
	//New DownloadHandler
	handler := handler.NewDownloadsHandler()

	server.App.Get("/status", requestid.New(), server.HealthCheck)
	server.App.Get("/products", requestid.New(), handler.ProductsHandler)
	server.App.Get("/platforms", requestid.New(), handler.PlatformsHandler)
	server.App.Get("/architectures", requestid.New(), handler.ArchitecturesHandler)
	server.App.Get("/:channel/:product/versions/latest", requestid.New(), handler.LatestVersionHandler)
	server.App.Get("/:channel/:product/versions/all", requestid.New(), handler.ProductVersionsHandler)
	server.App.Get("/:channel/:product/packages", requestid.New(), handler.ProductPackagesHandler)
	server.App.Get("/:channel/:product/metadata", requestid.New(), handler.ProductMetadataHandler)
	server.App.Get("/:channel/:product/download", requestid.New(), handler.ProductDownloadHandler)
	server.App.Get("/relatedProducts", requestid.New(), handler.RelatedProductsHandler)
	server.App.Get("/:channel/:product/fileName", requestid.New(), handler.FileNameHandler)
	server.App.Get("/install.sh", requestid.New(), handler.DownloadLinuxScript)
	server.App.Get("/install.ps1", requestid.New(), handler.DownloadWindowsScript)
}
