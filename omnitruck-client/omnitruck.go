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
	client *sling.Sling
	log    *logrus.Entry
}

type ResponseInterface interface {
}

type ProductList []string
type PlatformList map[string]string

func NewOmnitruckClient() Omnitruck {
	return Omnitruck{
		client: sling.New().Base(omnitruckApi),
		log:    log.WithField("pkg", "client/omnitruck"),
	}
}

func (ot *Omnitruck) Get(url string, data ResponseInterface) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return resp.StatusCode, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, data)

	return resp.StatusCode, err
}

func (ot *Omnitruck) Products() (int, ProductList, error) {
	url := fmt.Sprintf("%s/products", omnitruckApi)

	var products ProductList
	ot.log.Infof("fetching products from %s", url)
	code, err := ot.Get(url, &products)
	if err != nil {
		return code, products, err
	}

	return code, products, err
}

func (ot *Omnitruck) Platforms() (int, PlatformList, error) {
	url := fmt.Sprintf("%s/platforms", omnitruckApi)

	var platforms PlatformList

	ot.log.Infof("fetching products from %s", url)
	code, err := ot.Get(url, &platforms)
	if err != nil {
		return code, platforms, err
	}

	return code, platforms, err
}
