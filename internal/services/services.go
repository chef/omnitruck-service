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
	dbconnection "github.com/chef/omnitruck-service/middleware/db"
	"github.com/chef/omnitruck-service/utils/awsutils"
	"github.com/chef/omnitruck-service/utils/template"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do"
	log "github.com/sirupsen/logrus"
)

type DownloadService struct {
	Log               *log.Entry
	Validator         omnitruck.RequestValidator
	DatabaseService   dboperations.IDbOperations
	TemplateRenderer  template.TemplateRender
	Replicated        replicated.IReplicated
	LicenseClient     clients.ILicense
	LicenseServiceUrl string
	Mode              constants.ApiType
	locals            map[string]interface{}
	Config            config.ServiceConfig
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

	config := do.MustInvokeNamed[config.ServiceConfig](reqInjector, "config")
	mode := do.MustInvokeNamed[constants.ApiType](reqInjector, "mode")

	templateRenderer := template.NewTemplateRender()
	service.DatabaseService = dboperations.NewDbOperationsService(dbconnection.NewDbConnectionService(awsutils.NewAwsUtils(), config), config)
	service.TemplateRenderer = templateRenderer
	service.Mode = mode
	service.Config = config
	return &service, nil
}
func (svc *DownloadService) SetLocals(c *fiber.Ctx) {
	locals := map[string]interface{}{}
	if c.Locals("valid_license") != nil {
		requestId := c.Locals("valid_license").(bool)
		locals["valid_license"] = requestId

	} else {
		locals["valid_license"] = false
	}

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
		locals["base_url"] = c.BaseURL()
	}
	if c.Locals("license_id") != nil {
		licenseId := c.Locals("license_id").(string)
		locals["license_id"] = licenseId
	} else {
		locals["license_id"] = ""
	}
	svc.locals = locals
}

func (svc *DownloadService) Omnitruck() *omnitruck.Omnitruck {
	client := omnitruck.New(svc.logCtx())

	return &client
}

func (svc *DownloadService) DynamoServices(db dboperations.IDbOperations) *omnitruck.DynamoServices {
	service := omnitruck.NewDynamoServices(db, svc.logCtx())

	return &service
}

func (svc *DownloadService) PlatformServices() *omnitruck.PlatformServices {
	service := omnitruck.NewPlatformServices(svc.logCtx())
	return &service
}

func (svc *DownloadService) ReplicatedService(config config.ReplicatedConfig, log logger.Logger) replicated.IReplicated {
	service := replicated.NewReplicatedImpl(config, log)
	return service
}

func (svc *DownloadService) logCtx() *log.Entry {
	return svc.Log.WithField("license_id", svc.locals["license_id"])
}

func (svc *DownloadService) Products(params *omnitruck.RequestParams) (data omnitruck.ItemList, request *clients.Request) {
	request = svc.Omnitruck().Products(params, &data)

	data = svc.DynamoServices(svc.DatabaseService).Products(data, params.Eol)

	getServerStrategy := strategy.SelectModeStrategy(svc.Mode)
	eol := params.Eol == "true"
	data = getServerStrategy.FilterProducts(data, eol)

	// if svc.Mode == Opensource {
	//  svc.logCtx(c).Info("filtering opensource products")
	//  data = omnitruck.SelectList(data, omnitruck.OsProductName)
	// } else if params.Eol != "true" {
	//  svc.logCtx(c).Info("filtering eol products")
	//  data = omnitruck.FilterList(data, omnitruck.EolProductName)
	// }

	// if svc.Mode == Trial {
	//  data = omnitruck.FilterProductsForFreeTrial(data, omnitruck.ProductsForFreeTrial)
	//  omnitruck.ProductDisplayName(data)
	// }

	// if svc.Mode == Commercial {
	//  data = append(data, constants.PLATFORM_SERVICE_PRODUCT)
	// }

	return data, request

}

func (svc *DownloadService) Platforms() (data omnitruck.PlatformList, request *clients.Request) {
	request = svc.Omnitruck().Platforms().ParseData(&data)

	data = svc.DynamoServices(svc.DatabaseService).Platforms(data)

	return data, request
}

