package strategy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	s3aws "github.com/chef/omnitruck-service/clients/omnitruck/aws"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	helpers "github.com/chef/omnitruck-service/internal/helper"
	"github.com/chef/omnitruck-service/utils"
	log "github.com/sirupsen/logrus"
)

type InfraProductStrategy struct {
	DynamoService    omnitruck.IDynamoServices
	AWSConfig        config.AWSConfig
	Log              *log.Entry
}

func (s *InfraProductStrategy) normalizePackageManager(params *omnitruck.RequestParams) error {
	provided := strings.ToLower(strings.TrimSpace(params.PackageManager))
	params.PackageManager = provided

	if params.PackageManager == "" {
		if params.Platform == "" {
			return fmt.Errorf(utils.PlatformParamsError)
		}
		derived := omnitruck.DerivePackageManager(params.Platform)
		if derived == "" {
			return fmt.Errorf("Unable to derive package manager for platform '%s'", params.Platform)
		}
		s.Log.WithField("platform", params.Platform).
			WithField("package_manager", derived).
			Debug("Auto-derived package manager from platform")
		params.PackageManager = derived
	}

	// Normalize platform to DB key (e.g. ubuntu → linux) after pm is resolved
	params.Platform = omnitruck.NormalizePlatformForDatabase(params.Platform)
	return nil
}

func (s *InfraProductStrategy) GetLatestVersion(params *omnitruck.RequestParams) (omnitruck.ProductVersion, *clients.Request) {
	request := clients.Request{}
	data, err := s.DynamoService.VersionLatest(params)
	if err != nil {
		code, msg := helpers.GetErrorCodeAndMsg(err)
		s.Log.WithError(err).Error("Error while fetching latest versions for "+params.Product+": ", err)
		request.Failure(code, msg)
		return data, &request
	}
	request.Success()
	return data, &request
}

func (s *InfraProductStrategy) GetAllVersions(params *omnitruck.RequestParams) ([]omnitruck.ProductVersion, *clients.Request) {
	request := clients.Request{}
	data, err := s.DynamoService.VersionAll(params)
	if err != nil {
		s.Log.WithError(err).Error("Error while fetching all versions for "+params.Product+": ", err)
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
		return nil, &request
	}

	// Sort versions before returning
	data = omnitruck.SortProductVersions(data)

	request.Success()
	return data, &request
}

func (s *InfraProductStrategy) Download(params *omnitruck.RequestParams) (url string, resp io.ReadCloser, header http.Header, msg string, code int, err error) {
	if err = s.normalizePackageManager(params); err != nil {
		return "", nil, nil, err.Error(), http.StatusBadRequest, err
	}
	fileName, err := s.DynamoService.GetFilename(params)
	if err != nil {
		s.Log.WithError(err).Error("Error while fetching download filename for "+params.Product+": ", err)
		code, msg := helpers.GetErrorCodeAndMsg(err)
		return "", nil, nil, msg, code, err
	}
	if fileName == "" {
		s.Log.Error("Download filename is empty for " + params.Product)
		return "", nil, nil, "Download filename is empty", http.StatusInternalServerError, nil
	}

	s.Log.Infof("Downloading file %s from S3 bucket %s in region %s", fileName, s.AWSConfig.S3Config.Bucket, s.AWSConfig.Region)
	return s.downloadFromS3(params, fileName)
}

func (s *InfraProductStrategy) downloadFromS3(params *omnitruck.RequestParams, fileName string) (url string, resp io.ReadCloser, header http.Header, msg string, code int, err error) {
	// Validate S3 configuration
	err = s3aws.ValidateS3Config(s.AWSConfig)
	if err != nil {
		s.Log.Errorf("Invalid S3 config found : %s ", err.Error())
		return "", nil, nil, err.Error(), http.StatusInternalServerError, nil
	}
	bucket := s.AWSConfig.S3Config.Bucket
	roleArn := s.AWSConfig.S3Config.RoleArn
	region := s.AWSConfig.Region
	path := ""
	if params.Channel == constants.CURRENT_CHANNEL {
		path = s.AWSConfig.S3Config.CurrentPath
	} else if params.Channel == constants.STABLE_CHANNEL {
		path = s.AWSConfig.S3Config.StablePath
	}

	//key should be formulated as path/product/version/platform/architecture/
	key := path + "/" + params.Product + "/" + params.Version + "/" + params.Platform + "/" + params.Architecture + "/" + fileName
	s.Log.Debugf("S3 key for download: %s", key)

	sess, err := s3aws.NewS3Session(region)
	if err != nil {
		s.Log.WithError(err).Error("Failed to create AWS session")
		return "", nil, nil, "Failed to create AWS session", http.StatusInternalServerError, err
	}
	creds := s3aws.NewS3Credentials(sess, roleArn)
	result, err := s3aws.GetS3Object(context.Background(), sess, creds, bucket, key)
	if err != nil {
		s.Log.WithError(err).Error("Failed to get object from S3")
		return "", nil, nil, "Failed to get object from S3", http.StatusInternalServerError, err
	}

	headers := http.Header{}
	if result.ContentType != nil {
		headers.Set("Content-Type", *result.ContentType)
	}
	if result.ContentLength != nil {
		headers.Set("Content-Length", strconv.FormatInt(*result.ContentLength, 10))

	}
	if result.ContentDisposition != nil {
		headers.Set("Content-Disposition", *result.ContentDisposition)
	} else {
		headers.Set("Content-Disposition", "attachment; filename="+fileName)
	}

	return "", result.Body, headers, "", 0, nil
}

