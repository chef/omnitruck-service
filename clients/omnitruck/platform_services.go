package omnitruck

import (
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/logger"
	"github.com/gofiber/fiber/v2"
)

type PlatformServices struct {
	Logger logger.Logger
}

func NewPlatformServices(logger logger.Logger) PlatformServices {
	return PlatformServices{
		Logger: logger,
	}
}

func (r *PlatformServices) PlatformVersionsAll(req *RequestParams, serverMode int) ([]ProductVersion, error) {
	productVersions := []ProductVersion{}
	flags := RequestParamsFlags{
		Channel: true,
	}
	requestParams := ValidateRequest(req, flags)
	if !requestParams.Ok {
		r.Logger.Error("", requestParams.Message)
		return productVersions, fiber.NewError(requestParams.Code, requestParams.Message)
	}
	if serverMode == 2 && req.Product == constants.PLATFORM_SERVICE {
		productVersions = append(productVersions, "latest")
		return productVersions, nil
	}
	return productVersions, fiber.NewError(fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}

func (r *PlatformServices) PlatformVersionLatest(req *RequestParams, serverMode int) (ProductVersion, error) {
	flags := RequestParamsFlags{
		Channel: true,
	}
	requestParams := ValidateRequest(req, flags)
	if !requestParams.Ok {
		r.Logger.Error("", requestParams.Message)
		return "", fiber.NewError(requestParams.Code, requestParams.Message)
	}
	if serverMode == 2 {
		return "latest", nil
	}
	return "", fiber.NewError(fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}

func (r *PlatformServices) PlatformMetadata(req *RequestParams, serverMode int) (PackageMetadata, error) {
	flags := RequestParamsFlags{
		Channel: true,
	}
	requestParams := ValidateRequest(req, flags)
	if !requestParams.Ok {
		r.Logger.Error("", requestParams.Message)
		return PackageMetadata{}, fiber.NewError(requestParams.Code, requestParams.Message)
	}
	if serverMode == 2 && req.Product == constants.PLATFORM_SERVICE {
		return PackageMetadata{
			Sha1:    "",
			Sha256:  "",
			Url:     "",
			Version: req.Version,
		}, nil
	}
	return PackageMetadata{}, fiber.NewError(fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}

func (r *PlatformServices) PlatformPackages(req *RequestParams, serverMode int) (PackageList, error) {
	packageList := PackageList{}
	flags := RequestParamsFlags{
		Channel: true,
	}
	requestParams := ValidateRequest(req, flags)
	if !requestParams.Ok {
		r.Logger.Error("", requestParams.Message)
		return PackageList{}, fiber.NewError(requestParams.Code, requestParams.Message)
	}
	if req.Version == "" {
		req.Version = "latest"
	}
	if serverMode == 2 && req.Product == constants.PLATFORM_SERVICE {
		packageList["linux"] = PlatformVersionList{}
		packageList["linux"]["pv"] = ArchList{}
		packageList["linux"]["pv"]["amd64"] = PackageMetadata{
			Sha1:    "",
			Sha256:  "",
			Url:     "",
			Version: req.Version,
		}
		return packageList, nil
	}
	return packageList, fiber.NewError(fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}

func (r *PlatformServices) PlatformFilename(req *RequestParams, serverMode int) (string, error) {
	flags := RequestParamsFlags{
		Channel: true,
	}
	requestParams := ValidateRequest(req, flags)
	if !requestParams.Ok {
		r.Logger.Error("", requestParams.Message)
		return "", fiber.NewError(requestParams.Code, requestParams.Message)
	}
	if serverMode == 2 {
		return constants.PLATFORM_SERVICE + ".zip", nil
	}
	return "", fiber.NewError(fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}
