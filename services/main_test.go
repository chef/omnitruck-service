package services

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chef/omnitruck-service/dboperations"
	_ "github.com/chef/omnitruck-service/docs"
	"github.com/chef/omnitruck-service/models"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestApiService_productMetadataHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestPath      string
		expectedStatus   int
		expectedResponse string
		metadata         models.MetaData
		err              error
		version          string
		version_err      error
	}{
		{
			name:             "automate success",
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
			err:         nil,
			version:     "latest",
			version_err: nil,
		},
		{
			name:             "automate parameter incorrect",
			requestPath:      "/stable/automate/metadata?p=linux&m=x86&eol=false",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Product information not found. Please check the input parameters.", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              nil,
		},
		{
			name:             "automate db connection error",
			requestPath:      "/stable/automate/metadata?p=linux&m=x86_64&eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			err:              errors.New("ResourceNotFoundException: Requested resource not found"),
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

func TestApiService_productPackagesHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestPath      string
		expectedStatus   int
		expectedResponse string
		details          models.ProductDetails
		err              error
		version          string
		version_err      error
	}{
		{
			name:             "success",
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
			err:         nil,
			version:     "latest",
			version_err: nil,
		},
		{
			name:             "db connection  error",
			requestPath:      "/stable/automate/packages?eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product.", "status_text":"Internal Server Error"}`,
			details:          models.ProductDetails{},
			err:              nil,
			version:          "",
			version_err:      errors.New("ResourceNotFoundException: Requested resource not found"),
		},
		{
			name:             "success",
			requestPath:      "/stable/habitat/packages?eol=false",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Product information not found. Please check the input parameters.", "status_text":"Bad Request"}`,
			details: models.ProductDetails{
				Product:  "habitat",
				Version:  "1.6.826",
				MetaData: []models.MetaData{},
			},
			err:         nil,
			version:     "1.6.826",
			version_err: nil,
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
