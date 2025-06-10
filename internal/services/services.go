package services

import (
	"io"
	"net/http"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/dboperations"
	helpers "github.com/chef/omnitruck-service/internal/helper"
	"github.com/chef/omnitruck-service/internal/strategy"
	"github.com/chef/omnitruck-service/logger"
	"github.com/chef/omnitruck-service/models"
	"github.com/chef/omnitruck-service/utils/template"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do"
	log "github.com/sirupsen/logrus"
)

type DownloadService struct {
	Log              *log.Entry
	Validator        omnitruck.RequestValidator
	DatabaseService  dboperations.IDbOperations
	TemplateRenderer template.TemplateRender
	Replicated       replicated.IReplicated
	LicenseClient    clients.ILicense
	Mode             models.ApiType
	locals           map[string]interface{}
}

func NewDownloadService(c *fiber.Ctx, log *log.Entry) (*DownloadService, error) {
	service := DownloadService{}
	service.SetLocals(c)
	service.Log = log
	// Retrieve the injector from the Fiber context
	reqInjectorI := c.Locals("reqinjector")
	reqInjector, ok := reqInjectorI.(*do.Injector)
	if !ok {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve request injector")
	}

	validator := do.MustInvokeNamed[omnitruck.RequestValidator](reqInjector, "validator")
	databaseService := do.MustInvokeNamed[dboperations.IDbOperations](reqInjector, "dbService")
	templateRenderer := do.MustInvokeNamed[template.TemplateRender](reqInjector, "templateRenderer")
	replicatedService := do.MustInvokeNamed[replicated.IReplicated](reqInjector, "replicated")
	licenseClient := do.MustInvokeNamed[clients.ILicense](reqInjector, "licenseClient")
	mode := do.MustInvokeNamed[models.ApiType](reqInjector, "mode")
	service.Validator = validator
	service.DatabaseService = databaseService
	service.TemplateRenderer = templateRenderer
	service.Replicated = replicatedService
	service.LicenseClient = licenseClient
	service.Mode = mode
	return &service, nil
}
func (server *DownloadService) SetLocals(c *fiber.Ctx) {
	locals := map[string]interface{}{}
	validLicense := c.Locals("valid_license").(bool)
	//check if c.Locals("request_id") is present
	if c.Locals("request_id") != nil {
		requestId := c.Locals("request_id").(string)
		locals["request_id"] = requestId

	} else {
		locals["request_id"] = ""
	}
	if c.Locals("base_url") != nil {
		baseUrl := c.Locals("base_url").(string)
		locals["base_url"] = baseUrl
	} else {
		locals["base_url"] = ""
	}
	if c.Locals("license_id") != nil {
		licenseId := c.Locals("license_id").(string)
		locals["license_id"] = licenseId
	} else {
		locals["license_id"] = ""
	}
	locals["valid_license"] = validLicense
	server.locals = locals
}

func (server *DownloadService) Omnitruck() *omnitruck.Omnitruck {
	client := omnitruck.New(server.logCtx())

	return &client
}

func (server *DownloadService) DynamoServices(db dboperations.IDbOperations) *omnitruck.DynamoServices {
	service := omnitruck.NewDynamoServices(db, server.logCtx())

	return &service
}

func (server *DownloadService) PlatformServices() *omnitruck.PlatformServices {
	service := omnitruck.NewPlatformServices(server.logCtx())
	return &service
}

func (server *DownloadService) ReplicatedService(config config.ReplicatedConfig, log logger.Logger) replicated.IReplicated {
	service := replicated.NewReplicatedImpl(config, log)
	return service
}

func (server *DownloadService) logCtx() *log.Entry {
	return server.Log.WithField("license_id", server.locals["license_id"])
}

func (server *DownloadService) validLicense() bool {
	v := server.locals["valid_license"]
	return v != nil && v.(bool)
}

func (server *DownloadService) ValidateRequest(params *omnitruck.RequestParams) (string, int, bool) {
	server.logCtx().Debugf("Validating request %+v", params)
	context := omnitruck.Context{
		License: server.validLicense(),
	}

	errors := server.Validator.Params(params, context)
	if errors != nil {
		msgs, code := server.Validator.ErrorMessages(errors)

		return msgs, code, false
	}

	// server.logCtx(c).WithField("errors", msgs).Error("Error validating request")
	// 	return c.Status(code).JSON(ErrorResponse{
	// 		Code:       code,
	// 		StatusText: http.StatusText(code),
	// 		Message:    msgs,
	// 	}), false

	return "", 0, true
}

