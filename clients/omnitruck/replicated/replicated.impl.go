package replicated

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/logger"
	"github.com/chef/omnitruck-service/utils"
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
		log.Errorln("error on response.\n[ERROR] - ", err.Error())
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

func (r ReplicatedImpl) DownloadFromReplicated(url, requestid, authorization string) (res *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", authorization)

	// Perform the request
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
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

func (r *ReplicatedImpl) GetDowloadUrl(customer Customer, requestId string) (url string, err error) {
	log := utils.AddLogFields("GetDowloadUrl", requestId, r.Logger)
	if len(customer.Channels) == 0 {
		log.Error("No channel found for download ")
		return "", fmt.Errorf("no channels found for customer %s", customer.ID)
	}

	channel := customer.Channels[0]

	if channel.AppSlug == "" || channel.ChannelSlug == "" {
		log.Error("Empty app or channel slug")
		return "", fmt.Errorf("empty app or channel slug found for customer %s", customer.ID)
	}
	url = constants.REPLICATED_DOWNLOAD_URL + "/" + channel.AppSlug + "/" + channel.ChannelSlug

	if customer.Airgap {
		url += "?airgap=true"
	}
	return url, nil
}