func (svc *DownloadService) Architectures() (data omnitruck.ItemList, request *clients.Request) {

	request = svc.Omnitruck().Architectures().ParseData(&data)

	return data, request
}

func (svc *DownloadService) LatestVersion(params *omnitruck.RequestParams) (data omnitruck.ProductVersion, request *clients.Request) {
	// Two-Level Strategy: select both product and mode strategies
	// Get all versions using product strategy
	// Filter versions using mode strategy
	// Return the latest version (assume last in filtered list is latest)
	filtered, err := svc.getFilteredVersions(params)
	if err != nil {
		return "", err
	}

	latest := filtered[len(filtered)-1]
	return latest, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "Latest version found",
	}
}

func (svc *DownloadService) ProductVersions(params *omnitruck.RequestParams) (data []omnitruck.ProductVersion, request *clients.Request) {
	// Two-Level Strategy: select both product and mode strategies
	// Get all versions using product strategy
	// Filter versions using mode strategy
	filtered, err := svc.getFilteredVersions(params)
	if err != nil {
		return nil, err
	}

	return filtered, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "All versions found",
	}
}

func (svc *DownloadService) ProductPackages(params *omnitruck.RequestParams) (data omnitruck.PackageList, request *clients.Request) {
	productStrategy := strategy.SelectProductStrategy(params.Product, params.Channel, svc.ProductStrategyDeps())
	filtered, err := svc.getFilteredVersions(params)
	if err != nil {
		return nil, err
	}

	// If a version is provided, validate it is in the filtered list
	if err := helpers.ValidateOrSetVersion(params, filtered); err != nil {
		return nil, &clients.Request{
			Ok:      false,
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		}
	}

	data, perr := productStrategy.GetPackages(params)
	if perr != nil {
		code, msg := helpers.GetErrorCodeAndMsg(perr)
		return nil, &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}
	productStrategy.UpdatePackages(&data, params, svc.locals["base_url"].(string))

	return data, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "Packages retrieved successfully",
	}
}

func (svc *DownloadService) ProductMetadata(params *omnitruck.RequestParams) (data omnitruck.PackageMetadata, request *clients.Request) {
	productStrategy := strategy.SelectProductStrategy(params.Product, params.Channel, svc.ProductStrategyDeps())

	// Get all versions using product strategy
	filtered, err := svc.getFilteredVersions(params)
	if err != nil {
		return omnitruck.PackageMetadata{}, err
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
	url := helpers.GetDownloadUrl(params, svc.locals["base_url"].(string))
	data.Url = url

	if request.Ok {
		return data, &clients.Request{
			Ok:      true,
			Code:    fiber.StatusOK,
			Message: "Metadata retrieved successfully",
		}
	} else {
		return omnitruck.PackageMetadata{}, request
	}
}

func (svc *DownloadService) RelatedProducts(params *omnitruck.RequestParams) (data map[string]interface{}, request *clients.Request) {
	svc.logCtx().Info("Validating related products API for " + params.BOM)

	relatedProducts, err := svc.DynamoServices(svc.DatabaseService).GetRelatedProducts(params)

	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		svc.logCtx().Error("Error while fetching related products for "+params.BOM, err.Error())
		return nil, &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}

	response := map[string]interface{}{
		"relatedProducts": relatedProducts.Products,
	}
	svc.logCtx().Info("Returning success response from related products API for " + params.BOM)
	return response, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "Related products retrieved successfully",
	}
}

