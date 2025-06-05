package services

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/gofiber/fiber/v2"
)

const substring = ".metadata.json"

func buildEndpointUrl(baseUrl string, endpoint string, params *omnitruck.RequestParams) *url.URL {
	u, _ := url.Parse(baseUrl)
	path, _ := url.JoinPath(params.Channel, params.Product, endpoint)
	u.Path = path
	u.RawQuery = params.UrlParams().Encode()

	return u
}

func getDownloadUrl(params *omnitruck.RequestParams, baseUrl string) string {
	return buildEndpointUrl(baseUrl, "download", params).String()
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
		BOM:             c.Query(("bom")),
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

func validateOrSetVersion(params *omnitruck.RequestParams, filtered []omnitruck.ProductVersion) error {
	if params.Version != "" && params.Version != "latest" {
		for _, v := range filtered {
			if string(v) == params.Version {
				return nil
			}
		}
		return fmt.Errorf("requested version is not latest")
	}
	// Use the latest version from filtered list if not provided
	params.Version = string(filtered[len(filtered)-1])
	return nil
}

func getFileNameFromURL(url string) string {
	segments := strings.Split(url, "/")
	return segments[len(segments)-1]
}

func getErrorCodeAndMsg(err error) (code int, msg string) {
	var fiberErr *fiber.Error

	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
		msg = fiberErr.Message
		return code, msg
	}
	return fiber.StatusInternalServerError, ""
}
