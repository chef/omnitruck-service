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

	switch params.Product {
	case "automate":
		if params.Version == "" {
			params.Version = AUTOMATE_CLI_VERSION
		}
		params.Channel = AUTOMATE_CHANNEL
		params.PlatformVersion = ""
	case "habitat":
		if params.Version == "" || params.Version == "latest" {
			params.Version, err = svc.db.GetVersionLatest(params.Product)
			if err != nil {
				svc.log.WithError(err).Error("Error while fetching metadata")
				return "", err
			}
		}
		params.PlatformVersion = ""
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
	case "automate":
		url = fmt.Sprintf(DOWNLOAD_URL, params.Channel, params.Version, CHEF_AUTOMATE_CLI, details.FileName)
	case "habitat":
		url = fmt.Sprintf(DOWNLOAD_URL, params.Channel, params.Product, params.Version, details.FileName)
	}

	return url, nil
}

func (svc *DynamoServices) ProductMetadata(params *RequestParams) (PackageMetadata, error) {
	var err error
	version := ""
	switch params.Product {
	case "automate":
		if params.Version == "" {
			version = AUTOMATE_CLI_VERSION
		}
		params.PlatformVersion = ""
	case "habitat":
		if params.Version == "" || params.Version == "latest" {
			version, err = svc.db.GetVersionLatest(params.Product)
			if err != nil {
				svc.log.WithError(err).Error("Error while fetching metadata")
				return PackageMetadata{}, err
			}
		}
		params.PlatformVersion = ""
	}

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

	switch params.Product {
	case "automate":
		params.PlatformVersion = ""
	case "habitat":
		if params.Version == "" || params.Version == "latest" {
			params.Version, err = svc.db.GetVersionLatest(params.Product)
			if err != nil {
				svc.log.WithError(err).Error("Error while fetching metadata")
				return PackageList{}, err
			}
		}
	}

	details, err := svc.db.GetPackages(params.Product, params.Version)
	if err != nil {
		svc.log.WithError(err).Error("Error while fetching metadata")
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
