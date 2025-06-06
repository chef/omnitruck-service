package strategy

import (
	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	helpers "github.com/chef/omnitruck-service/internal/helper"
	"github.com/gofiber/fiber/v2"
)

// DefaultProductStrategy implements ProductStrategy for all other products
type DefaultProductStrategy struct {
	OmnitruckService *omnitruck.Omnitruck
	locals           map[string]interface{}
}

func (s *DefaultProductStrategy) GetLatestVersion(params *omnitruck.RequestParams) (omnitruck.ProductVersion, *clients.Request) {
	var data omnitruck.ProductVersion
	request := s.OmnitruckService.LatestVersion(params).ParseData(&data)
	return data, request
}

func (s *DefaultProductStrategy) GetAllVersions(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, *clients.Request) {
	var data []omnitruck.ProductVersion
	request := s.OmnitruckService.ProductVersions(params).ParseData(&data)
	return data, request
}

func (s *DefaultProductStrategy) GetPackages(params *omnitruck.RequestParams) (omnitruck.PackageList, error) {
	var data omnitruck.PackageList
	request := s.OmnitruckService.ProductPackages(params).ParseData(&data)
	if !request.Ok {
		return data, fiber.NewError(request.Code, request.Message)
	}
	return data, nil
}

func (s *DefaultProductStrategy) GetMetadata(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, *clients.Request) {
	var data omnitruck.PackageMetadata
	request := s.OmnitruckService.ProductMetadata(params).ParseData(&data)
	return data, request
}

func (s *DefaultProductStrategy) Download(params *omnitruck.RequestParams, c *fiber.Ctx) error {
	var data omnitruck.PackageMetadata
	request := s.OmnitruckService.ProductDownload(params).ParseData(&data)
	if request.Ok {
		return c.Redirect(data.Url, 302)
	}
	return s.Server.SendError(c, request)
}

func (s *DefaultProductStrategy) GetFileName(params *omnitruck.RequestParams) (string, error) {
	var data omnitruck.PackageMetadata
	request := s.OmnitruckService.ProductMetadata(params).ParseData(&data)
	if !request.Ok {
		return "", fiber.NewError(request.Code, request.Message)
	}
	return helpers.GetFileNameFromURL(data.Url), nil
}

func (s *DefaultProductStrategy) UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, baseUrl string) {
	data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
		params.Version = m.Version
		params.Platform = platform
		params.PlatformVersion = pv
		params.Architecture = arch

		m.Url = helpers.GetDownloadUrl(params, baseUrl)

		return m
	})
}
