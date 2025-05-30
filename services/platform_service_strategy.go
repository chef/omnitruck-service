package services

import (
	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
	"github.com/gofiber/fiber/v2"
)

// PlatformServiceStrategy implements ProductStrategy for PlatformService product
type PlatformServiceStrategy struct {
	Server *ApiService
}

func (s *PlatformServiceStrategy) GetLatestVersion(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.ProductVersion, *clients.Request) {
	request := clients.Request{}
	data, err := s.Server.PlatformServices(c).PlatformVersionLatest(params, int(s.Server.Mode))
	if err != nil {
		code, msg := getErrorCodeAndMsg(err)
		request.Failure(code, msg)
		return data, &request
	}
	request.Success()
	return data, &request
}

func (s *PlatformServiceStrategy) GetAllVersions(params *omnitruck.RequestParams, c *fiber.Ctx) ([]omnitruck.ProductVersion, *clients.Request) {
	data, err := s.Server.PlatformServices(c).PlatformVersionsAll(params, int(s.Server.Mode))
	request := &clients.Request{}
	if err != nil {
		code, msg := getErrorCodeAndMsg(err)
		request.Failure(code, msg)
		return nil, request
	}
	request.Success()
	return data, request
}

func (s *PlatformServiceStrategy) GetPackages(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.PackageList, error) {
	return s.Server.PlatformServices(c).PlatformPackages(params, int(s.Server.Mode))
}

func (s *PlatformServiceStrategy) GetMetadata(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.PackageMetadata, *clients.Request) {
	request := &clients.Request{}
	data, err := s.Server.PlatformServices(c).PlatformMetadata(params, int(s.Server.Mode))
	if err != nil {
		code, msg := getErrorCodeAndMsg(err)
		request.Failure(code, msg)
	} else {
		request.Success()
	}
	return data, request
}

func (s *PlatformServiceStrategy) Download(params *omnitruck.RequestParams, c *fiber.Ctx) error {
	if s.Server.Mode == Commercial {
		return s.Server.downloadChefPlatform(params, c)
	}
	return s.Server.SendErrorResponse(c, fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}

func (s *PlatformServiceStrategy) GetFileName(params *omnitruck.RequestParams, c *fiber.Ctx) (string, error) {
	return s.Server.PlatformServices(c).PlatformFilename(params, int(s.Server.Mode))
}

func (s *PlatformServiceStrategy) UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, c *fiber.Ctx) {
	data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
		params.Version = m.Version
		params.Platform = platform
		params.Architecture = arch

		m.Url = getDownloadUrl(params, c)

		return m
	})
}
