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

// FilesPathParams holds the parsed components extracted from the /files URL path tail.
type FilesPathParams struct {
	PlatformVersion string
	Architecture    string
	PackageManager  string
	FileName        string
}

// TailParser is implemented by each ProductStrategy to parse the /files URL path tail
// according to that product's segment layout.
type TailParser interface {
	ParseTail(segments []string) FilesPathParams
}

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

// filesURLBuilder holds all components needed to construct a /files endpoint URL.
type filesURLBuilder struct {
	baseURL         string
	channel         string
	product         string
	version         string
	platform        string
	platformVersion string
	architecture    string
	packageManager  string
	fileName        string
	licenseID       string
}

func (b filesURLBuilder) build() string {
	u, _ := url.Parse(b.baseURL)
	segments := []string{"files", b.channel, b.product, b.version, b.platform}
	if b.platformVersion != "" {
		segments = append(segments, b.platformVersion)
	}
	segments = append(segments, b.architecture)
	if b.packageManager != "" && b.packageManager != constants.DUMMY_PACKAGE_MANAGER {
		segments = append(segments, b.packageManager)
	}
	segments = append(segments, b.fileName)
	path, _ := url.JoinPath("", segments...)
	u.Path = path
	q := url.Values{}
	q.Set("license_id", b.licenseID)
	u.RawQuery = q.Encode()
	return u.String()
}

// GetFilesUrl constructs a /files endpoint URL. pathPackageManager should be non-empty
// only when the package manager segment must appear in the path (e.g. infra products).
func GetFilesUrl(params *omnitruck.RequestParams, baseUrl string, fileName string, pathPackageManager string) string {
	return filesURLBuilder{
		baseURL:         baseUrl,
		channel:         params.Channel,
		product:         params.Product,
		version:         params.Version,
		platform:        params.Platform,
		platformVersion: params.PlatformVersion,
		architecture:    params.Architecture,
		packageManager:  pathPackageManager,
		fileName:        fileName,
		licenseID:       params.LicenseId,
	}.build()
}

// GetPackageUrl routes to /files when direct=true, otherwise to /download.
// For infra products, PM is included in the path when available (and not dummy value).
func GetPackageUrl(params *omnitruck.RequestParams, baseUrl string, fileName string) string {
	if strings.EqualFold(params.Direct, "true") && fileName != "" {
		// Only pass PM if it's not the dummy placeholder
		pm := ""
		if params.PackageManager != "" && params.PackageManager != constants.DUMMY_PACKAGE_MANAGER {
			pm = params.PackageManager
		}
		return GetFilesUrl(params, baseUrl, fileName, pm)
	}
	return GetDownloadUrl(params, baseUrl)
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
		Direct:          c.Query("direct"),
	}
}

// GetFilesRequestParamsWithStrategy parses the /files URL path tail using the
// strategy-owned parser and returns a fully populated RequestParams.
func GetFilesRequestParamsWithStrategy(c omnitruck.FiberContext, parser TailParser) *omnitruck.RequestParams {
	segments := strings.Split(strings.Trim(c.Params("*"), "/"), "/")
	parsed := parser.ParseTail(segments)

	return &omnitruck.RequestParams{
		Channel:         c.Params("channel"),
		Product:         c.Params("product"),
		Version:         c.Params("version"),
		Platform:        c.Params("platform"),
		PlatformVersion: parsed.PlatformVersion,
		Architecture:    parsed.Architecture,
		PackageManager:  parsed.PackageManager,
		FileName:        parsed.FileName,
		LicenseId:       c.Query("license_id"),
		Eol:             c.Query("eol", "false"),
	}
}

func ValidateCommonRequiredFilesParams(params *omnitruck.RequestParams) error {
	if params == nil {
		return errors.New("request params cannot be empty")
	}
	if strings.TrimSpace(params.Channel) == "" {
		return errors.New("Channel path param cannot be empty")
	}
	if strings.TrimSpace(params.Product) == "" {
		return errors.New("Product path param cannot be empty")
	}
	if strings.TrimSpace(params.Platform) == "" {
		return errors.New("Platform path param cannot be empty")
	}
	if strings.TrimSpace(params.FileName) == "" {
		return errors.New("Filename path param cannot be empty")
	}
	if strings.TrimSpace(params.Architecture) == "" {
		return errors.New("Architecture (m) params cannot be empty")
	}
	return nil
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