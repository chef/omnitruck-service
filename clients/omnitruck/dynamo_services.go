package omnitruck

import (
	"fmt"
	"sort"

	"github.com/chef/omnitruck-service/dboperations"
	"github.com/chef/omnitruck-service/models"
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

	if params.Version == "" || params.Version == "latest" {
		params.Version, err = svc.db.GetVersionLatest(params.Product)
		if err != nil {
			svc.log.WithError(err).Error("Error while fetching metadata")
			return "", err
		}
	}
	params.PlatformVersion = ""

	if params.Product == AUTOMATE_PRODUCT {
		params.Channel = AUTOMATE_CHANNEL
	}

	details, err := svc.db.GetMetaData(params.Product, params.Version, params.Platform, params.PlatformVersion, params.Architecture)

	if err != nil {
		svc.log.WithError(err).Error("Error while fetching filename")
		return "", err
	}
	if *details == (models.MetaData{}) {
		return "", nil
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

	if params.Version == "" || params.Version == "latest" {
		version, err = svc.db.GetVersionLatest(params.Product)
		if err != nil {
			svc.log.WithError(err).Error("Error while latest version for fetching metadata")
			return PackageMetadata{}, err
		}
	}
	params.PlatformVersion = ""

	details, err := svc.db.GetMetaData(params.Product, version, params.Platform, params.PlatformVersion, params.Architecture)

	if err != nil {
		svc.log.WithError(err).Error("Error while fetching metadata")
		return PackageMetadata{}, err
	}
	if *details == (models.MetaData{}) {
		return PackageMetadata{}, nil
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

	if params.Version == "" || params.Version == "latest" {
		params.Version, err = svc.db.GetVersionLatest(params.Product)
		if err != nil {
			svc.log.WithError(err).Error("Error while fetching latest version for packages")
			return PackageList{}, err
		}
	}

	details, err := svc.db.GetPackages(params.Product, params.Version)
	if err != nil {
		svc.log.WithError(err).Error("Error while fetching packages")
		return PackageList{}, err
	}
	if len(details.MetaData) == 0 {
		return PackageList{}, nil
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
	var version string
	versions, err := svc.db.GetVersionAll(params.Product)
	if err != nil {
		svc.log.WithError(err).Error("Error while fetching all versions")
		return version, err
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
	versions, err := svc.db.GetVersionAll(params.Product)

	if err != nil {
		svc.log.WithError(err).Error("Error while fetching Versions")
		return []ProductVersion{}, err
	}
	if len(versions) == 0 {
		svc.log.Warn("Recieved empty version list while fetching Versions")
	}
	productVersions := []ProductVersion{}
	sort.Strings(versions)

	for _, version := range versions {
		productVersions = append(productVersions, ProductVersion(version))
	}
	return productVersions, nil
}

func (svc *DynamoServices) VersionLatest(params *RequestParams) (ProductVersion, error) {

	version, err := svc.db.GetVersionLatest(params.Product)
	if err != nil {
		svc.log.WithError(err).Error("Error while fetching Versions")
		return "", err
	}

	return ProductVersion(version), nil

}
