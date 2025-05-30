package services

import (
	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/gofiber/fiber/v2"
)

// ProductDynamoStrategy implements ProductStrategy for Automate and Habitat products
// Uses DynamoDB for most operations
type ProductDynamoStrategy struct {
	Server *ApiService
}

func (s *ProductDynamoStrategy) GetLatestVersion(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.ProductVersion, *clients.Request) {
	request := clients.Request{}
	data, err := s.Server.DynamoServices(s.Server.DatabaseService, c).VersionLatest(params)
	if err != nil {
		code, msg := getErrorCodeAndMsg(err)
		s.Server.logCtx(c).WithError(err).Error("Error while fetching latest version for Automate/Habitat")
		request.Failure(code, msg)
		return data, &request
	}
	request.Success()
	return data, &request
}

func (s *ProductDynamoStrategy) GetAllVersions(params *omnitruck.RequestParams, c *fiber.Ctx) ([]omnitruck.ProductVersion, *clients.Request) {
	request := clients.Request{}
	data, err := s.Server.DynamoServices(s.Server.DatabaseService, c).VersionAll(params)
	if err != nil {
		code, msg := getErrorCodeAndMsg(err)
		request.Failure(code, msg)
		return nil, &request
	}
	request.Success()
	return data, &request
}

func (s *ProductDynamoStrategy) GetPackages(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.PackageList, error) {
	return s.Server.DynamoServices(s.Server.DatabaseService, c).ProductPackages(params)
}

func (s *ProductDynamoStrategy) GetMetadata(params *omnitruck.RequestParams, c *fiber.Ctx) (omnitruck.PackageMetadata, *clients.Request) {
	request := &clients.Request{}
	data, err := s.Server.DynamoServices(s.Server.DatabaseService, c).ProductMetadata(params)
	if err != nil {
		code, msg := getErrorCodeAndMsg(err)
		request.Failure(code, msg)
	} else {
		request.Success()
	}
	return data, request
}

func (s *ProductDynamoStrategy) Download(params *omnitruck.RequestParams, c *fiber.Ctx) error {
	url, err := s.Server.DynamoServices(s.Server.DatabaseService, c).ProductDownload(params)
	if err != nil {
		code, msg := getErrorCodeAndMsg(err)
		return s.Server.SendErrorResponse(c, code, msg)
	}
	s.Server.logCtx(c).Infof("Redirecting user to %s", url)
	return c.Redirect(url, 302)
}

func (s *ProductDynamoStrategy) GetFileName(params *omnitruck.RequestParams, c *fiber.Ctx) (string, error) {
	fileName, err := s.Server.DynamoServices(s.Server.DatabaseService, c).GetFilename(params)
	return fileName, err
}

func (s *ProductDynamoStrategy) UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, c *fiber.Ctx) {
	data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
		params.Version = m.Version
		params.Platform = platform
		params.Architecture = arch

		m.Url = getDownloadUrl(params, c)

		return m
	})
}
