package replicated

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/logger"
	"github.com/chef/omnitruck-service/utils"
	"github.com/gofiber/fiber/v2"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ReplicatedImpl struct {
	ReplicatedConfig config.ReplicatedConfig
	Client           HTTPClient
	Logger           logger.Logger
}

func NewReplicatedImpl(config config.ReplicatedConfig, logger logger.Logger) IReplicated {
	return &ReplicatedImpl{
		Client:           &http.Client{},
		ReplicatedConfig: config,
		Logger:           logger,
	}
}

var ReadFile = func(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

var NewRequest = func(method string, url string, payload io.Reader) (res *http.Request, err error) {
	return http.NewRequest(method, url, payload)
}

func (r ReplicatedImpl) makeRequest(url, method, requestId string, payload io.Reader) (int, []byte, error) {
	log := utils.AddLogFields("makeRequest", requestId, r.Logger)

	req, err := NewRequest(method, url, payload)
	if err != nil {
		log.Errorln("error in creating new request.\n[ERROR] -", err.Error())
		return 0, nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", r.ReplicatedConfig.Token)

	res, err := r.Client.Do(req)
	if err != nil {
		log.Errorln("error on response.\n[ERROR] -", err.Error())
		return 0, nil, err
	}

	defer res.Body.Close()
	body, err := ReadFile(res.Body)
	if err != nil {
		log.Errorln("error on reading body.\n[ERROR] -", err.Error())
		return 0, nil, err
	}
	return res.StatusCode, body, nil
}

func (r ReplicatedImpl) SearchCustomersByEmail(email string, requestId string) (customers []Customer, err error) {
	log := utils.AddLogFields("SearchCustomersByEmail", requestId, r.Logger)

	url := fmt.Sprintf("%s/customers/search", r.ReplicatedConfig.URL)
	method := http.MethodPost

	payload := strings.NewReader(fmt.Sprintf(`{
	  "app_id": "%s",
	  "include_trial": true,
	  "include_community": true,
	  "include_paid": true,
	  "include_dev": true,
	  "include_active": true,
	  "include_inactive": true,
	  "query": "email:%s"
  	}`, r.ReplicatedConfig.AppID, email))

	respStatusCode, respBody, err := r.makeRequest(url, method, requestId, payload)
	if err != nil {
		log.Errorln("failed to search the customer: ", err.Error())
		return nil, err
	}

	if respStatusCode != http.StatusOK {
		err = fmt.Errorf("search customers by email failed with statusCode %d", respStatusCode)
		log.Errorln("error on search customers by email.\n[ERROR] -", err.Error())
		return nil, err
	}

	var respObj CustomerSearchResponse
	err = json.Unmarshal(respBody, &respObj)
	if err != nil {
		log.Errorln(constants.UNMARSHAL_ERR_MSG, err.Error())
		return nil, err
	}

	return respObj.Customers, nil
}

func (r *ReplicatedImpl) PlatformVersionsAll(req *omnitruck.RequestParams, serverMode int) ([]omnitruck.ProductVersion, error) {
	productVersions := []omnitruck.ProductVersion{}
	flags := omnitruck.RequestParamsFlags{
		Channel: true,
	}
	requestParams := omnitruck.ValidateRequest(req, flags)
	if !requestParams.Ok {
		r.Logger.Error("", requestParams.Message)
		return productVersions, fiber.NewError(requestParams.Code, requestParams.Message)
	}
	if serverMode == 2 && req.Product == constants.PLATFORM_SERVICE {
		productVersions = append(productVersions, "latest")
		return productVersions, nil
	}
	return productVersions, fiber.NewError(fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}

func (r *ReplicatedImpl) PlatformVersionLatest(req *omnitruck.RequestParams, serverMode int) (omnitruck.ProductVersion, error) {
	flags := omnitruck.RequestParamsFlags{
		Channel: true,
	}
	requestParams := omnitruck.ValidateRequest(req, flags)
	if !requestParams.Ok {
		r.Logger.Error("", requestParams.Message)
		return "", fiber.NewError(requestParams.Code, requestParams.Message)
	}
	if serverMode == 2 {
		return "latest", nil
	}
	return "", fiber.NewError(fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}

func (r *ReplicatedImpl) PlatformMetadata(req *omnitruck.RequestParams, serverMode int) (omnitruck.PackageMetadata, error) {
	flags := omnitruck.RequestParamsFlags{
		Channel: true,
	}
	requestParams := omnitruck.ValidateRequest(req, flags)
	if !requestParams.Ok {
		r.Logger.Error("", requestParams.Message)
		return omnitruck.PackageMetadata{}, fiber.NewError(requestParams.Code, requestParams.Message)
	}
	if serverMode == 2 && req.Product == constants.PLATFORM_SERVICE {
		return omnitruck.PackageMetadata{
			Sha1:    "",
			Sha256:  "",
			Url:     "",
			Version: req.Version,
		}, nil
	}
	return omnitruck.PackageMetadata{}, fiber.NewError(fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}

func (r *ReplicatedImpl) PlatformPackages(req *omnitruck.RequestParams, serverMode int) (omnitruck.PackageList, error) {
	packageList := omnitruck.PackageList{}
	flags := omnitruck.RequestParamsFlags{
		Channel: true,
	}
	requestParams := omnitruck.ValidateRequest(req, flags)
	if !requestParams.Ok {
		r.Logger.Error("", requestParams.Message)
		return omnitruck.PackageList{}, fiber.NewError(requestParams.Code, requestParams.Message)
	}
	if req.Version == "" {
		req.Version = "latest"
	}
	if serverMode == 2 && req.Product == constants.PLATFORM_SERVICE {
		packageList["linux"] = omnitruck.PlatformVersionList{}
		packageList["linux"]["pv"] = omnitruck.ArchList{}
		packageList["linux"]["pv"]["amd64"] = omnitruck.PackageMetadata{
			Sha1:    "",
			Sha256:  "",
			Url:     "",
			Version: req.Version,
		}
		return packageList, nil
	}
	return packageList, fiber.NewError(fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}

func (r *ReplicatedImpl) PlatformFilename(req *omnitruck.RequestParams, serverMode int) (string, error) {
	flags := omnitruck.RequestParamsFlags{
		Channel: true,
	}
	requestParams := omnitruck.ValidateRequest(req, flags)
	if !requestParams.Ok {
		r.Logger.Error("", requestParams.Message)
		return "", fiber.NewError(requestParams.Code, requestParams.Message)
	}
	if serverMode == 2 {
		return constants.PLATFORM_SERVICE + ".zip", nil
	}
	return "", fiber.NewError(fiber.StatusBadRequest, constants.PLATFORM_ERROR)
}
