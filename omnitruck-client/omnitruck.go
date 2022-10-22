package omnitruck_client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const omnitruckApi = "https://omnitruck.chef.io"

type Omnitruck struct {
	successResponses []int
	client           *http.Client
	log              *logrus.Entry
}

type RequestParamsInterface interface {
	Get(string) string
}
type ResponseInterface interface {
}
type ResponseError string
type ItemList []string
type ProductList ItemList
type VersionList ItemList
type PlatformList map[string]string
type ArchitectureList ItemList
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
		successResponses: []int{200},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log.WithField("pkg", "client/omnitruck"),
	}
}

func (ot *Omnitruck) IsSuccess(code int) bool {
	for _, value := range ot.successResponses {
		if value == code {
			return true
		}
	}
	return false
}

func (ot *Omnitruck) handleRequestError(msg string, code int, body string, err error) {
	ot.log.WithError(err).
		WithField("status", code).
		WithField("body", body).
		Error(msg)
}

func (ot *Omnitruck) Get(url string, data ResponseInterface) (int, string, bool) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		ot.log.WithError(err).Error("Error creating request")
		// Return 900 if it was an error on our side vs remote issue
		return 900, "Internal error creating API request", false
	}

	ot.log.Infof("Fetching data from %s", url)

	req.Header.Add("Accept", "application/json")
	resp, err := ot.client.Do(req)
	if err != nil {
		ot.handleRequestError(
			"Error fetching omnitruck data",
			resp.StatusCode, "", err)
		return resp.StatusCode, "", false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ot.handleRequestError(
			"Error reading reponse body from omnitruck api",
			resp.StatusCode, string(body), err)
		return resp.StatusCode, "Error reading response body from Omnitruck API", false
	}

	if !ot.IsSuccess(resp.StatusCode) {
		ot.handleRequestError(
			fmt.Sprintf("Omnitruck API returned %d", resp.StatusCode),
			resp.StatusCode, string(body), err)
		return resp.StatusCode, string(body), false
	}

	err = json.Unmarshal(body, data)
	if err != nil {
		ot.handleRequestError(
			"Error parsing JSON response from Omnitruck",
			resp.StatusCode, string(body), err)
		return resp.StatusCode, "", false
	}

	return resp.StatusCode, "", true
}

func (ot *Omnitruck) Products() (int, ResponseInterface, bool) {
	url := fmt.Sprintf("%s/products", omnitruckApi)

	var data ProductList

	code, msg, success := ot.Get(url, &data)
	if success {
		return code, data, success
	} else {
		return code, msg, success
	}
}

func (ot *Omnitruck) Platforms() (int, ResponseInterface, bool) {
	url := fmt.Sprintf("%s/platforms", omnitruckApi)

	var data PlatformList

	code, msg, success := ot.Get(url, &data)
	if success {
		return code, data, success
	} else {
		return code, msg, success
	}
}

func (ot *Omnitruck) Architectures() (int, ResponseInterface, bool) {
	url := fmt.Sprintf("%s/architectures", omnitruckApi)

	var data ArchitectureList

	code, msg, success := ot.Get(url, &data)
	if success {
		return code, data, success
	} else {
		return code, msg, success
	}
}

func (ot *Omnitruck) LatestVersion(p RequestParamsInterface) (int, ResponseInterface, bool) {
	url := fmt.Sprintf("%s/%s/%s/versions/latest", omnitruckApi, p.Get("channel"), p.Get("product"))

	var data ProductVersion

	code, msg, success := ot.Get(url, &data)
	if success {
		return code, data, success
	} else {
		return code, msg, success
	}
}

func (ot *Omnitruck) ProductVersions(p RequestParamsInterface) (int, ResponseInterface, bool) {
	url := fmt.Sprintf("%s/%s/%s/versions/all", omnitruckApi, p.Get("channel"), p.Get("product"))

	var data VersionList

	code, msg, success := ot.Get(url, &data)
	if success {
		return code, data, success
	} else {
		return code, msg, success
	}
}

func (ot *Omnitruck) ProductPackages(p RequestParamsInterface) (int, ResponseInterface, bool) {
	url := fmt.Sprintf("%s/%s/%s/packages?v=%s", omnitruckApi, p.Get("channel"), p.Get("product"), p.Get("version"))

	var data PackageList

	code, msg, success := ot.Get(url, &data)
	if success {
		return code, data, success
	} else {
		return code, msg, success
	}
}

func (ot *Omnitruck) ProductMetadata(p RequestParamsInterface) (int, ResponseInterface, bool) {
	url := fmt.Sprintf("%s/%s/%s/metadata?v=%s&p=%s&pv=%s&m=%s", omnitruckApi,
		p.Get("channel"),
		p.Get("product"),
		p.Get("version"),
		p.Get("platform"),
		p.Get("platformVersion"),
		p.Get("architecture"),
	)

	var data PackageMetadata

	code, msg, success := ot.Get(url, &data)
	if success {
		return code, data, success
	} else {
		return code, msg, success
	}
}

func (ot *Omnitruck) ProductDownload(p RequestParamsInterface) (int, ResponseInterface, bool) {
	url := fmt.Sprintf("%s/%s/%s/metadata?v=%s&p=%s&pv=%s&m=%s", omnitruckApi,
		p.Get("channel"),
		p.Get("product"),
		p.Get("version"),
		p.Get("platform"),
		p.Get("platformVersion"),
		p.Get("architecture"),
	)

	var data PackageMetadata

	code, msg, success := ot.Get(url, &data)
	if success {
		return code, data, success
	} else {
		return code, msg, success
	}
}