func (server *DownloadService) Products(params *omnitruck.RequestParams) (data omnitruck.ItemList, request *clients.Request) {
	request = server.Omnitruck().Products(params, &data)

	data = server.DynamoServices(server.DatabaseService).Products(data, params.Eol)

	getServerStrategy := strategy.SelectModeStrategy(server.Mode)
	data = getServerStrategy.FilterProducts(data)

	// if server.Mode == Opensource {
	// 	server.logCtx(c).Info("filtering opensource products")
	// 	data = omnitruck.SelectList(data, omnitruck.OsProductName)
	// } else if params.Eol != "true" {
	// 	server.logCtx(c).Info("filtering eol products")
	// 	data = omnitruck.FilterList(data, omnitruck.EolProductName)
	// }

	// if server.Mode == Trial {
	// 	data = omnitruck.FilterProductsForFreeTrial(data, omnitruck.ProductsForFreeTrial)
	// 	omnitruck.ProductDisplayName(data)
	// }

	// if server.Mode == Commercial {
	// 	data = append(data, constants.PLATFORM_SERVICE_PRODUCT)
	// }

	return data, request

}

func (server *DownloadService) Platforms() (data omnitruck.PlatformList, request *clients.Request) {
	request = server.Omnitruck().Platforms().ParseData(&data)

	data = server.DynamoServices(server.DatabaseService).Platforms(data)

	return data, request
}

func (server *DownloadService) Architectures() (data omnitruck.ItemList, request *clients.Request) {

	request = server.Omnitruck().Architectures().ParseData(&data)

	return data, request
}

func (server *DownloadService) LatestVersion(params *omnitruck.RequestParams) (data omnitruck.ProductVersion, request *clients.Request) {
	// Two-Level Strategy: select both product and mode strategies
	// Get all versions using product strategy
	// Filter versions using mode strategy
	// Return the latest version (assume last in filtered list is latest)

	productStrategyDeps := &strategy.ProductStrategyDeps{
		DynamoService:    server.DynamoServices(server.DatabaseService),
		PlatformService:  server.PlatformServices(),
		OmnitruckService: server.Omnitruck(),
		Log:              server.logCtx(),
	}
	productStrategy := strategy.SelectProductStrategy(params.Product, productStrategyDeps)
	modeStrategy := strategy.SelectModeStrategy(server.Mode)

	versions, req := productStrategy.GetAllVersions(params)

	if !req.Ok {
		return "", req
	}

	filtered := modeStrategy.FilterVersions(versions, params.Product)
	if len(filtered) == 0 {
		return "", &clients.Request{
			Ok:      false,
			Code:    fiber.StatusNotFound,
			Message: "No versions found for this product/mode",
		}
	}

	latest := filtered[len(filtered)-1]
	return latest, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "Latest version found",
	}
}

func (server *DownloadService) ProductVersions(params *omnitruck.RequestParams) (data []omnitruck.ProductVersion, request *clients.Request) {
	// Two-Level Strategy: select both product and mode strategies
	// Get all versions using product strategy
	// Filter versions using mode strategy

	productStrategyDeps := &strategy.ProductStrategyDeps{
		DynamoService:    server.DynamoServices(server.DatabaseService),
		PlatformService:  server.PlatformServices(),
		OmnitruckService: server.Omnitruck(),
		Log:              server.logCtx(),
	}
	productStrategy := strategy.SelectProductStrategy(params.Product, productStrategyDeps)
	modeStrategy := strategy.SelectModeStrategy(server.Mode)

	versions, req := productStrategy.GetAllVersions(params)

	if !req.Ok {
		return nil, req
	}

	filtered := modeStrategy.FilterVersions(versions, params.Product)
	return filtered, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "All versions found",
	}
}

func (server *DownloadService) ProductPackages(params *omnitruck.RequestParams) (data omnitruck.PackageList, request *clients.Request) {
	msg, code, ok := server.ValidateRequest(params)
	if !ok {
		return nil, &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}

	productStrategyDeps := &strategy.ProductStrategyDeps{
		DynamoService:    server.DynamoServices(server.DatabaseService),
		PlatformService:  server.PlatformServices(),
		OmnitruckService: server.Omnitruck(),
		Log:              server.logCtx(),
	}
	productStrategy := strategy.SelectProductStrategy(params.Product, productStrategyDeps)
	modeStrategy := strategy.SelectModeStrategy(server.Mode)

	// Get all versions using product strategy
	versions, req := productStrategy.GetAllVersions(params)
	if !req.Ok || len(versions) == 0 {
		return nil, req
	}

	// Filter versions using mode strategy
	filtered := modeStrategy.FilterVersions(versions, params.Product)
	if len(filtered) == 0 {
		return nil, &clients.Request{
			Ok:      false,
			Code:    fiber.StatusNotFound,
			Message: "No versions found for this product/mode",
		}
	}

	// If a version is provided, validate it is in the filtered list
	if err := helpers.ValidateOrSetVersion(params, filtered); err != nil {
		return nil, &clients.Request{
			Ok:      false,
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	data, err := productStrategy.GetPackages(params)
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		return nil, &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}
	productStrategy.UpdatePackages(&data, params, server.locals["base_url"].(string))

	return data, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "Packages retrieved successfully",
	}
}

