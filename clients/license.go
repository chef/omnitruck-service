package clients

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chef/omnitruck-service/utils"
	"github.com/gofiber/fiber/v2"
)

type GetReplicatedCustomerResponse struct {
	ReplicatedEmail string `json:"replicatedEmail"`
	Message         string `json:"message"`
	StatusCode      string `json:"status_code"`
}
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
type License struct {
	client HTTPClient
}

type RequestParams struct {
	LicenseId string
}

type Response struct {
	Data    bool   //json: "data"
	Message string //json: "message"
	Code    int    //json: "status_code"
}

func NewLicenseClient() ILicense {
	return &License{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *License) Get(url string) *Request {
	request := Request{
		Url: url,
	}

	req, err := http.NewRequest("GET", request.Url, nil)

	if err != nil {
		return request.Failure(fiber.StatusBadRequest, utils.LicenseReqError)
	}

	req.Header.Add("Accept", "application/json")
	resp, err := c.client.Do(req)
	request.Code = resp.StatusCode

	if err != nil {
		return request.Failure(request.Code, utils.LicenseApiError)
	}

	request.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		return request.Failure(fiber.StatusBadRequest, utils.LicenseApiError)
	}

	if request.Code != 200 {
		// Set our response message to what the server responsed with
		// so we pass on the omnitruck error message to the user
		return request.Failure(request.Code, string(request.Body))
	}

	return request.Success()
}

func (c *License) Validate(id, licenseServiceUrl string, data *Response) *Request {
	licenseApi := licenseServiceUrl
	url := fmt.Sprintf("%s/v1/validate?licenseId=%s", licenseApi, id)
	return c.Get(url).ParseLicenseResp(&data)
}

func (c *License) GetReplicatedCustomerEmail(licenseId, licenseServiceUrl string, data *Response) *Request {
	requestUrl := fmt.Sprintf("%s/v1/getReplicatedCustomer?licenseId=%s", licenseServiceUrl, licenseId)
	return c.Get(requestUrl).ParseLicenseResp(&data)
}

func (c *License) IsTrial(l string) bool {
	return strings.Contains(l, "tmns")
}