func (s *InfraProductStrategy) GetPackages(params *omnitruck.RequestParams) (omnitruck.PackageList, error) {
	return s.DynamoService.ProductPackages(params)
}

func (s *InfraProductStrategy) GetMetadata(params *omnitruck.RequestParams) (omnitruck.PackageMetadata, *clients.Request) {
	request := &clients.Request{}
	if err := s.normalizePackageManager(params); err != nil {
		request.Failure(http.StatusBadRequest, err.Error())
		return omnitruck.PackageMetadata{}, request
	}
	data, err := s.DynamoService.ProductMetadata(params)
	if err != nil {
		s.Log.WithError(err).Error("Error while fetching metadata for "+params.Product+": ", err)
		code, msg := helpers.GetErrorCodeAndMsg(err)
		request.Failure(code, msg)
	} else {
		request.Success()
	}
	return data, request
}

func (s *InfraProductStrategy) GetFileName(params *omnitruck.RequestParams) (string, error) {
	if err := s.normalizePackageManager(params); err != nil {
		return "", err
	}
	fileName, err := s.DynamoService.GetFilename(params)
	return fileName, err
}

// ValidateFilesParams ensures that PackageManager is set for infra products,
// deriving from platform if not provided in URL.
func (s *InfraProductStrategy) ValidateFilesParams(params *omnitruck.RequestParams) error {
	// Normalize/derive package manager and platform
	if err := s.normalizePackageManager(params); err != nil {
		return fmt.Errorf(utils.PackageManagerParamsError)
	}

	s.Log.Infof("params after normalization: %+v", params)
	// Validate filename matches what database expects (infra products only)
	correctFileName, err := s.GetFileName(params)
	if err != nil {
		return err
	}
	if params.FileName != correctFileName {
		s.Log.Info("Filename validation failed for /files endpoint")
		return fmt.Errorf("invalid filename for the specified product parameters")
	}

	return nil
}

// ParseTail parses the /files URL tail in Infra format: {arch}/{pm}/{fileName}
// Expected format: arch/pm/filename (exactly 3 segments)
// PM is extracted from URL; normalizePackageManager will derive it if empty
func (s *InfraProductStrategy) ParseTail(segments []string) helpers.FilesPathParams {
	p := helpers.FilesPathParams{}

	switch len(segments) {
	case 0:
		return p
	case 1:
		p.FileName = segments[0]
	case 2:
		p.Architecture = segments[0]
		p.FileName = segments[1]
	default:
		// Expected: arch/pm/filename
		p.Architecture = segments[0]
		p.PackageManager = segments[1]
		p.FileName = segments[2]
	}

	return p
}

func (s *InfraProductStrategy) UpdatePackages(data *omnitruck.PackageList, params *omnitruck.RequestParams, baseUrl string) {
	data.UpdatePackages(func(platform string, arch string, packageManager string, m omnitruck.PackageMetadata) omnitruck.PackageMetadata {
		params.Version = m.Version
		params.Platform = platform
		params.PackageManager = packageManager
		params.Architecture = arch

		fileName := ""
		if params.Direct == "true" {
			if resolvedFileName, err := s.GetFileName(params); err == nil {
				fileName = resolvedFileName
			}
		}

		if strings.EqualFold(params.Direct, "true") && fileName != "" {
			// Infra products include the package manager segment in the /files URL path.
			m.Url = helpers.GetFilesUrl(params, baseUrl, fileName, packageManager)
		} else {
			m.Url = helpers.GetDownloadUrl(params, baseUrl)
		}
		return m
	})
}
