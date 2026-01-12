package strategy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/constants"
	helpers "github.com/chef/omnitruck-service/internal/helper"
	log "github.com/sirupsen/logrus"
)

// PlatformServiceStrategy implements ProductStrategy for PlatformService product
type PlatformServiceStrategy struct {
	PlatformService   omnitruck.IPlatformServices
	Replicated        replicated.IReplicated
	LicenseClient     clients.ILicense
	LicenseServiceUrl string
	Log               *log.Entry
	Mode              constants.ApiType
	Locals            map[string]interface{}
}

var JsonUnmarshal = func(data []byte, v any) error {
	return json.Unmarshal(data, &v)
}

func (s *PlatformServiceStrategy) GetLatestVersion(params *omnitruck.RequestParams) (omnitruck.ProductVersion, *clients.Request) {
	request := clients.Request{}
	data, err := s.PlatformService.PlatformVersionLatest(params, int(s.Mode))
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
		return data, &request
	}
	request.Success()
	return data, &request
}

func (s *PlatformServiceStrategy) GetAllVersions(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, *clients.Request) {
	data, err := s.PlatformService.PlatformVersionsAll(params, int(s.Mode))
	request := &clients.Request{}
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
		return nil, request
	}

	// Sort versions before returning
	data = omnitruck.SortProductVersions(data)

	request.Success()
	return data, request
}

func (s *PlatformServiceStrategy) GetPackages(params *omnitruck.RequestParams) (omnitruck.PackageList, error) {
	return s.PlatformService.PlatformPackages(params, int(s.Mode))
}

func (s *PlatformServiceStrategy) GetMetadata(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, *clients.Request) {
	request := &clients.Request{}
	params.PackageManager = constants.DUMMY_PACKAGE_MANAGER
	data, err := s.PlatformService.PlatformMetadata(params, int(s.Mode))
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
	} else {
		request.Success()
	}
	return data, request
}

func (s *PlatformServiceStrategy) Download(params *omnitruck.RequestParams) (url string, resp io.ReadCloser, header http.Header, msg string, code int, err error) {
	if s.Mode == constants.Commercial {
		return s.DownloadChefPlatform(params)
	}
	return "", nil, nil, "Platform Service does not support download in Open Source mode", http.StatusBadRequest, fmt.Errorf("%d Error: Platform Service does not support download in Open Source mode", http.StatusBadRequest)
}

func (s *PlatformServiceStrategy) DownloadChefPlatform(params *omnitruck.RequestParams) (url string, respBody io.ReadCloser, header http.Header, msg string, code int, err error) {
	resp := clients.Response{}
	request := s.LicenseClient.GetReplicatedCustomerEmail(params.LicenseId, s.LicenseServiceUrl, &resp)

	if !request.Ok {
		s.Log.Errorf("Received error response from getReplicatedCustomer")
		return "", nil, nil, request.Message, request.Code, fmt.Errorf("%d Error while fetching replicated customer email: %s", request.Code, request.Message)
	}

	var replicatedEmailResp clients.GetReplicatedCustomerResponse
	err = JsonUnmarshal(request.Body, &replicatedEmailResp)

	if err != nil {
		s.Log.Errorf("Error while unmarshalling getReplicatedCustomer response : %s", err.Error())
		return "", nil, nil, constants.UNMARSHAL_ERR_MSG, http.StatusInternalServerError, fmt.Errorf("%d Error while fetching replicated customer email: %s", request.Code, request.Message)
	}

	if replicatedEmailResp.StatusCode != http.StatusOK {
		s.Log.Errorf("Received error response from getReplicatedCustomer")
		return "", nil, nil, replicatedEmailResp.Message, replicatedEmailResp.StatusCode, fmt.Errorf("%d Error while fetching replicated customer email: %s", replicatedEmailResp.StatusCode, replicatedEmailResp.Message)
	}
	s.Log.Debug("Successfully fetched replicated customer email")

	//2. Run a search customer on replicated with email
	requestId := s.Locals["requestid"].(string)
	customers, err := s.Replicated.SearchCustomersByEmail(replicatedEmailResp.ReplicatedEmail, requestId)

	if err != nil {
		s.Log.Errorf("Error while fetching replicated customers with Email : %s", err.Error())
		return "", nil, nil, constants.REPLICATED_CUSTOMER_ERROR, http.StatusInternalServerError, fmt.Errorf("%d Error while fetching replicated customers with Email: %s", http.StatusInternalServerError, constants.REPLICATED_CUSTOMER_ERROR)
	}

	if len(customers) == 0 {
		s.Log.Errorf("No replicated customers found with Email : %s", replicatedEmailResp.ReplicatedEmail)
		return "", nil, nil, constants.REPLICATED_CUSTOMER_ERROR, http.StatusInternalServerError, fmt.Errorf("%d No replicated customers found with Email: %s", http.StatusInternalServerError, constants.REPLICATED_CUSTOMER_ERROR)
	}
	customer := customers[0]
	s.Log.Debug("Successfully fetched replicated customer details")

	//3. Based on Airgap flag, formulate the download URL
	url, err = s.Replicated.GetDowloadUrl(customer, requestId)
	if err != nil {
		s.Log.Errorf("Error while formulating download url from replicated : %s", err.Error())
		return "", nil, nil, constants.REPLICATED_DOWNLOAD_ERROR, http.StatusInternalServerError, fmt.Errorf("%d Error while formulating download url from replicated: %s", http.StatusInternalServerError, constants.REPLICATED_DOWNLOAD_ERROR)
	}
	s.Log.Info("Successfully formulated download url")

	s.Log.Info("Performing download from replicated")
	downloadResp, err := s.Replicated.DownloadFromReplicated(url, requestId, customer.InstallationId)
	if err != nil {
		s.Log.Errorf("Error while downloading from replicated : %s", err.Error())
		return "", nil, nil, constants.REPLICATED_DOWNLOAD_ERROR, http.StatusInternalServerError, fmt.Errorf("%d Error while downloading from replicated: %s", http.StatusInternalServerError, constants.REPLICATED_DOWNLOAD_ERROR)
	}
	s.Log.Info("Successfully downloaded from replicated")
	headers := downloadResp.Header
	headers.Set("Content-Disposition", constants.PLATFORM_SERVICE_CONTENT_DISPOSITION)
	return "", downloadResp.Body, headers, "", downloadResp.StatusCode, nil
}

func (s *PlatformServiceStrategy) GetFileName(params *omnitruck.RequestParams) (string, error) {
	return s.PlatformService.PlatformFilename(params, int(s.Mode))
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
