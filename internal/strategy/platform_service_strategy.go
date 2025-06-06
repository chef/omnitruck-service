package strategy

import (
	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
	helpers "github.com/chef/omnitruck-service/internal/helper"
	"github.com/chef/omnitruck-service/models"
	"github.com/gofiber/fiber/v2"
)

// PlatformServiceStrategy implements ProductStrategy for PlatformService product
type PlatformServiceStrategy struct {
	PlatformService *omnitruck.PlatformServices
	locals          map[string]interface{}
}

func (s *PlatformServiceStrategy) GetLatestVersion(params *omnitruck.RequestParams) (omnitruck.ProductVersion, *clients.Request) {
	request := clients.Request{}
	data, err := s.PlatformService.PlatformVersionLatest(params, int(s.Server.Mode))
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
		return data, &request
	}
	request.Success()
	return data, &request
}

func (s *PlatformServiceStrategy) GetAllVersions(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, *clients.Request) {
	data, err := s.PlatformService.PlatformVersionsAll(params, int(s.Server.Mode))
	request := &clients.Request{}
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
		return nil, request
	}
	request.Success()
	return data, request
}

func (s *PlatformServiceStrategy) GetPackages(params *omnitruck.RequestParams) (omnitruck.PackageList, error) {
	return s.PlatformService.PlatformPackages(params, int(s.Server.Mode))
}

func (s *PlatformServiceStrategy) GetMetadata(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, *clients.Request) {
	request := &clients.Request{}
	data, err := s.PlatformService.PlatformMetadata(params, int(s.Server.Mode))
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
	} else {
		request.Success()
	}
	return data, request
}

func (s *PlatformServiceStrategy) Download(params *omnitruck.RequestParams, c *fiber.Ctx) error {
	if s.Server.Mode == models.Commercial {
		return s.Server.downloadChefPlatform(params, c)
	}
	return s.Server.SendErrorResponse(c, fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}

func (s *PlatformServiceStrategy) GetFileName(params *omnitruck.RequestParams) (string, error) {
	return s.PlatformService.PlatformFilename(params, int(s.Server.Mode))
}

func (s *PlatformServiceStrategy) UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, baseUrl string) {
	data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
		params.Version = m.Version
		params.Platform = platform
		params.Architecture = arch

		m.Url = helpers.GetDownloadUrl(params, baseUrl)

		return m
	})
}
