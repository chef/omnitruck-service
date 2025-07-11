package httpserver

import (
	"time"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/dboperations"
	"github.com/chef/omnitruck-service/internal/api/handler"
	"github.com/chef/omnitruck-service/utils/template"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"github.com/samber/do"
)

// @title        Licensed Omnitruck API
// @version      1.0
// @description  Licensed Omnitruck API
// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// routes sets up all HTTP routes for the ApiService
func (server *ApiServer) buildRouter() {
	server.App.Static("/swagger", "./docs")
	server.App.Get("/swagger/*", swagger.New(swagger.Config{
		URL: "/swagger/OmnitruckApi_openapi3.json",
	}))

	server.App.Static("/", "./static", fiber.Static{
		Compress:      true,
		ByteRange:     true,
		Browse:        false,
		Index:         "index.html",
		CacheDuration: 10 * time.Second,
		MaxAge:        3600,
	})

	// Register Injector middleware with dependencies from ApiServer
	server.App.Use(Injector(server))
	//New DownloadHandler
	handler := handler.NewDownloadsHandler(server.Log)

	server.App.Get("/status", requestid.New(), server.HealthCheck)
	server.App.Get("/products", requestid.New(), handler.ProductsHandler)
	server.App.Get("/platforms", requestid.New(), handler.PlatformsHandler)
	server.App.Get("/architectures", requestid.New(), handler.ArchitecturesHandler)
	server.App.Get("/package-managers", requestid.New(), handler.PackageManagersHandler)
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

// Injector middleware now accepts ApiServer and sets dependencies in the injector
func Injector(server *ApiServer) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		reqInjector := do.New()
		// Example: set dependencies from ApiServer. Replace or add your actual fields here.
		do.ProvideNamedValue[omnitruck.IRequestValidator](reqInjector, "validator", &server.Validator)
		do.ProvideNamedValue[dboperations.IDbOperations](reqInjector, "dbService", server.DatabaseService)
		do.ProvideNamedValue[template.TemplateRenderer](reqInjector, "templateRenderer", server.TemplateRenderer)
		do.ProvideNamedValue[replicated.IReplicated](reqInjector, "replicated", server.Replicated)
		do.ProvideNamedValue[clients.ILicense](reqInjector, "licenseClient", server.LicenseClient)
		do.ProvideNamedValue[constants.ApiType](reqInjector, "mode", server.Mode)
		do.ProvideNamedValue[config.ServiceConfig](reqInjector, "config", server.Config.ServiceConfig)
		c.Locals("reqinjector", reqInjector)
		err := c.Next()
		reqInjector.Shutdown()
		return err
	}
}
