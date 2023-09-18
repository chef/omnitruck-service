package services

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	_ "github.com/chef/omnitruck-service/docs"

	"github.com/chef/omnitruck-service/dboperations"
	"github.com/chef/omnitruck-service/models"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

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
			requestPath:      "/relatedProducts?sku=Chef%20Desktop%20Management",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"relatedProducts": {"inspec": "Chef InSpec"}}`,
			relatedProducts:  models.RelatedProducts{Products: map[string]string{"inspec": "Chef InSpec"}},
			err:              nil,
		},
		{
			name:             "Invalid SKU",
			requestPath:      "/relatedProducts?sku=invalid-sku",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Unable to retrieve related products for invalid-sku", "status_text":"Internal Server Error"}`,
			relatedProducts:  models.RelatedProducts{},
			err:              errors.New("No Related products found for SKU "),
		},
		{
			name:             "No related products",
			requestPath:      "/relatedProducts?sku=Chef%20123",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"No related products found for SKU", "status_text":"Bad Request"}`,
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
				Log:             logrus.NewEntry(logrus.New()),
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

func TestApiService_productMetadataHandler(t *testing.T) {
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
			name:             "automate db connection error",
			serverMode:       Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=x86_64&eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product.", "status_text":"Internal Server Error"}`,
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
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product.", "status_text":"Internal Server Error"}`,
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
				Log:             logrus.NewEntry(logrus.New()),
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

func TestApiService_productPackagesHandler(t *testing.T) {
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
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product.", "status_text":"Internal Server Error"}`,
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
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product.", "status_text":"Internal Server Error"}`,
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
				Log:             logrus.NewEntry(logrus.New()),
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
		requestPath      string
		expectedStatus   int
		expectedResponse string
		metadata         models.MetaData
		version          string
		err              error
	}{
		{
			name:             "AUTOMATE_SUCCESS",
			requestPath:      "/current/automate/fileName?p=linux&pv=16.04&m=x86_64&v=latest",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"fileName":"automate_4.7.52-1_amd64.deb"}`,
			metadata:         models.MetaData{FileName: "automate_4.7.52-1_amd64.deb"},
			err:              nil,
		},
		{
			name:             "HABITAT_SUCCESS",
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=latest",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"fileName":"hab-x86_64-linux.tar.gz"}`,
			metadata:         models.MetaData{FileName: "hab-x86_64-linux.tar.gz"},
			err:              nil,
			version:          "1.6.652",
		},
		{
			name:             "AUTOMATE_FAIL",
			requestPath:      "/current/automate/fileName?p=linux&pv=16.04&m=x86_64&v=latest",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Unable to get latest version of automate", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			err:              errors.New("Unable to get latest version of automate"),
		},
		{
			name:             "HABITAT_FAIL",
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=latest",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Unable to get latest version of habitat", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			err:              errors.New("Unable to get latest version of habitat"),
		},
		{
			name:             "HABITAT_FAIL_1",
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=1.6.652",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while getting the requested information.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			err:              errors.New("Error while fetching file name"),
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
					return &test.metadata, test.err
				}
				mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
					return test.version, test.err
				}

				server := &ApiService{
					App:             app,
					DatabaseService: mockDbService,
					Log:             logrus.NewEntry(logrus.New()),
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
