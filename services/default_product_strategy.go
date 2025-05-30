package services

import (
	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/gofiber/fiber/v2"
)

// DefaultProductStrategy implements ProductStrategy for all other products
type DefaultProductStrategy struct {
	Server *ApiService
}

func (s *DefaultProductStrategy) GetLatestVersion(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.ProductVersion, *clients.Request) {
	var data omnitruck.ProductVersion
	request := s.Server.Omnitruck(c).LatestVersion(params).ParseData(&data)
	return data, request
}

func (s *DefaultProductStrategy) GetAllVersions(params *omnitruck.RequestParams, c *fiber.Ctx) ([]omnitruck.ProductVersion, *clients.Request) {
	var data []omnitruck.ProductVersion
	request := s.Server.Omnitruck(c).ProductVersions(params).ParseData(&data)
	return data, request
}

func (s *DefaultProductStrategy) GetPackages(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.PackageList, error) {
	var data omnitruck.PackageList
	request := s.Server.Omnitruck(c).ProductPackages(params).ParseData(&data)
	if !request.Ok {
		return data, fiber.NewError(request.Code, request.Message)
	}
	return data, nil
}

func (s *DefaultProductStrategy) GetMetadata(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.PackageMetadata, *clients.Request) {
	var data omnitruck.PackageMetadata
	request := s.Server.Omnitruck(c).ProductMetadata(params).ParseData(&data)
	return data, request
}

func (s *DefaultProductStrategy) Download(params *omnitruck.RequestParams, c *fiber.Ctx) error {
	var data omnitruck.PackageMetadata
	request := s.Server.Omnitruck(c).ProductDownload(params).ParseData(&data)
	if request.Ok {
		return c.Redirect(data.Url, 302)
	}
	return s.Server.SendError(c, request)
}

func (s *DefaultProductStrategy) GetFileName(params *omnitruck.RequestParams, c *fiber.Ctx) (string, error) {
	var data omnitruck.PackageMetadata
	request := s.Server.Omnitruck(c).ProductMetadata(params).ParseData(&data)
	if !request.Ok {
		return "", fiber.NewError(request.Code, request.Message)
	}
	return getFileNameFromURL(data.Url), nil
}

func (s *DefaultProductStrategy) UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, c *fiber.Ctx) {
	data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
		params.Version = m.Version
		params.Platform = platform
		params.PlatformVersion = pv
		params.Architecture = arch

		m.Url = getDownloadUrl(params, c)

		return m
	})
}
