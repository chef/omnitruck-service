package services

import (
	"fmt"
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
	"github.com/chef/omnitruck-service/utils/template"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do"
	log "github.com/sirupsen/logrus"
)

type DownloadService struct {
	log               *log.Entry
	databaseService   dboperations.IDbOperations
	templateRenderer  template.TemplateRenderer
	replicated        replicated.IReplicated
	licenseClient     clients.ILicense
	licenseServiceUrl string
	mode              constants.ApiType
	locals            map[string]interface{}
	config            config.ServiceConfig
}

func NewDownloadService(injector *do.Injector, log *log.Entry, locals map[string]interface{}) (*DownloadService, error) {
	service := &DownloadService{
		log:    log,
		locals: locals,
	}

	var err error

	if service.databaseService, err = do.InvokeNamed[dboperations.IDbOperations](injector, "dbService"); err != nil {
		return nil, fmt.Errorf("could not resolve dbService: %w", err)
	}
	if service.templateRenderer, err = do.InvokeNamed[template.TemplateRenderer](injector, "templateRenderer"); err != nil {
		return nil, fmt.Errorf("could not resolve templateRenderer: %w", err)
	}
	if service.replicated, err = do.InvokeNamed[replicated.IReplicated](injector, "replicated"); err != nil {
		return nil, fmt.Errorf("could not resolve replicated: %w", err)
	}
	if service.licenseClient, err = do.InvokeNamed[clients.ILicense](injector, "licenseClient"); err != nil {
		return nil, fmt.Errorf("could not resolve licenseClient: %w", err)
	}
	if service.mode, err = do.InvokeNamed[constants.ApiType](injector, "mode"); err != nil {
		return nil, fmt.Errorf("could not resolve mode: %w", err)
	}
	if cfg, err := do.InvokeNamed[config.ServiceConfig](injector, "config"); err != nil {
		return nil, fmt.Errorf("could not resolve config: %w", err)
	} else {
		service.config = cfg
		service.licenseServiceUrl = cfg.LicenseServiceUrl
	}

	return service, nil
}

func (svc *DownloadService) setMode(mode constants.ApiType) {
	svc.mode = mode
}

func (svc *DownloadService) Omnitruck() *omnitruck.Omnitruck {
	client := omnitruck.New(svc.logCtx(), svc.config.OmnitruckUrl)

	return &client
}

func (svc *DownloadService) DynamoServices(db dboperations.IDbOperations) *omnitruck.DynamoServices {
	service := omnitruck.NewDynamoServices(db, svc.logCtx())

	return &service
}

func (svc *DownloadService) PlatformServices() omnitruck.IPlatformServices {
	service := omnitruck.NewPlatformServices(svc.logCtx())
	return service
}

func (svc *DownloadService) ReplicatedService(config config.ReplicatedConfig, log logger.Logger) replicated.IReplicated {
	service := replicated.NewReplicatedImpl(config, log)
	return service
}

func (svc *DownloadService) logCtx() *log.Entry {
	return svc.log.WithField("license_id", svc.locals["license_id"])
}

func (svc *DownloadService) Products(params *omnitruck.RequestParams) (data omnitruck.ItemList, request *clients.Request) {
	request = svc.Omnitruck().Products(params, &data)

	data = svc.DynamoServices(svc.databaseService).Products(data, params.Eol)
	// This a temporary fix to hide CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT when infra19Enabled is false
	if !svc.config.SupportInfra19 {
		// Remove CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT from the data array
		var filtered omnitruck.ItemList
		for _, item := range data {
			if item != constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT && item != constants.MIGRATE_ICE {
				filtered = append(filtered, item)
			}
		}
		data = filtered
	}
	getServerStrategy := strategy.SelectModeStrategy(svc.mode)
	eol := params.Eol == "true"
	data = getServerStrategy.FilterProducts(data, eol)

	if !svc.config.SupportInfra19 {
		var filtered omnitruck.ItemList
		for _, item := range data {
			if item == constants.CHEF_INFRA_CLIENT_NEW_VALUE {
				filtered = append(filtered, constants.CHEF_INFRA_CLIENT_OLD_VALUE)
			} else {
				filtered = append(filtered, item)
			}
		}
		data = filtered
	}

	return data, request

}

