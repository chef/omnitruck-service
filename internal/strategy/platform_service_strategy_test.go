package strategy_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"testing"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/models"
	"github.com/gofiber/fiber/v2"

	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/internal/strategy"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPlatformServiceStrategy_GetLatestVersion(t *testing.T) {
	mockPlatform := &omnitruck.MockPlatformServices{
		PlatformVersionLatestFunc: func(params *omnitruck.RequestParams, mode int) (omnitruck.ProductVersion, error) {
			return "3.1.0", nil
		},
	}

	s := &strategy.PlatformServiceStrategy{
		PlatformService: mockPlatform,
		Log:             log.NewEntry(log.New()),
		Mode:            constants.Commercial,
	}
	version, req := s.GetLatestVersion(&omnitruck.RequestParams{})
	assert.Equal(t, "3.1.0", string(version))
	assert.True(t, req.Ok)
}

func TestPlatformServiceStrategy_GetAllVersions(t *testing.T) {
	mockPlatform := &omnitruck.MockPlatformServices{
		PlatformVersionsAllFunc: func(params *omnitruck.RequestParams, mode int) ([]omnitruck.ProductVersion, error) {
			return []omnitruck.ProductVersion{"1.0.0", "2.0.0"}, nil
		},
	}
	s := &strategy.PlatformServiceStrategy{
		PlatformService: mockPlatform,
		Log:             log.NewEntry(log.New()),
		Mode:            constants.Commercial,
	}
	versions, req := s.GetAllVersions(&omnitruck.RequestParams{})
	assert.Len(t, versions, 2)
	assert.True(t, req.Ok)
}

func TestPlatformServiceStrategy_GetPackages(t *testing.T) {
	mockPlatform := &omnitruck.MockPlatformServices{
		PlatformPackagesFunc: func(params *omnitruck.RequestParams, mode int) (omnitruck.PackageList, error) {
			return omnitruck.PackageList{}, nil
		},
	}
	s := &strategy.PlatformServiceStrategy{
		PlatformService: mockPlatform,
	}
	_, err := s.GetPackages(&omnitruck.RequestParams{})
	assert.NoError(t, err)
}

func TestPlatformServiceStrategy_GetMetadata(t *testing.T) {
	mockPlatform := &omnitruck.MockPlatformServices{
		PlatformMetadataFunc: func(params *omnitruck.RequestParams, mode int) (omnitruck.PackageMetadata, error) {
			return omnitruck.PackageMetadata{Version: "3.2.1"}, nil
		},
	}
	s := &strategy.PlatformServiceStrategy{
		PlatformService: mockPlatform,
		Log:             log.NewEntry(log.New()),
	}
	meta, req := s.GetMetadata(&omnitruck.RequestParams{})
	assert.Equal(t, "3.2.1", meta.Version)
	assert.True(t, req.Ok)
}

func TestPlatformServiceStrategy_Download_OpenSource(t *testing.T) {
	s := &strategy.PlatformServiceStrategy{
		Mode: constants.Opensource,
	}
	url, resp, header, msg, code, err := s.Download(&omnitruck.RequestParams{})
	assert.Equal(t, "Platform Service does not support download in Open Source mode", msg)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Error(t, err)
	assert.Empty(t, url)
	assert.Nil(t, resp)
	assert.Nil(t, header)
}

func TestPlatformServiceStrategy_DownloadChefPlatform_Success(t *testing.T) {
	mockLicense := &clients.MockLicense{
		GetReplicatedCustomerEmailFunc: func(licenseID, url string, resp *clients.Response) *clients.Request {
			return &clients.Request{
				Ok:      true,
				Code:    200,
				Message: "success",
				Body:    []byte(`{"replicatedEmail":"test@example.com","status_code":200,"message":"success"}`),
			}
		},
	}

	mockReplicated := &replicated.MockReplicated{
		SearchCustomersByEmailFunc: func(email, requestId string) ([]models.Customer, error) {
			return []models.Customer{{InstallationId: "install123"}}, nil
		},
		GetDowloadUrlFunc: func(customer models.Customer, requestId string) (string, error) {
			return "http://example.com/download", nil
		},
		DownloadFromReplicatedFunc: func(url, requestId, auth string) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte("ok"))),
				Header:     http.Header{"Content-Disposition": []string{constants.PLATFORM_SERVICE_CONTENT_DISPOSITION}},
			}, nil
		},
	}

	s := &strategy.PlatformServiceStrategy{
		Mode:              constants.Commercial,
		LicenseClient:     mockLicense,
		Replicated:        mockReplicated,
		LicenseServiceUrl: "http://licenseservice",
		Log:               log.NewEntry(log.New()),
	}
	s.SetLocals(map[string]interface{}{
		"requestId": "req123",
	})

	url, resp, headers, msg, code, err := s.DownloadChefPlatform(&omnitruck.RequestParams{LicenseId: "lic123"})

	assert.NoError(t, err)
	assert.Equal(t, 200, code)
	assert.NotNil(t, resp)
	assert.Equal(t, "", msg)
	assert.Equal(t, constants.PLATFORM_SERVICE_CONTENT_DISPOSITION, headers.Get("Content-Disposition"))
	assert.Equal(t, "", url)
}

