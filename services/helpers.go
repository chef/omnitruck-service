package services

import (
	"net/url"

	"github.com/chef/omnitruck-service/clients/omnitruck"
)

func buildEndpointUrl(baseUrl string, endpoint string, params *omnitruck.RequestParams) *url.URL {
	u, _ := url.Parse(baseUrl)
	path, _ := url.JoinPath(params.Channel, params.Product, endpoint)
	u.Path = path
	u.RawQuery = params.UrlParams().Encode()

	return u
}

func getDownloadUrl(params *omnitruck.RequestParams, c omnitruck.FiberContext) string {
	return buildEndpointUrl(c.BaseURL(), "download", params).String()
}

func getRequestParams(c omnitruck.FiberContext) *omnitruck.RequestParams {
	return &omnitruck.RequestParams{
		Channel:         c.Params("channel"),
		Product:         c.Params("product"),
		Version:         c.Query("v"),
		Platform:        c.Query("p"),
		PlatformVersion: c.Query("pv"),
		Architecture:    c.Query("m"),
		LicenseId:       c.Query("license_id"),
		Eol:             c.Query("eol", "false"),
	}
}
