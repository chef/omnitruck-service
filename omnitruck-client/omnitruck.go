package omnitruck_client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dghubble/sling"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const omnitruckApi = "https://omnitruck.chef.io"

type Omnitruck struct {
	successResponses []int
	client           *sling.Sling
	log              *logrus.Entry
}

type ResponseInterface interface {
}

type ProductList []string
type PlatformList map[string]string
type ArchitectureList []string

func NewOmnitruckClient() Omnitruck {
	return Omnitruck{
		successResponses: []int{200},
		client:           sling.New().Base(omnitruckApi),
		log:              log.WithField("pkg", "client/omnitruck"),
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

func (ot *Omnitruck) Get(url string, data ResponseInterface) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		ot.log.WithError(err).Error("Error fetching omnitruck data")
		return resp.StatusCode, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ot.log.WithError(err).Error("Error reading omnitruck response")
	}

	if !ot.IsSuccess(resp.StatusCode) {
		return resp.StatusCode, fmt.Errorf("%s", body)
	}

	err = json.Unmarshal(body, data)
	if err != nil {
		err = fmt.Errorf("Error parsing omnitruck response: %s\n%w", body, err)
		ot.log.WithError(err)
	}

	return resp.StatusCode, err
}

func (ot *Omnitruck) Products() (int, ProductList, error) {
	url := fmt.Sprintf("%s/products", omnitruckApi)

	var products ProductList
	ot.log.Infof("fetching products from %s", url)
	code, err := ot.Get(url, &products)

	return code, products, err
}

func (ot *Omnitruck) Platforms() (int, PlatformList, error) {
	url := fmt.Sprintf("%s/platforms", omnitruckApi)

	var platforms PlatformList

	ot.log.Infof("fetching platforms from %s", url)
	code, err := ot.Get(url, &platforms)

	return code, platforms, err
}

func (ot *Omnitruck) Architectures() (int, ArchitectureList, error) {
	url := fmt.Sprintf("%s/architectures", omnitruckApi)

	var archs ArchitectureList

	ot.log.Infof("fetching architectures from %s", url)
	code, err := ot.Get(url, &archs)

	return code, archs, err
}

func (ot *Omnitruck) LatestVersion(channel string, product string) (int, string, error) {
	url := fmt.Sprintf("%s/%s/%s/versions/latest", omnitruckApi, channel, product)

	var version string

	ot.log.Infof("fetching latest version from %s", url)
	code, err := ot.Get(url, &version)

	return code, version, err
}