func TestPlatformServiceStrategy_GetFileName(t *testing.T) {
	mockPlatform := &omnitruck.MockPlatformServices{
		PlatformFilenameFunc: func(params *omnitruck.RequestParams, mode int) (string, error) {
			return "platformservice.tar.gz", nil
		},
	}
	s := &strategy.PlatformServiceStrategy{
		PlatformService: mockPlatform,
		Mode:            constants.Commercial,
	}
	filename, err := s.GetFileName(&omnitruck.RequestParams{})
	assert.NoError(t, err)
	assert.Equal(t, "platformservice.tar.gz", filename)
}

func TestPlatformServiceStrategy_UpdatePackages(t *testing.T) {
	baseUrl := "http://example.com"
	params := &omnitruck.RequestParams{
		Product:      "platform-service",
		Version:      "1.2.3",
		Platform:     "linux",
		Architecture: "amd64",
	}

	packageList := omnitruck.PackageList{
		"linux": omnitruck.PlatformVersionList{
			"pv": omnitruck.ArchList{
				"amd64": omnitruck.PackageMetadata{
					Version: "1.2.3",
					Url:     "",
				},
			},
		},
	}

	s := &strategy.PlatformServiceStrategy{}

	s.UpdatePackages(&packageList, params, baseUrl)

	updated := packageList["linux"]["pv"]["amd64"]
	assert.Contains(t, updated.Url, baseUrl)
	assert.Equal(t, "1.2.3", updated.Version)
	assert.Equal(t, "linux", params.Platform)
	assert.Equal(t, "amd64", params.Architecture)
}

func TestPlatformServiceStrategy_GetLatestVersion_Error(t *testing.T) {
	mockPlatform := &omnitruck.MockPlatformServices{
		PlatformVersionLatestFunc: func(params *omnitruck.RequestParams, mode int) (omnitruck.ProductVersion, error) {
			return "", fiber.NewError(http.StatusInternalServerError, "db error")
		},
	}
	s := &strategy.PlatformServiceStrategy{
		PlatformService: mockPlatform,
		Log:             log.NewEntry(log.New()),
		Mode:            constants.Commercial,
	}
	version, req := s.GetLatestVersion(&omnitruck.RequestParams{})
	assert.Empty(t, version)
	assert.False(t, req.Ok)
	assert.Contains(t, req.Message, "db error")
}

func TestPlatformServiceStrategy_GetAllVersions_Error(t *testing.T) {
	mockPlatform := &omnitruck.MockPlatformServices{
		PlatformVersionsAllFunc: func(params *omnitruck.RequestParams, mode int) ([]omnitruck.ProductVersion, error) {
			return nil, fiber.NewError(http.StatusInternalServerError, "db error")
		},
	}
	s := &strategy.PlatformServiceStrategy{
		PlatformService: mockPlatform,
		Log:             log.NewEntry(log.New()),
		Mode:            constants.Commercial,
	}
	versions, req := s.GetAllVersions(&omnitruck.RequestParams{})
	assert.Nil(t, versions)
	assert.False(t, req.Ok)
	assert.Contains(t, req.Message, "db error")
}

func TestPlatformServiceStrategy_GetMetadata_Error(t *testing.T) {
	mockPlatform := &omnitruck.MockPlatformServices{
		PlatformMetadataFunc: func(params *omnitruck.RequestParams, mode int) (omnitruck.PackageMetadata, error) {
			return omnitruck.PackageMetadata{}, fiber.NewError(http.StatusInternalServerError, "db error")
		},
	}
	s := &strategy.PlatformServiceStrategy{
		PlatformService: mockPlatform,
		Log:             log.NewEntry(log.New()),
		Mode:            constants.Commercial,
	}
	meta, req := s.GetMetadata(&omnitruck.RequestParams{})
	assert.Empty(t, meta.Version)
	assert.False(t, req.Ok)
	assert.Contains(t, req.Message, "db error")
}

