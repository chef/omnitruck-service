package strategy

import (
	"io"
	"net/http"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
	helpers "github.com/chef/omnitruck-service/internal/helper"
	log "github.com/sirupsen/logrus"
)

// ProductDynamoStrategy implements ProductStrategy for Automate and Habitat products
// Uses DynamoDB for most operations
type ProductDynamoStrategy struct {
	DynamoService *omnitruck.DynamoServices
	Log           *log.Entry
}

func (s *ProductDynamoStrategy) GetLatestVersion(params *omnitruck.RequestParams) (omnitruck.ProductVersion, *clients.Request) {
	request := clients.Request{}
	data, err := s.DynamoService.VersionLatest(params)
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		s.Log.WithError(err).Error("Error while fetching latest version for Automate/Habitat")
		request.Failure(code, msg)
		return data, &request
	}
	request.Success()
	return data, &request
}

func (s *ProductDynamoStrategy) GetAllVersions(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, *clients.Request) {
	request := clients.Request{}
	data, err := s.DynamoService.VersionAll(params)
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
		return nil, &request
	}
	request.Success()
	return data, &request
}

func (s *ProductDynamoStrategy) GetPackages(params *omnitruck.RequestParams) (omnitruck.PackageList, error) {
	return s.DynamoService.ProductPackages(params)
}

func (s *ProductDynamoStrategy) GetMetadata(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, *clients.Request) {
	request := &clients.Request{}
	params.PackageManager = constants.DUMMY_PACKAGE_MANAGER
	data, err := s.DynamoService.ProductMetadata(params)
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
	} else {
		request.Success()
	}
	return data, request
}

func (s *ProductDynamoStrategy) Download(params *omnitruck.RequestParams) (url string, resp io.ReadCloser, header http.Header, msg string, code int, err error) {
	url, err = s.DynamoService.ProductDownload(params)
	return url, nil, nil, "", 0, err
	// if err != nil {
	// 	code, msg := helpers.GetErrorCodeAndMsg(err)
	// 	return s.Server.SendErrorResponse(c, code, msg)
	// }
	// s.Log.Infof("Redirecting user to %s", url)
	// return c.Redirect(url, 302)
}

func (s *ProductDynamoStrategy) GetFileName(params *omnitruck.RequestParams) (string, error) {
	params.PackageManager = constants.DUMMY_PACKAGE_MANAGER
	fileName, err := s.DynamoService.GetFilename(params)
	return fileName, err
}

func (s *ProductDynamoStrategy) UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, baseUrl string) {
	data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
		params.Version = m.Version
		params.Platform = platform
		params.Architecture = arch

		m.Url = helpers.GetDownloadUrl(params, baseUrl)

		return m
	})
}
