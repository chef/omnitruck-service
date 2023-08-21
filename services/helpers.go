package services

import (
	"net/url"
	"strings"

	"github.com/chef/omnitruck-service/clients/omnitruck"
)

const substring = ".metadata.json"

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

func verifyRequestType(params *omnitruck.RequestParams) bool {
	if strings.Contains(params.Architecture, substring) {
		params.Architecture = strings.Replace(params.Architecture, substring, "", 1)
		return true
	} else if strings.Contains(params.Platform, substring) {
		params.Platform = strings.Replace(params.Platform, substring, "", 1)
		return true
	} else if strings.Contains(params.PlatformVersion, substring) {
		params.PlatformVersion = strings.Replace(params.PlatformVersion, substring, "", 1)
		return true
	} else if strings.Contains(params.Eol, substring) {
		params.Eol = strings.Replace(params.Eol, substring, "", 1)
		return true
	} else if strings.Contains(params.Version, substring) {
		params.Version = strings.Replace(params.Version, substring, "", 1)
		return true
	}
	return false
}
