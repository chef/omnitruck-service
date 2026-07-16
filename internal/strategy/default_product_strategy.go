package strategy

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	// Sort versions if the request was successful
	if request.Ok {
		data = omnitruck.SortProductVersions(data)
	}

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

// ValidateFilesParams enforces that PlatformVersion is provided for default products.
func (s *DefaultProductStrategy) ValidateFilesParams(params *omnitruck.RequestParams) error {
	if strings.TrimSpace(params.PlatformVersion) == "" {
		return errors.New("Platform Version (pv) params cannot be empty")
	}

	// Validate filename matches what database expects
	correctFileName, err := s.GetFileName(params)
	if err != nil {
		return err
	}
	if params.FileName != correctFileName {
		return fmt.Errorf("invalid filename for the specified product parameters")
	}

	return nil
}

// ParseTail parses the /files URL tail in Default format: {platformVersion}/{arch}/{fileName}
// Expected format: platformVersion/arch/filename (exactly 3 segments)
func (s *DefaultProductStrategy) ParseTail(segments []string) helpers.FilesPathParams {
	p := helpers.FilesPathParams{}

	if len(segments) < 3 {
		return p // Invalid format, will fail validation later
	}

	p.PlatformVersion = segments[0]
	p.Architecture = segments[1]
	p.FileName = segments[2]

	return p
}

func (s *DefaultProductStrategy) UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, baseUrl string) {
	data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
		params.Version = m.Version
		params.Platform = platform
		params.PlatformVersion = pv
		params.Architecture = arch

		fileName := ""
		if params.Direct == "true" {
			if resolvedFileName, err := s.GetFileName(params); err == nil {
				fileName = resolvedFileName
			}
		}

		m.Url = helpers.GetPackageUrl(params, baseUrl, fileName)

		return m
	})
}
