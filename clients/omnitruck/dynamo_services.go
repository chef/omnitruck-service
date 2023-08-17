package omnitruck

import (
	"errors"
	"fmt"
	"sort"

	"github.com/chef/omnitruck-service/dboperations"
	"github.com/chef/omnitruck-service/models"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type DynamoServices struct {
	db  dboperations.IDbOperations
	log *logrus.Entry
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

func (svc *DynamoServices) Products(p []string, eol string) []string {
	p = append(p, "habitat")
	if eol == "true" {
		p = append(p, "automate-1")
	}
	sort.Strings(p)
	return p
}

func (svc *DynamoServices) Platforms(pl PlatformList) PlatformList {
	pl["linux"] = "Linux"
	pl["linux-kernel2"] = "Linux Kernel 2"
	pl["darwin"] = "Darwin"
	return pl
}

func (svc *DynamoServices) ProductDownload(p *RequestParams) (string, error) {
	var url string

	if p.Product == "automate" {
		p.Version = AUTOMATE_CLI_VERSION
		p.Channel = AUTOMATE_CHANNEL
		p.PlatformVersion = ""
	}

	details, err := svc.db.GetMetaData(p.Product, p.Version, p.Platform, p.PlatformVersion, p.Architecture)

	if err != nil {
		svc.log.WithError(err).Error("Error while fetching filename")
		return "", err
	}
	if *details == (models.MetaData{}) {
		return "", errors.New("Product information not found. Please check the input parameters")
	}

	switch p.Product {
	case "automate":
		url = fmt.Sprintf(DOWNLOAD_URL, p.Channel, p.Version, CHEF_AUTOMATE_CLI, details.FileName)
	}

	return url, nil
}

func (svc *DynamoServices) ProductMetadata(p *RequestParams) (PackageMetadata, error) {

	if p.Product == "automate" {
		p.Version = AUTOMATE_CLI_VERSION
		p.PlatformVersion = ""
	}
	details, err := svc.db.GetMetaData(p.Product, p.Version, p.Platform, p.PlatformVersion, p.Architecture)

	if err != nil {
		svc.log.WithError(err).Error("Error while fetching metadata")
		return PackageMetadata{}, err
	}
	if *details == (models.MetaData{}) {
		return PackageMetadata{}, errors.New("Product information not found. Please check the input parameters")
	}

	metadata := PackageMetadata{
		Url:     "",
		Sha1:    details.SHA1,
		Sha256:  details.SHA256,
		Version: p.Version,
	}
	return metadata, nil
}