func (svc *DownloadService) GetFileName(params *omnitruck.RequestParams) (string, *clients.Request) {
	// Two-Level Strategy: select both product and mode strategies
	productStrategy := strategy.SelectProductStrategy(params.Product, params.Channel, svc.ProductStrategyDeps())
	modeStrategy := strategy.SelectModeStrategy(svc.Mode)

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

	if len(filtered) == 0 {
		return "", &clients.Request{
			Ok:      false,
			Code:    fiber.StatusNotFound,
			Message: params.Product + " is not supported for opensource",
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
		svc.logCtx().Error("Error while fetching fileName for "+params.Product, err.Error())
		return "", &clients.Request{
			Ok:      false,
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	svc.logCtx().Info(constants.SUCCESS_RESPONSE_FROM_FILENAME_MSG + params.Product)
	return fileName, &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "File name retrieved successfully",
	}
}

func (svc *DownloadService) GetLinuxScript(params *omnitruck.RequestParams) (string, *clients.Request) {
	if svc.Mode == constants.Opensource {
		params.LicenseId = ""
	}
	filePath := "templates/install.sh.tmpl"
	resp, err := svc.TemplateRenderer.GetScript(svc.locals["base_url"].(string), params, filePath)
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

func (svc *DownloadService) GetWindowsScript(params *omnitruck.RequestParams) (string, *clients.Request) {
	if svc.Mode == constants.Opensource {
		params.LicenseId = ""
	}
	filePath := "templates/install.ps1.tmpl"
	resp, err := svc.TemplateRenderer.GetScript(svc.locals["base_url"].(string), params, filePath)
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

func (svc *DownloadService) ProductDownload(params *omnitruck.RequestParams, c *fiber.Ctx) (string, io.ReadCloser, http.Header, string, int, error) {
	svc.logCtx().Infof("Received product download request for %s", params.Product)
	// Two-Level Strategy: select both product and mode strategies
	productStrategy := strategy.SelectProductStrategy(params.Product, params.Channel, svc.ProductStrategyDeps())
	modeStrategy := strategy.SelectModeStrategy(svc.Mode)

	// Get all versions using product strategy
	versions, req := productStrategy.GetAllVersions(params)
	if !req.Ok || len(versions) == 0 {
		return "", nil, nil, req.Message, req.Code, fiber.NewError(req.Code, req.Message)
		//return svc.SendError(c, req)
	}

	// Filter versions using mode strategy
	filtered := modeStrategy.FilterVersions(versions, params.Product)
	if len(filtered) == 0 {
		return "", nil, nil, "No versions found for this product/mode", fiber.StatusNotFound, fiber.NewError(fiber.StatusNotFound, "No versions found for this product/mode")
		//return svc.SendErrorResponse(c, fiber.StatusNotFound, "No versions found for this product/mode")
	}

	// Validate or set version
	if err := helpers.ValidateOrSetVersion(params, filtered); err != nil {
		return "", nil, nil, err.Error(), fiber.StatusBadRequest, err
		//return svc.SendErrorResponse(c, fiber.StatusBadRequest, err.Error())
	}

	// Download using the product strategy
	return productStrategy.Download(params)
}

func (svc *DownloadService) GetPackageManagers() (data omnitruck.ItemList, request *clients.Request) {
	svc.logCtx().Info("Fetching package managers")
	packageManagers, err := svc.DynamoServices(svc.DatabaseService).GetPackageManagers()
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

func (svc *DownloadService) ProductStrategyDeps() *strategy.ProductStrategyDeps {
	return &strategy.ProductStrategyDeps{
		DynamoService:     svc.DynamoServices(svc.DatabaseService),
		PlatformService:   svc.PlatformServices(),
		OmnitruckService:  svc.Omnitruck(),
		Log:               svc.logCtx(),
		Replicated:        svc.Replicated,
		LicenseClient:     svc.LicenseClient,
		LicenseServiceUrl: svc.LicenseServiceUrl,
		Validator:         svc.Validator,
		Mode:              svc.Mode,
		Config:            svc.Config,
	}
}

func (svc *DownloadService) getFilteredVersions(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, *clients.Request) {
	productStrategy := strategy.SelectProductStrategy(params.Product, params.Channel, svc.ProductStrategyDeps())
	modeStrategy := strategy.SelectModeStrategy(svc.Mode)
	versions, req := productStrategy.GetAllVersions(params)
	if !req.Ok || len(versions) == 0 {
		return nil, req
	}
	filtered := modeStrategy.FilterVersions(versions, params.Product)
	if len(filtered) == 0 {
		return nil, &clients.Request{
			Ok:      false,
			Code:    fiber.StatusBadRequest,
			Message: "No versions found for this product/mode",
		}
	}
	return filtered, nil
}