func (svc *DownloadService) Platforms() (data omnitruck.PlatformList, request *clients.Request) {
	request = svc.Omnitruck().Platforms().ParseData(&data)

	data = svc.DynamoServices(svc.databaseService).Platforms(data)

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
	filtered, req := svc.getFilteredVersions(params)
	if req != nil {
		return omnitruck.PackageMetadata{}, req
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

	relatedProducts, err := svc.DynamoServices(svc.databaseService).GetRelatedProducts(params)

	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		svc.logCtx().Error("Error while fetching related products for "+params.BOM, err.Error())
		return nil, &clients.Request{
			Ok:      false,
			Code:    code,
			Message: msg,
		}
	}

	if !svc.config.SupportInfra19 && relatedProducts != nil && relatedProducts.Products != nil {
		// This a temporary fix to hide CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT when supportInfra19 is false
		delete(relatedProducts.Products, constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT)
		delete(relatedProducts.Products, constants.MIGRATE_ICE)
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

	// Get all versions using product strategy
	versions, req := productStrategy.GetAllVersions(params)
	if !req.Ok || len(versions) == 0 {
		return "", req
	}

	// Get all versions using product strategy
	filtered, req := svc.getFilteredVersions(params)
	if req != nil {
		return "", req
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
	// Call omnitruck API to get the install.sh script
	client := omnitruck.New(svc.logCtx(), svc.config.OmnitruckUrl)
	request := client.InstallSh(params.LicenseId)

	if !request.Ok {
		return "", &clients.Request{
			Ok:      false,
			Code:    request.Code,
			Message: "Error fetching install.sh from omnitruck: " + request.Message,
		}
	}

	return string(request.Body), &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "Linux script retrieved successfully",
	}
}

func (svc *DownloadService) GetWindowsScript(params *omnitruck.RequestParams) (string, *clients.Request) {
	// Call omnitruck API to get the install.ps1 script
	client := omnitruck.New(svc.logCtx(), svc.config.OmnitruckUrl)
	request := client.InstallPs1(params.LicenseId)

	if !request.Ok {
		return "", &clients.Request{
			Ok:      false,
			Code:    request.Code,
			Message: "Error fetching install.ps1 from omnitruck: " + request.Message,
		}
	}

	return string(request.Body), &clients.Request{
		Ok:      true,
		Code:    fiber.StatusOK,
		Message: "Windows script retrieved successfully",
	}
}

func (svc *DownloadService) ProductDownload(params *omnitruck.RequestParams, c *fiber.Ctx) (string, io.ReadCloser, http.Header, string, int, error) {
	svc.logCtx().Infof("Received product download request for %s", params.Product)
	// Two-Level Strategy: select both product and mode strategies
	productStrategy := strategy.SelectProductStrategy(params.Product, params.Channel, svc.ProductStrategyDeps())

	// Get all versions using product strategy
	versions, req := productStrategy.GetAllVersions(params)
	if !req.Ok || len(versions) == 0 {
		return "", nil, nil, req.Message, req.Code, fiber.NewError(req.Code, req.Message)
		//return svc.SendError(c, req)
	}

	// Get all versions using product strategy
	filtered, req := svc.getFilteredVersions(params)
	if req != nil {
		return "", nil, nil, req.Message, req.Code, fiber.NewError(req.Code, req.Message)
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
	packageManagers, err := svc.DynamoServices(svc.databaseService).GetPackageManagers()
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
		DynamoService:     svc.DynamoServices(svc.databaseService),
		PlatformService:   svc.PlatformServices(),
		OmnitruckService:  svc.Omnitruck(),
		Log:               svc.logCtx(),
		Replicated:        svc.replicated,
		LicenseClient:     svc.licenseClient,
		LicenseServiceUrl: svc.licenseServiceUrl,
		Mode:              svc.mode,
		Config:            svc.config,
		Locals:            svc.locals,
	}
}

func (svc *DownloadService) getFilteredVersions(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, *clients.Request) {
	productStrategy := strategy.SelectProductStrategy(params.Product, params.Channel, svc.ProductStrategyDeps())
	modeStrategy := strategy.SelectModeStrategy(svc.mode)
	versions, req := productStrategy.GetAllVersions(params)
	if !req.Ok || len(versions) == 0 {
		return nil, req
	}

	// Versions are already sorted by the product strategy
	filtered := modeStrategy.FilterVersions(versions, params.Product, params.Eol)
	if len(filtered) == 0 {
		return nil, &clients.Request{
			Ok:      false,
			Code:    fiber.StatusBadRequest,
			Message: "No versions found for this product/mode",
		}
	}
	return filtered, nil
}
