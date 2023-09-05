package clients

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type License struct {
	client *http.Client
}

type RequestParams struct {
	LicenseId string
}

type Response struct {
	Data    bool   //json: "data"
	Message string //json: "message"
	Code    int    //json: "status_code"
}

func NewLicenseClient() *License {
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
		return request.Failure(900, "Error creating request")
	}

	req.Header.Add("Accept", "application/json")
	resp, err := c.client.Do(req)
	request.Code = resp.StatusCode

	if err != nil {
		return request.Failure(request.Code, "Error fetching omnitruck data")
	}

	request.Body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return request.Failure(900, "Error reading response body from omnitruck api")
	}

	if request.Code >= 400 {
		// Set our response message to what the server responsed with
		// so we pass on the omnitruck error message to the user
		return request.Failure(request.Code, string(request.Body))
	}

	return request.Success()
}

func (c *License) Validate(id string, data *Response) *Request {
	licenseApi := os.Getenv("LICENSE_API")
	url := fmt.Sprintf("%s/License/v1/validate?licenseId=%s", licenseApi, id)
	return c.Get(url).ParseData(&data)
}

func (c *License) IsTrial(l string) bool {
	return strings.Contains(l, "tmns")
}
