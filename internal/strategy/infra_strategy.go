package strategy

import (
	"io"
	"net/http"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	helpers "github.com/chef/omnitruck-service/internal/helper"
	log "github.com/sirupsen/logrus"
)

type InfraProductStrategy struct {
	DynamoService *omnitruck.DynamoServices
	Log           *log.Entry
}

func (s *InfraProductStrategy) GetLatestVersion(params *omnitruck.RequestParams) (omnitruck.ProductVersion, *clients.Request) {
	request := clients.Request{}
	data, err := s.DynamoService.VersionLatest(params)
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		s.Log.WithError(err).Error("Error while fetching latest versions for "+ params.Product+ ": ", err)
		request.Failure(code, msg)
		return data, &request
	}
	request.Success()
	return data, &request
}

func (s *InfraProductStrategy) GetAllVersions(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, *clients.Request) {
	request := clients.Request{}
	data, err := s.DynamoService.VersionAll(params)
	if err != nil {
		s.Log.WithError(err).Error("Error while fetching all versions for "+ params.Product+ ": ", err)
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
		return nil, &request
	}
	request.Success()
	return data, &request
}

func (s *InfraProductStrategy) Download(params *omnitruck.RequestParams) (url string, resp io.ReadCloser, header http.Header, msg string, code int, err error) {
	url, err = s.DynamoService.ProductDownload(params)
	return url, nil, nil, "", 0, err
}

func (s *InfraProductStrategy) GetPackages(params *omnitruck.RequestParams) (omnitruck.PackageList, error) {
	return s.DynamoService.ProductPackages(params)
}

func (s *InfraProductStrategy) GetMetadata(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, *clients.Request) {
	request := &clients.Request{}
	data, err := s.DynamoService.ProductMetadata(params)
	if err != nil {
		s.Log.WithError(err).Error("Error while fetching metadata for "+ params.Product+ ": ", err)
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
	} else {
		request.Success()
	}
	return data, request
}

func (s *InfraProductStrategy) GetFileName(params *omnitruck.RequestParams) (string, error) {
	fileName, err := s.DynamoService.GetFilename(params)
	return fileName, err
}

func (s *InfraProductStrategy) UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, baseUrl string) {
	data.UpdatePackages(func(platform string, pv string, arch string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
		params.Version = m.Version
		params.Platform = platform
		params.Architecture = arch

		m.Url = helpers.GetDownloadUrl(params, baseUrl)

		return m
	})
}
