package omnitruck

import (
	"fmt"
	"sort"

	"github.com/chef/omnitruck-service/dboperations"
	"github.com/chef/omnitruck-service/models"
	"github.com/chef/omnitruck-service/utils"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

type DynamoServices struct {
	db  dboperations.IDbOperations
	log *log.Entry
}

const (
	DOWNLOAD_URL         = `https://packages.chef.io/files/%s/%s/%s/%s`
	CHEF_AUTOMATE_CLI    = "chef-automate-cli"
	AUTOMATE_CLI_VERSION = "latest"
	AUTOMATE_CHANNEL     = "current"
	AUTOMATE_PRODUCT     = "automate"
	HABITAT_PRODUCT      = "habitat"
	validating_log       = "Error while validating params:"
)

func NewDynamoServices(db dboperations.IDbOperations, log *log.Entry) DynamoServices {
	return DynamoServices{
		db:  db,
		log: log.WithField("pkg", "client/omnitruck"),
	}
}

func (svc *DynamoServices) Products(products []string, eol string) []string {
	products = append(products, "habitat")
	if eol == "true" {
		products = append(products, "automate-1")
	}
	sort.Strings(products)
	return products
}

func (svc *DynamoServices) Platforms(platforms PlatformList) PlatformList {
	platforms["linux"] = "Linux"
	platforms["linux-kernel2"] = "Linux Kernel 2"
	platforms["darwin"] = "Darwin"
	return platforms
}

func (svc *DynamoServices) ProductDownload(params *RequestParams) (string, error) {
	var url string
	var err error

	flags := RequestParamsFlags{
		Channel:      true,
		Platform:     true,
		Architecture: true,
	}

	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(validating_log, requestParams.Message)
		return "", fiber.NewError(requestParams.Code, requestParams.Message)
	}
	if params.Version == "" || params.Version == "latest" {
		params.Version, err = svc.db.GetVersionLatest(params.Product)
		if err != nil {
			svc.log.WithError(err).Error("Error while fetching latest version for download")
			return "", fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
		}
	}
	params.PlatformVersion = ""

	if params.Product == AUTOMATE_PRODUCT {
		params.Channel = AUTOMATE_CHANNEL
	}

	details, err := svc.db.GetMetaData(params.Product, params.Version, params.Platform, params.PlatformVersion, params.Architecture)

	if err != nil {
		svc.log.WithError(err).Error("Error while fetching filename")
		return "", fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
	}
	if *details == (models.MetaData{}) {
		return "", fiber.NewError(fiber.StatusBadRequest, utils.BadRequestError)
	}

	switch params.Product {
	case AUTOMATE_PRODUCT:
		url = fmt.Sprintf(DOWNLOAD_URL, params.Channel, params.Version, CHEF_AUTOMATE_CLI, details.FileName)
	case HABITAT_PRODUCT:
		url = fmt.Sprintf(DOWNLOAD_URL, params.Channel, params.Product, params.Version, details.FileName)
	}

	return url, nil
}

func (svc *DynamoServices) ProductMetadata(params *RequestParams) (PackageMetadata, error) {
	var err error
	version := params.Version

	flags := RequestParamsFlags{
		Channel:      true,
		Platform:     true,
		Architecture: true,
	}

	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(validating_log, requestParams.Message)
		return PackageMetadata{}, fiber.NewError(requestParams.Code, requestParams.Message)
	}

	if params.Version == "" || params.Version == "latest" {
		version, err = svc.db.GetVersionLatest(params.Product)
		if err != nil {
			svc.log.WithError(err).Error("Error while fetching latest version for metadata")
			return PackageMetadata{}, fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
		}
	}
	params.PlatformVersion = ""

	details, err := svc.db.GetMetaData(params.Product, version, params.Platform, params.PlatformVersion, params.Architecture)

	if err != nil {
		svc.log.WithError(err).Error("Error while fetching metadata")
		return PackageMetadata{}, fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
	}
	if *details == (models.MetaData{}) {
		return PackageMetadata{}, fiber.NewError(fiber.StatusBadRequest, utils.BadRequestError)
	}

	metadata := PackageMetadata{
		Url:     "",
		Sha1:    details.SHA1,
		Sha256:  details.SHA256,
		Version: version,
	}
	return metadata, nil
}

func (svc *DynamoServices) ProductPackages(params *RequestParams) (PackageList, error) {
	var err error
	flags := RequestParamsFlags{
		Channel: true,
	}

	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(validating_log, requestParams.Message)
		return PackageList{}, fiber.NewError(requestParams.Code, requestParams.Message)
	}

	if params.Version == "" || params.Version == "latest" {
		params.Version, err = svc.db.GetVersionLatest(params.Product)
		if err != nil {
			svc.log.WithError(err).Error("Error while fetching latest version for packages")
			return PackageList{}, fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
		}
	}

	details, err := svc.db.GetPackages(params.Product, params.Version)
	if err != nil {
		svc.log.WithError(err).Error("Error while fetching packages")
		return PackageList{}, fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
	}
	if len(details.MetaData) == 0 {
		return PackageList{}, fiber.NewError(fiber.StatusBadRequest, utils.BadRequestError)
	}

	packageList := PackageList{}
	for _, v := range details.MetaData {
		v.Platform_Version = "pv"
		if _, ok := packageList[v.Platform]; !ok {
			packageList[v.Platform] = PlatformVersionList{}
		}
		if _, ok := packageList[v.Platform][v.Platform_Version]; !ok {
			packageList[v.Platform][v.Platform_Version] = ArchList{}
		}
		if _, ok := packageList[v.Platform][v.Platform_Version][v.Architecture]; !ok {
			packageList[v.Platform][v.Platform_Version][v.Architecture] = PackageMetadata{
				Sha1:    v.SHA1,
				Sha256:  v.SHA256,
				Url:     "",
				Version: details.Version,
			}
		}
	}
	return packageList, nil
}