func (server *DownloadService) ProductMetadata(params *omnitruck.RequestParams) (data omnitruck.PackageMetadata, request *clients.Request) {
	msg, code, ok := server.ValidateRequest(params)
	if !ok {
		return omnitruck.PackageMetadata{}, &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}

	productStrategyDeps := &strategy.ProductStrategyDeps{
		DynamoService:    server.DynamoServices(server.DatabaseService),
		PlatformService:  server.PlatformServices(),
		OmnitruckService: server.Omnitruck(),
		Log:              server.logCtx(),
	}
	productStrategy := strategy.SelectProductStrategy(params.Product, productStrategyDeps)
	modeStrategy := strategy.SelectModeStrategy(server.Mode)

	// Get all versions using product strategy
	versions, req := productStrategy.GetAllVersions(params)
	if !req.Ok || len(versions) == 0 {
		return omnitruck.PackageMetadata{}, req
	}

	// Filter versions using mode strategy
	filtered := modeStrategy.FilterVersions(versions, params.Product)
	if len(filtered) == 0 {
		return omnitruck.PackageMetadata{}, &clients.Request{
			Ok:      false,
			Code:    fiber.StatusNotFound,
			Message: "No versions found for this product/mode",
		}
	}

	// If a version is provided, validate it is in the filtered list
	if err := helpers.ValidateOrSetVersion(params, filtered); err != nil {
		return omnitruck.PackageMetadata{}, &clients.Request{
			Ok:      false,
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	data, request = productStrategy.GetMetadata(params)

	// Remap the package url to our download URL
	url := helpers.GetDownloadUrl(params, server.locals["base_url"].(string))
	data.Url = url

	if request.Ok {
		return data, &clients.Request{
			Ok:      true,
			Code:    fiber.StatusOK,
			Message: "Metadata retrieved successfully",
		}
	} else {
		return omnitruck.PackageMetadata{}, req
	}
}

func (server *DownloadService) RelatedProducts(params *omnitruck.RequestParams) (data map[string]interface{}, request *clients.Request) {
	server.logCtx().Info("Validating related products API for " + params.BOM)

	msg, code, ok := server.ValidateRequest(params)
	if !ok {
		server.logCtx().Error("Validation of related products API for "+params.BOM+"failed", msg)
		return nil, &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}

	relatedProducts, err := server.DynamoServices(server.DatabaseService).GetRelatedProducts(params)

	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		server.logCtx().Error("Error while fetching related products for "+params.BOM, err.Error())
		return nil, &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}

	response := map[string]interface{}{
		"relatedProducts": relatedProducts.Products,
	}
	server.logCtx().Info("Returning success response from related products API for " + params.BOM)
	return response, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "Related products retrieved successfully",
	}
}

