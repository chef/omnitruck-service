package omnitruck

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/utils"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

const omnitruckApi = "https://omnitruck.chef.io"

type Omnitruck struct {
	client *http.Client
	log    *log.Entry
}

type FiberContext interface {
	Params(string, ...string) string
	Query(string, ...string) string
	BaseURL() string
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

type RequestParams struct {
	Channel         string
	Product         string
	Version         string
	Platform        string
	PlatformVersion string
	Architecture    string
	Eol             string
	LicenseId       string
	BOM             string
}

type RequestParamsFlags struct {
	Channel         bool
	Product         bool
	Version         bool
	Platform        bool
	PlatformVersion bool
	Architecture    bool
	Eol             bool
	LicenseId       bool
	BOM             bool
}

type PackageListUpdater func(platform string, platformVersion string, arch string, meta PackageMetadata) PackageMetadata

func (pl PackageList) UpdatePackages(fn PackageListUpdater) {
	for platform, versions := range pl {
		for version, arches := range versions {
			for arch, metadata := range arches {
				pl[platform][version][arch] = fn(platform, version, arch, metadata)
			}
		}
	}

}

func (rp *RequestParams) UrlParams() url.Values {
	v := url.Values{}
	if len(rp.Version) > 0 {
		v.Add("v", rp.Version)
	}
	if len(rp.Platform) > 0 {
		v.Add("p", rp.Platform)
	}
	if len(rp.PlatformVersion) > 0 {
		v.Add("pv", rp.PlatformVersion)
	}
	if len(rp.Architecture) > 0 {
		v.Add("m", rp.Architecture)
	}
	if len(rp.Eol) > 0 {
		v.Add("eol", rp.Eol)
	}
	if len(rp.LicenseId) > 0 {
		v.Add("license_id", rp.LicenseId)
	}

	return v
}

func New(log *log.Entry) Omnitruck {
	return Omnitruck{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log.WithField("pkg", "client/omnitruck"),
	}
}

func (ot *Omnitruck) logRequestError(msg string, request *clients.Request, err error) {
	ot.log.WithError(err).
		WithField("status", request.Code).
		WithField("body", string(request.Body)).
		Error(msg)
}

func (ot *Omnitruck) Get(url string) *clients.Request {
	request := clients.Request{
		Url: url,
	}

	req, err := http.NewRequest("GET", request.Url, nil)

	if err != nil {
		ot.logRequestError("Error creating request", &request, err)
		return request.Failure(fiber.StatusBadRequest, utils.OmnitruckReqError)
	}

	ot.log.Infof("Fetching data from %s", url)
	req.Header.Add("Accept", "application/json")
	resp, err := ot.client.Do(req)
	request.Code = resp.StatusCode

	if err != nil {
		ot.logRequestError("Error fetching omnitruck data", &request, err)
		return request.Failure(request.Code, utils.OmnitruckApiError)
	}

	request.Body, err = io.ReadAll(resp.Body)
	if err != nil {
		ot.logRequestError("Error reading response body from omnitruck api", &request, err)
		return request.Failure(fiber.StatusBadRequest, utils.OmnitruckApiError)
	}

	if request.Code == 404 {
		ot.logRequestError(fmt.Sprintf("Omnitruck returned failure response %d", request.Code), &request, nil)
		// Set our response message to what the server responsed with
		// so we pass on the omnitruck error message to the user
		return request.Failure(fiber.StatusBadRequest, utils.OmnitruckDataNotFoundError)
	}

	if request.Code >= 400 {
		ot.logRequestError(fmt.Sprintf("Omnitruck returned failure response %d", request.Code), &request, nil)
		// Set our response message to what the server responsed with
		// so we pass on the omnitruck error message to the user
		return request.Failure(request.Code, string(request.Body))
	}

	return request.Success()
}

func (ot *Omnitruck) Products(p *RequestParams, data clients.RequestDataInterface) *clients.Request {
	url := fmt.Sprintf("%s/products", omnitruckApi)

	return ot.Get(url).ParseData(data)
}

func (ot *Omnitruck) Platforms() *clients.Request {
	url := fmt.Sprintf("%s/platforms", omnitruckApi)

	return ot.Get(url)
}

func (ot *Omnitruck) Architectures() *clients.Request {
	url := fmt.Sprintf("%s/architectures", omnitruckApi)

	return ot.Get(url)
}

func (ot *Omnitruck) LatestVersion(p *RequestParams) *clients.Request {
	flags := RequestParamsFlags{
		Channel: true,
	}

	err := ValidateRequest(p, flags)
	if !err.Ok {
		return err
	}
	url := fmt.Sprintf("%s/%s/%s/versions/latest", omnitruckApi, p.Channel, p.Product)

	return ot.Get(url)
}

func (ot *Omnitruck) ProductVersions(p *RequestParams) *clients.Request {
	flags := RequestParamsFlags{
		Channel: true,
	}

	err := ValidateRequest(p, flags)
	if !err.Ok {
		return err
	}
	url := fmt.Sprintf("%s/%s/%s/versions/all", omnitruckApi, p.Channel, p.Product)

	return ot.Get(url)
}

func (ot *Omnitruck) ProductPackages(p *RequestParams) *clients.Request {
	flags := RequestParamsFlags{
		Channel: true,
	}

	err := ValidateRequest(p, flags)
	if !err.Ok {
		return err
	}
	url := fmt.Sprintf("%s/%s/%s/packages?v=%s", omnitruckApi, p.Channel, p.Product, p.Version)

	return ot.Get(url)
}

func (ot *Omnitruck) ProductMetadata(p *RequestParams) *clients.Request {
	flags := RequestParamsFlags{
		Channel:         true,
		Platform:        true,
		PlatformVersion: true,
		Architecture:    true,
	}

	err := ValidateRequest(p, flags)
	if !err.Ok {
		return err
	}
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
func (ot *Omnitruck) ProductDownload(p *RequestParams) *clients.Request {
	return ot.ProductMetadata(p)
}

func ValidateRequest(p *RequestParams, flags RequestParamsFlags) *clients.Request {
	request := clients.Request{}

	if flags.Channel {
		if !(p.Channel == "stable" || p.Channel == "current") {
			request.Failure(fiber.StatusBadRequest, utils.ChannelParamsError)
			return &request
		}
	}
	if flags.Architecture {
		if p.Architecture == "" {
			request.Failure(fiber.StatusBadRequest, utils.ArchitectureParamsError)
			return &request
		}
	}
	if flags.BOM {
		if p.BOM == "" {
			request.Failure(fiber.StatusBadRequest, utils.BOMParamsError)
			return &request
		}
	}
	if flags.Platform {
		if p.Platform == "" {
			request.Failure(fiber.StatusBadRequest, utils.PlatformParamsError)
			return &request
		}
	}
	if flags.PlatformVersion {
		if p.PlatformVersion == "" {
			request.Failure(fiber.StatusBadRequest, utils.PlatformVersionParamsError)
			return &request
		}
	}

	request.Success()
	return &request
}
