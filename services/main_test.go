package services

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/dboperations"
	_ "github.com/chef/omnitruck-service/docs"
	"github.com/chef/omnitruck-service/logger"
	"github.com/chef/omnitruck-service/models"
	"github.com/chef/omnitruck-service/utils/template"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func getLogger() logger.ILogger {
	log, _ := logger.NewStandardLogger()
	return log
}

func TestRelatedProductsHandler(t *testing.T) {

	tests := []struct {
		name             string
		requestPath      string
		expectedStatus   int
		expectedResponse string
		relatedProducts  models.RelatedProducts
		err              error
	}{
		{
			name:             "Valid SKU with related products",
			requestPath:      "/relatedProducts?bom=Chef%20Desktop%20Management",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"relatedProducts": {"inspec": "Chef InSpec"}}`,
			relatedProducts:  models.RelatedProducts{Products: map[string]string{"inspec": "Chef InSpec"}},
			err:              nil,
		},
		{
			name:             "Empty SKU",
			requestPath:      "/relatedProducts?bom=",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"BOM (bom) params cannot be empty", "status_text":"Bad Request"}`,
			relatedProducts:  models.RelatedProducts{},
			err:              errors.New("No Related products found for SKU "),
		},
		{
			name:             "Db error while fetching related products",
			requestPath:      "/relatedProducts?bom=Chef%20123",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			relatedProducts:  models.RelatedProducts{},
			err:              errors.New("Db connection error"),
		},
		{
			name:             "No related products",
			requestPath:      "/relatedProducts?bom=Chef%20123",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Product information not found. Please check the input parameters.", "status_text":"Bad Request"}`,
			relatedProducts:  models.RelatedProducts{},
			err:              nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetRelatedProductsfunc = func(partitionValue string) (*models.RelatedProducts, error) {
				return &test.relatedProducts, test.err
			}

			server := &ApiService{
				App:             app,
				DatabaseService: mockDbService,
				Log:             getLogger(),
			}
			server.buildRouter()
			req := httptest.NewRequest(http.MethodGet, test.requestPath, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedStatus, resp.StatusCode)

			if test.expectedResponse != "" {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, test.expectedResponse, string(bodyBytes))
			}
		})
	}
}

func TestLatestVersionsHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestPath      string
		serverMode       ApiType
		expectedStatus   int
		expectedResponse string
		version          string
		version_err      error
		versions         []string
		versions_err     error
	}{
		{
			name:             "success for opensource",
			requestPath:      "/stable/habitat/versions/latest",
			serverMode:       Opensource,
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `"0.9.3"`,
			versions:         []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0", "1.0.0"},
			versions_err:     nil,
		},
		{
			name:             "success for trail",
			requestPath:      "/stable/habitat/versions/latest",
			serverMode:       Trial,
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `"1.0.0"`,
			version:          "1.0.0",
			version_err:      nil,
			versions:         []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0", "1.0.0"},
			versions_err:     nil,
		},
		{
			name:             "failure validation",
			requestPath:      "/stale/automate/versions/latest",
			serverMode:       Trial,
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Channel can only be stable or current", "status_text":"Bad Request"}`,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return test.versions, test.versions_err
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return test.version, test.version_err
			}

			server := &ApiService{
				App:             app,
				DatabaseService: mockDbService,
				Log:             getLogger(),
				Mode:            test.serverMode,
			}
			server.buildRouter()
			req := httptest.NewRequest(http.MethodGet, test.requestPath, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedStatus, resp.StatusCode)

			if test.expectedResponse != "" {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, test.expectedResponse, string(bodyBytes))
			}
		})
	}
}

func TestProductVersionsHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestPath      string
		serverMode       ApiType
		expectedStatus   int
		expectedResponse string
		versions         []string
		versions_err     error
	}{
		{
			name:             "success for opensource",
			requestPath:      "/stable/habitat/versions/all",
			serverMode:       Opensource,
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `["0.3.2", "0.7.11", "0.9.0", "0.9.3"]`,
			versions:         []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0", "1.0.0"},
			versions_err:     nil,
		},
		{
			name:             "success for trail",
			requestPath:      "/stable/habitat/versions/all",
			serverMode:       Trial,
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `["1.0.0"]`,
			versions:         []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0", "1.0.0"},
			versions_err:     nil,
		},
		{
			name:             "failure validation",
			requestPath:      "/stale/automate/versions/all",
			serverMode:       Trial,
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Channel can only be stable or current", "status_text":"Bad Request"}`,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return test.versions, test.versions_err
			}

			server := &ApiService{
				App:             app,
				DatabaseService: mockDbService,
				Log:             getLogger(),
				Mode:            test.serverMode,
			}
			server.buildRouter()
			req := httptest.NewRequest(http.MethodGet, test.requestPath, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedStatus, resp.StatusCode)

			if test.expectedResponse != "" {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, test.expectedResponse, string(bodyBytes))
			}
		})
	}
}

func TestProductMetadataHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestPath      string
		serverMode       ApiType
		expectedStatus   int
		expectedResponse string
		metadata         models.MetaData
		err              error
		version          string
		version_err      error
		versions         []string
		versions_err     error
	}{
		{
			name:             "automate success",
			serverMode:       Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=amd64&eol=false&v=latest",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"sha1": "","sha256": "1234","url": "http://example.com/stable/automate/download?eol=false&m=amd64&p=linux&v=latest","version": "latest"}`,
			metadata: models.MetaData{
				Architecture:     "amd64",
				FileName:         "",
				Platform:         "linux",
				Platform_Version: "",
				SHA1:             "",
				SHA256:           "1234",
			},
			err:          nil,
			version:      "latest",
			version_err:  nil,
			versions:     []string{},
			versions_err: nil,
		},
		{
			name:             "platform parameter missing",
			serverMode:       Trial,
			requestPath:      "/stable/automate/metadata?&m=amd64&eol=false&v=latest",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Platfrom (p) params cannot be empty", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{},
			versions_err:     nil,
		},
		{
			name:             "automate parameter incorrect",
			serverMode:       Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=x86&eol=false",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Product information not found. Please check the input parameters.", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              nil,
			versions:         []string{},
			versions_err:     nil,
		},
		{
			name:             "automate not latest version for trail server",
			serverMode:       Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=x86_64&v=1.2",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Version is not latest.", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              errors.New("ResourceNotFoundException: Requested resource not found"),
			version:          "latest",
			version_err:      nil,
			versions:         []string{},
			versions_err:     nil,
		},
		{
			name:             "automate db connection error",
			serverMode:       Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=x86_64&eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			err:              errors.New("ResourceNotFoundException: Requested resource not found"),
			versions:         []string{},
			versions_err:     nil,
		},
		{
			name:             "opensource check success",
			serverMode:       Opensource,
			requestPath:      "/stable/habitat/metadata?p=linux&m=x86_64&eol=false",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"sha1":"", "sha256":"abcd", "url":"http://example.com/stable/habitat/download?eol=false&m=x86_64&p=linux&v=0.9.3", "version":"0.9.3"}`,
			metadata: models.MetaData{
				Architecture:     "x86_64",
				FileName:         "",
				Platform:         "linux",
				Platform_Version: "",
				SHA1:             "",
				SHA256:           "abcd",
			},
			err:          nil,
			version:      "",
			version_err:  nil,
			versions:     []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0"},
			versions_err: nil,
		},
		{
			name:             "opensource check failure",
			serverMode:       Opensource,
			requestPath:      "/stable/habitat/metadata?p=linux&m=x86_64&eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			err:              nil,
			version:          "",
			version_err:      nil,
			versions:         []string{},
			versions_err:     errors.New("ResourceNotFoundException: Requested resource not found"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture string) (*models.MetaData, error) {
				return &test.metadata, test.err
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return test.version, test.version_err
			}
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return test.versions, test.versions_err
			}

			server := &ApiService{
				App:             app,
				DatabaseService: mockDbService,
				Log:             getLogger(),
				Mode:            test.serverMode,
			}
			server.buildRouter()
			req := httptest.NewRequest(http.MethodGet, test.requestPath, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedStatus, resp.StatusCode)

			if test.expectedResponse != "" {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, test.expectedResponse, string(bodyBytes))
			}
		})
	}
}

func TestProductPackagesHandler(t *testing.T) {
	tests := []struct {
		name             string
		serverMode       ApiType
		requestPath      string
		expectedStatus   int
		expectedResponse string
		details          models.ProductDetails
		err              error
		version          string
		version_err      error
		versions         []string
		versions_err     error
	}{
		{
			name:             "success",
			serverMode:       Trial,
			requestPath:      "/stable/automate/packages?eol=false",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"linux": {"pv": {"amd64": {"sha1": "","sha256": "abcd","url": "http://example.com/stable/automate/download?eol=false&m=amd64&p=linux&v=latest","version": "latest"}}}}`,
			details: models.ProductDetails{
				Product: "automate",
				Version: "latest",
				MetaData: []models.MetaData{
					{
						Architecture:     "amd64",
						FileName:         "",
						Platform:         "linux",
						Platform_Version: "",
						SHA1:             "",
						SHA256:           "abcd",
					},
				},
			},
			err:          nil,
			version:      "latest",
			version_err:  nil,
			versions:     []string{},
			versions_err: nil,
		},
		{
			name:             "version is not latest",
			serverMode:       Trial,
			requestPath:      "/stable/automate/packages?eol=false&v=1",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Version is not latest.", "status_text":"Bad Request"}`,
			details:          models.ProductDetails{},
			err:              nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{},
			versions_err:     nil,
		},
		{
			name:             "db connection  error",
			serverMode:       Trial,
			requestPath:      "/stable/automate/packages?eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			details:          models.ProductDetails{},
			err:              nil,
			version:          "",
			version_err:      errors.New("ResourceNotFoundException: Requested resource not found"),
			versions:         []string{},
			versions_err:     nil,
		},
		{
			name:             "opensource check success",
			serverMode:       Opensource,
			requestPath:      "/stable/habitat/packages?eol=false",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"linux": {"pv": {"x86_64": {"sha1":"", "sha256":"abcd", "url":"http://example.com/stable/habitat/download?eol=false&m=x86_64&p=linux&v=0.9.3", "version":"0.9.3"}}}}`,
			details: models.ProductDetails{
				Product: "habitat",
				Version: "0.9.3",
				MetaData: []models.MetaData{{
					Architecture:     "x86_64",
					FileName:         "",
					Platform:         "linux",
					Platform_Version: "",
					SHA1:             "",
					SHA256:           "abcd",
				}},
			},
			err:          nil,
			version:      "",
			version_err:  nil,
			versions:     []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0"},
			versions_err: nil,
		},
		{
			name:             "opensource check failure",
			serverMode:       Opensource,
			requestPath:      "/stable/habitat/packages?eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			details:          models.ProductDetails{},
			err:              nil,
			version:          "",
			version_err:      nil,
			versions:         []string{},
			versions_err:     errors.New("ResourceNotFoundException: Requested resource not found"),
		},
		{
			name:             "empty metadate info",
			serverMode:       Trial,
			requestPath:      "/stable/habitat/packages?eol=false",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Product information not found. Please check the input parameters.", "status_text":"Bad Request"}`,
			details: models.ProductDetails{
				Product:  "habitat",
				Version:  "1.6.826",
				MetaData: []models.MetaData{},
			},
			err:          nil,
			version:      "1.6.826",
			version_err:  nil,
			versions:     []string{},
			versions_err: nil,
		},
		{
			name:             "opensource check success",
			serverMode:       Opensource,
			requestPath:      "/stable/habitat/packages?eol=false&v=0.3.2",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"linux": {"pv": {"x86_64": {"sha1":"", "sha256":"abcd", "url":"http://example.com/stable/habitat/download?eol=false&m=x86_64&p=linux&v=0.3.2", "version":"0.3.2"}}}}`,
			details: models.ProductDetails{
				Product: "habitat",
				Version: "0.3.2",
				MetaData: []models.MetaData{{
					Architecture:     "x86_64",
					FileName:         "",
					Platform:         "linux",
					Platform_Version: "",
					SHA1:             "",
					SHA256:           "abcd",
				}},
			},
			err:          nil,
			version:      "",
			version_err:  nil,
			versions:     []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0"},
			versions_err: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetPackagesfunc = func(partitionValue, sortValue string) (*models.ProductDetails, error) {
				return &test.details, test.err
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return test.version, test.version_err
			}
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return test.versions, test.versions_err
			}

			server := &ApiService{
				App:             app,
				DatabaseService: mockDbService,
				Log:             getLogger(),
				Mode:            test.serverMode,
			}
			server.buildRouter()
			req := httptest.NewRequest(http.MethodGet, test.requestPath, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedStatus, resp.StatusCode)

			if test.expectedResponse != "" {
				bodyBytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, test.expectedResponse, string(bodyBytes))
			}
		})
	}
}

func TestFileNameHandler(t *testing.T) {

	tests := []struct {
		name             string
		serverMode       ApiType
		requestPath      string
		expectedStatus   int
		expectedResponse string
		metadata         models.MetaData
		version          string
		metadata_err     error
		version_err      error
		versions         []string
		versions_err     error
	}{
		{
			name:             "AUTOMATE_SUCCESS",
			serverMode:       Trial,
			requestPath:      "/current/automate/fileName?p=linux&pv=16.04&m=x86_64&v=latest",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"fileName":"automate_4.7.52-1_amd64.deb"}`,
			metadata:         models.MetaData{FileName: "automate_4.7.52-1_amd64.deb"},
			metadata_err:     nil,
			version_err:      nil,
			versions:         []string{},
			versions_err:     nil,
		},
		{
			name:             "HABITAT_SUCCESS",
			serverMode:       Opensource,
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=latest",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"fileName":"hab-x86_64-linux.tar.gz"}`,
			metadata:         models.MetaData{FileName: "hab-x86_64-linux.tar.gz"},
			metadata_err:     nil,
			version_err:      nil,
			version:          "",
			versions:         []string{"0.78.0"},
			versions_err:     nil,
		},
		{
			name:             "AUTOMATE_FAIL",
			requestPath:      "/current/automate/fileName?p=linux&pv=16.04&m=x86_64&v=latest",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			version_err:      errors.New("Unable to get latest version of automate"),
		},
		{
			name:             "failure db connection while fetching latest version",
			serverMode:       Trial,
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=1.6.652",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			metadata_err:     nil,
			version:          "",
			version_err:      errors.New("Unable to get latest version of habitat"),
		},
		{
			name:             "failure channel is not stable/current",
			serverMode:       Trial,
			requestPath:      "/curret/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=1.6.652",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Channel can only be stable or current", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			metadata_err:     errors.New("Error while fetching file name"),
			version_err:      nil,
			version:          "1.6.652",
		},
		{
			name:             "failure db connection while fetching filename",
			serverMode:       Trial,
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=1.6.652",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			metadata_err:     errors.New("Error while fetching file name"),
			version_err:      nil,
			version:          "1.6.652",
		},
		{
			name:             "haitat opensource verion not supported ",
			serverMode:       Opensource,
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=0.79.0",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Version 0.79.0 not support on this persona.", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			metadata_err:     errors.New("Error while fetching file name"),
			version_err:      nil,
			version:          "",
			versions:         []string{"0.79.0", "0.78.0"},
			versions_err:     nil,
		},
		{
			name:             "haitat opensource version fetching error",
			serverMode:       Opensource,
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=0.79.0",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			metadata_err:     errors.New("Error while fetching file name"),
			version_err:      nil,
			version:          "",
			versions:         []string{},
			versions_err:     errors.New("Unable to get all versions of habitat"),
		},
	}

	for _, test := range tests {
		timeout := 1 * time.Minute
		t.Run(test.name, func(t *testing.T) {
			done := make(chan struct{})
			go func() {
				defer close(done)
				app := fiber.New()
				mockDbService := new(dboperations.MockIDbOperations)

				mockDbService.GetMetaDatafunc = func(partitionValue string, sortValue string, platform string, platformVersion string, architecture string) (*models.MetaData, error) {
					return &test.metadata, test.metadata_err
				}
				mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
					return test.version, test.version_err
				}
				mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
					return test.versions, test.versions_err
				}

				server := &ApiService{
					App:             app,
					DatabaseService: mockDbService,
					Log:             getLogger(),
					Mode:            test.serverMode,
				}
				server.buildRouter()
				req := httptest.NewRequest(http.MethodGet, test.requestPath, nil)
				resp, err := app.Test(req)
				assert.NoError(t, err)

				assert.Equal(t, test.expectedStatus, resp.StatusCode)

				if test.expectedResponse != "" {
					bodyBytes, err := io.ReadAll(resp.Body)
					assert.NoError(t, err)
					assert.JSONEq(t, test.expectedResponse, string(bodyBytes))
				}
			}()
			select {
			case <-done:
				// Test completed within the timeout, nothing to do here
			case <-time.After(timeout):
				t.Errorf("Test took too long to complete (timeout: %s)", timeout)
			}
		})
	}
}