func (svc *DynamoServices) FetchLatestOsVersion(params *RequestParams) (string, error) {
	flags := RequestParamsFlags{
		Channel: true,
	}
	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(validating_log, requestParams.Message)
		return "", fiber.NewError(requestParams.Code, requestParams.Message)
	}

	var version string
	versions, err := svc.db.GetVersionAll(params.Product)
	if err != nil {
		svc.log.WithError(err).Error("Error while fetching the latest opensource version for the product.")
		return version, fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
	}

	if len(versions) == 0 {
		return version, fiber.NewError(fiber.StatusBadRequest, utils.BadRequestError)
	}
	sort.Strings(versions)
	if params.Product == HABITAT_PRODUCT {
		versions = FilterList(versions, func(v string) bool {
			return !OsProductVersion(params.Product, ProductVersion(v))
		})
	}

	// Return the last opensource version
	// This assumes the versions are returned in ascending order
	// Also assuming versions list is not empty
	version = versions[len(versions)-1]

	return version, nil
}

func (svc *DynamoServices) VersionAll(params *RequestParams) ([]ProductVersion, error) {
	productVersions := []ProductVersion{}
	flags := RequestParamsFlags{
		Channel: true,
	}
	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(validating_log, requestParams.Message)
		return productVersions, fiber.NewError(requestParams.Code, requestParams.Message)
	}

	versions, err := svc.db.GetVersionAll(params.Product)

	if err != nil {
		svc.log.WithError(err).Error("Error while fetching Versions")
		return productVersions, fiber.NewError(fiber.StatusInternalServerError, utils.FetchVersionsError)
	}
	if len(versions) == 0 {
		svc.log.Error("Recieved empty version list while fetching Versions")
		return productVersions, fiber.NewError(fiber.StatusBadRequest, utils.BadRequestError)
	}

	sort.Strings(versions)

	for _, version := range versions {
		productVersions = append(productVersions, ProductVersion(version))
	}
	return productVersions, nil
}

func (svc *DynamoServices) VersionLatest(params *RequestParams) (ProductVersion, error) {
	flags := RequestParamsFlags{
		Channel: true,
	}
	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(validating_log, requestParams.Message)
		return "", fiber.NewError(requestParams.Code, requestParams.Message)
	}

	version, err := svc.db.GetVersionLatest(params.Product)
	if err != nil {
		svc.log.WithError(err).Error("Error while fetching the latest version for the product.")
		return "", fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
	}

	return ProductVersion(version), nil
}

func (svc *DynamoServices) GetRelatedProducts(params *RequestParams) (*models.RelatedProducts, error) {
	var relatedProducts *models.RelatedProducts
	flags := RequestParamsFlags{
		BOM: true,
	}
	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(validating_log, requestParams.Message)
		return relatedProducts, fiber.NewError(requestParams.Code, requestParams.Message)
	}

	relatedProducts, err := svc.db.GetRelatedProducts(params.BOM)

	if err != nil {
		svc.log.WithError(err).Error("Error while fetching related products for " + params.BOM)
		//return relatedProducts, fiber.NewError(fiber.StatusInternalServerError, "Unable to retrieve related products for "+params.BOM)
		return relatedProducts, fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
	}

	if relatedProducts.Products == nil {
		svc.log.Error("No related products found for " + params.BOM)
		//return &models.RelatedProducts{}, fiber.NewError(fiber.StatusBadRequest, "No related products found for BOM")
		return relatedProducts, fiber.NewError(fiber.StatusBadRequest, utils.BadRequestError)
	}

	return relatedProducts, err
}

func (svc *DynamoServices) GetFilename(params *RequestParams) (string, error) {
	var err error
	version := params.Version

	flags := RequestParamsFlags{
		Channel:      true,
		Platform:     true,
		Architecture: true,
	}

	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(validating_log, requestParams.Message)
		return "", fiber.NewError(requestParams.Code, requestParams.Message)
	}

	if params.Version == "" || params.Version == "latest" {
		version, err = svc.db.GetVersionLatest(params.Product)
		if err != nil {
			svc.log.WithError(err).Error("Error while getting latest version for fetching fileName for " + params.Product)
			return "", fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
		}
	}
	params.PlatformVersion = ""

	details, err := svc.db.GetMetaData(params.Product, version, params.Platform, params.PlatformVersion, params.Architecture)

	if err != nil {
		svc.log.WithError(err).Error("Error while fetching fileName for " + params.Product)
		return "", fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
	}
	if details == nil || details.FileName == "" {
		svc.log.Error("Error while fetching fileName for " + params.Product + ":- unable to find the product information for given parameters")
		return "", fiber.NewError(fiber.StatusBadRequest, utils.BadRequestError)
	}

	return details.FileName, nil
}