func TestPlatformServiceStrategy_DownloadChefPlatform_Errors(t *testing.T) {
	tests := []struct {
		name                  string
		requestOk             bool
		requestBody           []byte
		requestCode           int
		replicatedStatusCode  int
		replicatedMessage     string
		searchCustomersError  error
		customers             []models.Customer
		getDownloadUrlError   error
		downloadReplicatedErr error
		expectedMsg           string
		expectedCode          int
	}{
		{
			name:         "request not ok",
			requestOk:    false,
			requestCode:  400,
			expectedMsg:  "request failed",
			expectedCode: 400,
		},
		{
			name:         "invalid json in body",
			requestOk:    true,
			requestCode:  200,
			requestBody:  []byte("invalid"),
			expectedMsg:  constants.UNMARSHAL_ERR_MSG,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:                 "replicated email response status code not 200",
			requestOk:            true,
			requestCode:          200,
			requestBody:          []byte(`{"replicatedEmail":"test@example.com","status_code":400,"message":"failed"}`),
			replicatedStatusCode: 400,
			replicatedMessage:    "failed",
			expectedMsg:          "failed",
			expectedCode:         400,
		},

		{
			name:                 "error searching customers by email",
			requestOk:            true,
			requestCode:          200,
			requestBody:          []byte(`{"replicatedEmail":"test@example.com","status_code":200}`),
			replicatedStatusCode: 200,
			searchCustomersError: fmt.Errorf("some error"),
			expectedMsg:          constants.REPLICATED_CUSTOMER_ERROR,
			expectedCode:         http.StatusInternalServerError,
		},
		{
			name:                 "no replicated customers found",
			requestOk:            true,
			requestCode:          200,
			requestBody:          []byte(`{"replicatedEmail":"test@example.com","status_code":200}`),
			replicatedStatusCode: 200,
			customers:            []models.Customer{},
			expectedMsg:          constants.REPLICATED_CUSTOMER_ERROR,
			expectedCode:         http.StatusInternalServerError,
		},
		{
			name:                 "error from GetDownloadUrl",
			requestOk:            true,
			requestCode:          200,
			requestBody:          []byte(`{"replicatedEmail":"test@example.com","status_code":200}`),
			replicatedStatusCode: 200,
			customers:            []models.Customer{{InstallationId: "id123"}},
			getDownloadUrlError:  fmt.Errorf("url error"),
			expectedMsg:          constants.REPLICATED_DOWNLOAD_ERROR,
			expectedCode:         http.StatusInternalServerError,
		},
		{
			name:                  "error from DownloadFromReplicated",
			requestOk:             true,
			requestCode:           200,
			requestBody:           []byte(`{"replicatedEmail":"test@example.com","status_code":200}`),
			replicatedStatusCode:  200,
			customers:             []models.Customer{{InstallationId: "id123"}},
			downloadReplicatedErr: fmt.Errorf("download error"),
			expectedMsg:           constants.REPLICATED_DOWNLOAD_ERROR,
			expectedCode:          http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLicense := &clients.MockLicense{
				GetReplicatedCustomerEmailFunc: func(licenseID, url string, resp *clients.Response) *clients.Request {
					return &clients.Request{
						Ok:      tt.requestOk,
						Code:    tt.requestCode,
						Message: "request failed",
						Body:    tt.requestBody,
					}
				},
			}
			mockReplicated := &replicated.MockReplicated{
				SearchCustomersByEmailFunc: func(email, reqID string) ([]models.Customer, error) {
					return tt.customers, tt.searchCustomersError
				},
				GetDowloadUrlFunc: func(customer models.Customer, reqID string) (string, error) {
					return "http://example.com", tt.getDownloadUrlError
				},
				DownloadFromReplicatedFunc: func(url, reqID, auth string) (*http.Response, error) {
					if tt.downloadReplicatedErr != nil {
						return nil, tt.downloadReplicatedErr
					}
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewReader([]byte("ok"))),
						Header:     http.Header{},
					}, nil
				},
			}

			s := &strategy.PlatformServiceStrategy{
				LicenseClient:     mockLicense,
				Replicated:        mockReplicated,
				LicenseServiceUrl: "http://licenseservice",
				Log:               log.NewEntry(log.New()),
			}
			s.SetLocals(map[string]interface{}{"requestId": "req123"})

			_, _, _, msg, code, err := s.DownloadChefPlatform(&omnitruck.RequestParams{LicenseId: "lic123"})
			assert.Error(t, err)
			assert.Equal(t, tt.expectedMsg, msg)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}
