package omnitruck

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/chef/omnitruck-service/constants"
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

func NewDynamoServices(db dboperations.IDbOperations, log *log.Entry) DynamoServices {
	return DynamoServices{
		db:  db,
		log: log.WithField("pkg", "client/omnitruck"),
	}
}

func (svc *DynamoServices) SetDbInfo(table string, dbModelType reflect.Type) {
	// Setting the dyanomo table
	svc.db.SetDbInfo(table, dbModelType)
}

func (svc *DynamoServices) Products(products []string, eol string) []string {
	products = append(products, constants.HABITAT_PRODUCT, constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT, constants.MIGRATE_ICE)
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
		SampleAPI:    true,
	}

	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(constants.ERR_VALIDATING, requestParams.Message)
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

	if params.Product == constants.AUTOMATE_PRODUCT {
		params.Channel = constants.AUTOMATE_CHANNEL
	}

	details, err := svc.db.GetMetaData(params.Product, params.Version, params.Platform, params.PlatformVersion, params.Architecture, params.PackageManager)
	if err != nil {
		svc.log.WithError(err).Error("Error while fetching filename")
		return "", fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
	}
	if *details == (models.MetaData{}) {
		return "", fiber.NewError(fiber.StatusBadRequest, utils.BadRequestError)
	}

	switch params.Product {
	case constants.AUTOMATE_PRODUCT:
		url = fmt.Sprintf(constants.DOWNLOAD_URL, params.Channel, params.Version, constants.CHEF_AUTOMATE_CLI, details.FileName)
	case constants.HABITAT_PRODUCT:
		url = fmt.Sprintf(constants.DOWNLOAD_URL, params.Channel, params.Product, params.Version, details.FileName)
	}

	return url, nil
}

func (svc *DynamoServices) ProductMetadata(params *RequestParams) (PackageMetadata, error) {
	var err error
	version := params.Version

	flags := RequestParamsFlags{
		Channel:        true,
		Platform:       true,
		Architecture:   true,
		SampleAPI:      true,
		PackageManager: true,
	}

	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(constants.ERR_VALIDATING, requestParams.Message)
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

	details, err := svc.db.GetMetaData(params.Product, version, params.Platform, params.PlatformVersion, params.Architecture, params.PackageManager)
	if err != nil {
		svc.log.WithError(err).Error("Error while fetching metadata")
		return PackageMetadata{}, fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
	}
	if reflect.DeepEqual(*details, models.MetaData{}) {
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
	packageList := PackageList{}
	flags := RequestParamsFlags{
		Channel: true,
	}
	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(constants.ERR_VALIDATING, requestParams.Message)
		return PackageList{}, fiber.NewError(requestParams.Code, requestParams.Message)
	}

	if params.Version == "" || params.Version == "latest" {
		params.Version, err = svc.db.GetVersionLatest(params.Product)
		if err != nil {
			svc.log.WithError(err).Error("Error while fetching latest version for packages")
			return PackageList{}, fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
		}
	}

	data, err := svc.db.GetPackages(params.Product, params.Version)
	if err != nil {
		svc.log.WithError(err).Error("Error while fetching packages")
		return PackageList{}, fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
	}

	switch v := data.(type) {
	case *models.ProductDetails:
		if len(v.MetaData) == 0 {
			return PackageList{}, fiber.NewError(fiber.StatusBadRequest, utils.BadRequestError)
		}
		for _, meta := range v.MetaData {
			updatePackageList(packageList, meta.Platform, constants.PLATFORM_VERSION_KEY, meta.Architecture, PackageMetadata{
				Sha1:    meta.SHA1,
				Sha256:  meta.SHA256,
				Url:     "",
				Version: v.Version,
			})
		}
		return packageList, nil

	case *models.PackageDetails:
		for platform, archMap := range v.Metadata {
			for arch, pkgManagers := range archMap {
				for pkgMgr, pkg := range pkgManagers {
					updatePackageList(packageList, platform, arch, pkgMgr, PackageMetadata{
						Sha1:    pkg.SHA1,
						Sha256:  pkg.SHA256,
						Url:     "",
						Version: v.Version,
					})
				}
			}
		}
		return packageList, nil

	default:
		svc.log.Error(utils.ErrorLogUnsupportedPackageStructure)
		return nil, fiber.NewError(fiber.StatusInternalServerError, utils.ErrorMsgUnsupportedPackageStructure)
	}
}

// Utility to update nested map safely
func updatePackageList(pl PackageList, platform, versionKey, arch string, metadata PackageMetadata) {
	if pl[platform] == nil {
		pl[platform] = PlatformVersionList{}
	}
	if pl[platform][versionKey] == nil {
		pl[platform][versionKey] = ArchList{}
	}
	pl[platform][versionKey][arch] = metadata
}

func (svc *DynamoServices) FetchLatestOsVersion(params *RequestParams) (string, error) {
	flags := RequestParamsFlags{
		Channel: true,
	}
	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(constants.ERR_VALIDATING, requestParams.Message)
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
	if params.Product == constants.HABITAT_PRODUCT {
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
		svc.log.Error(constants.ERR_VALIDATING, requestParams.Message)
		return productVersions, fiber.NewError(requestParams.Code, requestParams.Message)
	}

	versions, err := svc.db.GetVersionAll(params.Product)

	if err != nil {
		svc.log.WithError(err).Error("Error while fetching Versions")
		return productVersions, fiber.NewError(fiber.StatusInternalServerError, utils.FetchVersionsError)
	}
	if len(versions) == 0 {
		svc.log.Error("Received empty version list while fetching Versions")
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
		svc.log.Error(constants.ERR_VALIDATING, requestParams.Message)
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
		svc.log.Error(constants.ERR_VALIDATING, requestParams.Message)
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
		Channel:        true,
		Platform:       true,
		Architecture:   true,
		SampleAPI:      true,
		PackageManager: true,
	}

	requestParams := ValidateRequest(params, flags)
	if !requestParams.Ok {
		svc.log.Error(constants.ERR_VALIDATING, requestParams.Message)
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

	details, err := svc.db.GetMetaData(params.Product, version, params.Platform, params.PlatformVersion, params.Architecture, params.PackageManager)
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

func (svc *DynamoServices) GetPackageManagers() ([]string, error) {
	result, err := svc.db.GetPackageManagers()
	if err != nil {
		svc.log.WithError(err).Error("Failed to fetch package managers from DB")
		return nil, fiber.NewError(fiber.StatusInternalServerError, utils.DBError)
	}
	return result, nil
}
