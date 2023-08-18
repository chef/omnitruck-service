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
		servermode       ApiType
		requestPath      string
		expectedStatus   int
		expectedResponse string
		metadata         models.MetaData
		err              error
	}{
		{
			name:             "automate success",
			servermode:       Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=x86_64&eol=false",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"sha1": "","sha256": "1234","url": "http://example.com/stable/automate/download?eol=false&m=x86_64&p=linux&v=latest","version": "latest"}`,
			metadata: models.MetaData{
				Architecture:     "x86_64",
				FileName:         "",
				Platform:         "linux",
				Platform_Version: "",
				SHA1:             "",
				SHA256:           "1234",
			},
			err: nil,
		},
		{
			name:             "automate parameter incorrect",
			servermode:       Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=x86&eol=false",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Product information not found. Please check the input parameters", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              nil,
		},
		{
			name:             "automate db connection error",
			servermode:       Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=x86_64&eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			err:              errors.New("ResourceNotFoundException: Requested resource not found"),
		},
		{
			name:             "server mode opensource",
			servermode:       Opensource,
			requestPath:      "/stable/automate/metadata?p=linux&m=x86_64&eol=false",
			expectedStatus:   fiber.StatusForbidden,
			expectedResponse: `{"code":403, "message":"Product not supported.", "status_text":"Forbidden"}`,
			metadata:         models.MetaData{},
			err:              nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture string) (*models.MetaData, error) {
				return &test.metadata, test.err
			}

			server := &ApiService{
				App:             app,
				DatabaseService: mockDbService,
				Log:             logrus.NewEntry(logrus.New()),
				Mode:            test.servermode,
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
