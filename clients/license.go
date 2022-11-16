package clients

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const licenseApi = "https://licensing-acceptance.chef.co"

type License struct {
	client *http.Client
	log    *log.Entry
}

type RequestParams struct {
	LicenseId string
}

type Response struct {
	Data    bool   //json: "data"
	Message string //json: "message"
	Code    int    //json: "status_code"
}

func NewLicenseClient(l *log.Entry) *License {
	return &License{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: l.WithField("pkg", "client/license"),
	}
}

func (c *License) Get(url string) *Request {
	request := Request{
		Url: url,
	}

	req, err := http.NewRequest("GET", request.Url, nil)

	if err != nil {
		c.logRequestError("Error creating request", &request, err)
		return request.Failure(900, "Error creating request")
	}

	c.log.Infof("Fetching data from %s", url)
	req.Header.Add("Accept", "application/json")
	resp, err := c.client.Do(req)
	request.Code = resp.StatusCode

	if err != nil {
		c.logRequestError("Error fetching omnitruck data", &request, err)
		return request.Failure(request.Code, "Error fetching omnitruck data")
	}

	request.Body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logRequestError("Error reading response body from omnitruck api", &request, err)
		return request.Failure(900, "Error reading response body from omnitruck api")
	}

	if request.Code >= 400 {
		c.logRequestError(fmt.Sprintf("Omnitruck returned failure response %d", request.Code), &request, nil)
		// Set our response message to what the server responsed with
		// so we pass on the omnitruck error message to the user
		return request.Failure(request.Code, string(request.Body))
	}

	return request.Success()
}

func (c *License) logRequestError(msg string, request *Request, err error) {
	c.log.WithError(err).
		WithField("status", request.Code).
		WithField("body", string(request.Body)).
		Error(msg)
}

func (c *License) Validate(id string, data *Response) *Request {
	url := fmt.Sprintf("%s/License/v1/validate?licenseId=%s", licenseApi, id)
	return c.Get(url).ParseData(&data)
}

func (c *License) IsTrial(l string) bool {
	return strings.Contains(l, "tmns")
}