func (server *DownloadService) GetFileName(params *omnitruck.RequestParams) (string, *clients.Request) {
	msg, code, ok := server.ValidateRequest(params)
	if !ok {
		server.logCtx().Error("Validation of file name API for " + params.Product + " failed")
		return "", &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}

	// Two-Level Strategy: select both product and mode strategies
	productStrategyDeps := &strategy.ProductStrategyDeps{
		DynamoService:    server.DynamoServices(server.DatabaseService),
		PlatformService:  server.PlatformServices(),
		OmnitruckService: server.Omnitruck(),
		Log:              server.logCtx(),
	}
	productStrategy := strategy.SelectProductStrategy(params.Product, productStrategyDeps)
	modeStrategy := strategy.SelectModeStrategy(server.Mode)

	// Get all versions using product strategy
	versions, req := productStrategy.GetAllVersions(params)
	if !req.Ok || len(versions) == 0 {
		return "", req
	}

	// Filter versions using mode strategy
	filtered := modeStrategy.FilterVersions(versions, params.Product)
	if len(filtered) == 0 {
		return "", &clients.Request{
			Ok:      false,
			Code:    fiber.StatusNotFound,
			Message: "No versions found for this product/mode",
		}
	}

	// If a version is provided, validate it is in the filtered list
	if err := helpers.ValidateOrSetVersion(params, filtered); err != nil {
		return "", &clients.Request{
			Ok:      false,
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	fileName, err := productStrategy.GetFileName(params)
	if err != nil {
		server.logCtx().Error("Error while fetching fileName for "+params.Product, err.Error())
		return "", &clients.Request{
			Ok:      false,
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	server.logCtx().Info(constants.SUCCESS_RESPONSE_FROM_FILENAME_MSG + params.Product)
	return fileName, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "File name retrieved successfully",
	}
}

func (server *DownloadService) GetLinuxScript(params *omnitruck.RequestParams) (string, *clients.Request) {
	msg, code, ok := server.ValidateRequest(params)
	if !ok {
		server.logCtx().Error("Validation of download linux script API failed: ", msg)
		return "", &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}
	if server.Mode == models.Opensource {
		params.LicenseId = ""
	}
	filePath := "../../templates/install.sh.tmpl"
	resp, err := server.TemplateRenderer.GetScript(server.locals["base_url"].(string), params, filePath)
	if err != nil {
		return "", &clients.Request{
			Ok:      false,
			Code:    fiber.StatusInternalServerError,
			Message: "Error generating script: " + err.Error(),
		}
	}
	return resp, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "Linux script generated successfully",
	}
}

func (server *DownloadService) GetWindowsScript(params *omnitruck.RequestParams) (string, *clients.Request) {
	msg, code, ok := server.ValidateRequest(params)
	if !ok {
		server.logCtx().Error("Validation of download windows script API failed: ", msg)
		return "", &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}
	if server.Mode == models.Opensource {
		params.LicenseId = ""
	}
	filePath := "../../templates/install.ps1.tmpl"
	resp, err := server.TemplateRenderer.GetScript(server.locals["base_url"].(string), params, filePath)
	if err != nil {
		return "", &clients.Request{
			Ok:      false,
			Code:    fiber.StatusInternalServerError,
			Message: "Error generating script: " + err.Error(),
		}
	}
	return resp, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "Windows script generated successfully",
	}
}

func (server *DownloadService) ProductDownload(params *omnitruck.RequestParams, c *fiber.Ctx) (string, io.ReadCloser, http.Header, string, int, error) {
	msg, code, ok := server.ValidateRequest(params)
	if !ok {
		return "", nil, nil, msg, code, fiber.NewError(code, msg)
	}
	server.logCtx().Infof("Recieved product download request for %s", params.Product)

	// Two-Level Strategy: select both product and mode strategies
	productStrategyDeps := &strategy.ProductStrategyDeps{
		DynamoService:    server.DynamoServices(server.DatabaseService),
		PlatformService:  server.PlatformServices(),
		OmnitruckService: server.Omnitruck(),
		Log:              server.logCtx(),
	}
	productStrategy := strategy.SelectProductStrategy(params.Product, productStrategyDeps)
	modeStrategy := strategy.SelectModeStrategy(server.Mode)

	// Get all versions using product strategy
	versions, req := productStrategy.GetAllVersions(params)
	if !req.Ok || len(versions) == 0 {
		return "", nil, nil, req.Message, req.Code, fiber.NewError(req.Code, req.Message)
		//return server.SendError(c, req)
	}

	// Filter versions using mode strategy
	filtered := modeStrategy.FilterVersions(versions, params.Product)
	if len(filtered) == 0 {
		return "", nil, nil, "No versions found for this product/mode", fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "No versions found for this product/mode")
		//return server.SendErrorResponse(c, fiber.StatusNotFound, "No versions found for this product/mode")
	}

	// Validate or set version
	if err := helpers.ValidateOrSetVersion(params, filtered); err != nil {
		return "", nil, nil, err.Error(), fiber.StatusBadRequest, err
		//return server.SendErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	// Download using the product strategy
	return productStrategy.Download(params, c)
}

func (server *DownloadService) GetPackageManagers() (data omnitruck.ItemList, request *clients.Request) {
	server.logCtx().Info("Fetching package managers")
	packageManagers, err := server.DynamoServices(server.DatabaseService).GetPackageManagers()
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		return nil, &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}
	return packageManagers, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "Package managers retrieved successfully",
	}
}

func isLatest(v string) bool {
	return len(v) == 0 || v == "latest"
}
