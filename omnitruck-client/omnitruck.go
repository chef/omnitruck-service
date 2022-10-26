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

type RequestParamsInterface interface {
	Get(string) string
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

func (ot *Omnitruck) Products(p RequestParamsInterface, data RequestDataInterface) *Request {
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

func (ot *Omnitruck) LatestVersion(p RequestParamsInterface) *Request {
	url := fmt.Sprintf("%s/%s/%s/versions/latest", omnitruckApi, p.Get("channel"), p.Get("product"))

	return ot.Get(url)
}

func (ot *Omnitruck) ProductVersions(p RequestParamsInterface) *Request {
	url := fmt.Sprintf("%s/%s/%s/versions/all", omnitruckApi, p.Get("channel"), p.Get("product"))

	return ot.Get(url)
}

func (ot *Omnitruck) ProductPackages(p RequestParamsInterface) *Request {
	url := fmt.Sprintf("%s/%s/%s/packages?v=%s", omnitruckApi, p.Get("channel"), p.Get("product"), p.Get("version"))

	return ot.Get(url)
}

func (ot *Omnitruck) ProductMetadata(p RequestParamsInterface) *Request {
	url := fmt.Sprintf("%s/%s/%s/metadata?v=%s&p=%s&pv=%s&m=%s", omnitruckApi,
		p.Get("channel"),
		p.Get("product"),
		p.Get("version"),
		p.Get("platform"),
		p.Get("platformVersion"),
		p.Get("architecture"),
	)

	return ot.Get(url)
}

func (ot *Omnitruck) ProductDownload(p RequestParamsInterface) *Request {
	url := fmt.Sprintf("%s/%s/%s/metadata?v=%s&p=%s&pv=%s&m=%s", omnitruckApi,
		p.Get("channel"),
		p.Get("product"),
		p.Get("version"),
		p.Get("platform"),
		p.Get("platformVersion"),
		p.Get("architecture"),
	)

	return ot.Get(url)
}