func TestDownloadLinuxScriptHandler(t *testing.T) {
	tests := []struct {
		name             string
		serverMode       ApiType
		mockTemplate    func(baseUrl string, params *omnitruck.RequestParams, filepath string) (string, error)
		requestPath      string
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:       "success",
			serverMode: 1,
			mockTemplate: func(baseUrl string, params *omnitruck.RequestParams, filepath string) (string, error) {
				return "", nil
			},
			requestPath:      `/install.sh`,
			expectedStatus:   200,
			expectedResponse: ``,
		},
		{
			name:       "error while parsing the file response",
			serverMode: 0,
			mockTemplate: func(baseUrl string, params *omnitruck.RequestParams, filepath string) (string, error) {
				return "", errors.New("filepath not found")
			},
			requestPath:      `/install.sh`,
			expectedStatus:   500,
			expectedResponse: ``,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			mockTemplate := new(template.MockTemplateRennder)
			mockTemplate.GetScriptfunc = test.mockTemplate
			server := &ApiService{
				App:              app,
				TemplateRenderer: mockTemplate,
				Log:              getLogger(),
				Mode:             test.serverMode,
			}
			server.buildRouter()
			req := httptest.NewRequest(http.MethodGet, test.requestPath, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedStatus, resp.StatusCode)
		})
	}
}

