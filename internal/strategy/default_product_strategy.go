package strategy

import (
	"fmt"
	"io"
	"net/http"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
	helpers "github.com/chef/omnitruck-service/internal/helper"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

// DefaultProductStrategy implements ProductStrategy for all other products
type DefaultProductStrategy struct {
	OmnitruckService omnitruck.IOmnitruck
	Log              *log.Entry
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
	params.PackageManager = constants.DUMMY_PACKAGE_MANAGER
	request := s.OmnitruckService.ProductMetadata(params).ParseData(&data)
	return data, request
}

func (s *DefaultProductStrategy) Download(params *omnitruck.RequestParams) (url string, resp io.ReadCloser, header http.Header, msg string, code int, err error) {
	var data omnitruck.PackageMetadata
	request := s.OmnitruckService.ProductDownload(params).ParseData(&data)
	if !request.Ok {
		return "", nil, nil, request.Message, request.Code, fiber.NewError(request.Code, request.Message)
	}

	// Append licenseId query parameter if present
	// Note: This URL does not have any existing query parameters
	if params.LicenseId != "" {
		data.Url = fmt.Sprintf("%s?licenseId=%s", data.Url, params.LicenseId)
	}

	return data.Url, nil, nil, request.Message, request.Code, nil
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
