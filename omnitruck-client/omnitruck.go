package omnitruck_client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const omnitruckApi = "https://omnitruck.chef.io"

type Omnitruck struct {
	client *http.Client
	log    *logrus.Entry
}

type RequestDataInterface interface {
}

type ItemList []string
type PlatformList map[string]string
type PackageList map[string]PlatformVersionList
type PlatformVersionList map[string]ArchList
type ArchList map[string]PackageMetadata
type ProductVersion string
type PackageMetadata struct {
	Sha1    string `json:"sha1"`
	Sha256  string `json:"sha256"`
	Url     string `json:"url"`
	Version string `json:"version"`
}

func NewOmnitruckClient() Omnitruck {
	return Omnitruck{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log.WithField("pkg", "client/omnitruck"),
	}
}

func (ot *Omnitruck) logRequestError(msg string, request *Request, err error) {
	ot.log.WithError(err).
		WithField("status", request.Code).
		WithField("body", string(request.Body)).
		Error(msg)
}

func (ot *Omnitruck) Get(url string) *Request {
	request := Request{
		Url: url,
	}

	req, err := http.NewRequest("GET", request.Url, nil)

	if err != nil {
		ot.logRequestError("Error creating request", &request, err)
		return request.Failure(900, "Error creating request")
	}

	ot.log.Infof("Fetching data from %s", url)
	req.Header.Add("Accept", "application/json")
	resp, err := ot.client.Do(req)
	request.Code = resp.StatusCode

	if err != nil {
		ot.logRequestError("Error fetching omnitruck data", &request, err)
		return request.Failure(request.Code, "Error fetching omnitruck data")
	}

	request.Body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		ot.logRequestError("Error reading response body from omnitruck api", &request, err)
		return request.Failure(900, "Error reading response body from omnitruck api")
	}

	if request.Code >= 400 {
		ot.logRequestError(fmt.Sprintf("Omnitruck returned failure response %d", request.Code), &request, nil)
		// Set our response message to what the server responsed with
		// so we pass on the omnitruck error message to the user
		return request.Failure(request.Code, string(request.Body))
	}

	return request.Success()
}

func (ot *Omnitruck) Products(p *RequestParams, data RequestDataInterface) *Request {
	url := fmt.Sprintf("%s/products", omnitruckApi)

	return ot.Get(url).ParseData(data)
}

func (ot *Omnitruck) Platforms() *Request {
	url := fmt.Sprintf("%s/platforms", omnitruckApi)

	return ot.Get(url)
}

func (ot *Omnitruck) Architectures() *Request {
	url := fmt.Sprintf("%s/architectures", omnitruckApi)

	return ot.Get(url)
}

func (ot *Omnitruck) LatestVersion(p *RequestParams) *Request {
	url := fmt.Sprintf("%s/%s/%s/versions/latest", omnitruckApi, p.Channel, p.Product)

	return ot.Get(url)
}

func (ot *Omnitruck) ProductVersions(p *RequestParams) *Request {
	url := fmt.Sprintf("%s/%s/%s/versions/all", omnitruckApi, p.Channel, p.Product)

	return ot.Get(url)
}

func (ot *Omnitruck) ProductPackages(p *RequestParams) *Request {
	url := fmt.Sprintf("%s/%s/%s/packages?v=%s", omnitruckApi, p.Channel, p.Product, p.Version)

	return ot.Get(url)
}

func (ot *Omnitruck) ProductMetadata(p *RequestParams) *Request {
	url := fmt.Sprintf("%s/%s/%s/metadata?v=%s&p=%s&pv=%s&m=%s", omnitruckApi,
		p.Channel,
		p.Product,
		p.Version,
		p.Platform,
		p.PlatformVersion,
		p.Architecture,
	)

	return ot.Get(url)
}

// Product Download needs to fetch the metadata record instead of the Omnitruck download API
// The Omnitruck API normall redirects the user to the download URL and we need to do this
// ourselves.
func (ot *Omnitruck) ProductDownload(p *RequestParams) *Request {
	return ot.ProductMetadata(p)
}