func TestDownloadWindowsScriptHandler(t *testing.T) {
	tests := []struct {
		name             string
		serverMode       ApiType
		mockTemplate    func(baseUrl string, params *omnitruck.RequestParams, filepath string) (string, error)
		requestPath      string
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:       "success",
			serverMode: 1,
			mockTemplate: func(baseUrl string, params *omnitruck.RequestParams, filepath string) (string, error) {
				return "", nil
			},
			requestPath:      `/install.ps1`,
			expectedStatus:   200,
			expectedResponse: ``,
		},
		{
			name:       "error while parsing the file response",
			serverMode: 0,
			mockTemplate: func(baseUrl string, params *omnitruck.RequestParams, filepath string) (string, error) {
				return "", errors.New("filepath not found")
			},
			requestPath:      `/install.ps1`,
			expectedStatus:   500,
			expectedResponse: ``,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			mockTemplate := new(template.MockTemplateRennder)
			mockTemplate.GetScriptfunc = test.mockTemplate
			server := &ApiService{
				App:              app,
				TemplateRenderer: mockTemplate,
				Log:              getLogger(),
				Mode:             test.serverMode,
			}
			server.buildRouter()
			req := httptest.NewRequest(http.MethodGet, test.requestPath, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedStatus, resp.StatusCode)
		})
	}
}
