package helpers

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
	"github.com/gofiber/fiber/v2"
)

const substring = ".metadata.json"

func BuildEndpointUrl(baseUrl string, endpoint string, params *omnitruck.RequestParams) *url.URL {
	clonedParams := *params
	if clonedParams.PackageManager == constants.DUMMY_PACKAGE_MANAGER {
		clonedParams.PackageManager = ""
	}
	u, _ := url.Parse(baseUrl)

	path, _ := url.JoinPath(clonedParams.Channel, clonedParams.Product, endpoint)
	u.Path = path
	u.RawQuery = clonedParams.UrlParams().Encode()

	return u
}

func GetDownloadUrl(params *omnitruck.RequestParams, baseUrl string) string {
	return BuildEndpointUrl(baseUrl, "download", params).String()
}

func GetRequestParams(c omnitruck.FiberContext) *omnitruck.RequestParams {
	return &omnitruck.RequestParams{
		Channel:         c.Params("channel"),
		Product:         c.Params("product"),
		Version:         c.Query("v"),
		Platform:        c.Query("p"),
		PlatformVersion: c.Query("pv"),
		Architecture:    c.Query("m"),
		PackageManager:  c.Query("pm"),
		LicenseId:       c.Query("license_id"),
		Eol:             c.Query("eol", "false"),
		BOM:             c.Query(("bom")),
	}
}

func VerifyRequestType(params *omnitruck.RequestParams) bool {
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

func ValidateOrSetVersion(params *omnitruck.RequestParams, filtered []omnitruck.ProductVersion) error {
	if params.Version == "" || params.Version == "latest" {
		// Use the latest version from filtered list if not provided
		params.Version = string(filtered[len(filtered)-1])
		return nil
	}

	// Iterate backwards to find the latest matching version efficiently
	requestedVersion := params.Version
	for i := len(filtered) - 1; i >= 0; i-- {
		vStr := string(filtered[i])
		// Check if version starts with the requested prefix (e.g., "16" matches "16.1.0", "16.2.5")
		if strings.HasPrefix(vStr, requestedVersion) {
			// Ensure it's a proper version boundary (e.g., "16" shouldn't match "160.0.0")
			if len(vStr) == len(requestedVersion) || vStr[len(requestedVersion)] == '.' {
				// Found the latest match, set and return immediately
				params.Version = vStr
				return nil
			}
		}
	}

	return fmt.Errorf("the requested version is not supported on the selected persona or channel")
}

func GetFileNameFromURL(url string) string {
	segments := strings.Split(url, "/")
	return segments[len(segments)-1]
}

func GetErrorCodeAndMsg(err error) (code int, msg string) {
	var fiberErr *fiber.Error

	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
		msg = fiberErr.Message
		return code, msg
	}
	return fiber.StatusInternalServerError, ""
}
