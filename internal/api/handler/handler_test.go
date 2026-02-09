package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"time"

	"testing"

	"github.com/chef/omnitruck-service/clients"

	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/models"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/dboperations"
	_ "github.com/chef/omnitruck-service/docs"

	"github.com/chef/omnitruck-service/utils/template"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testInjector(
	db dboperations.IDbOperations,
	mode constants.ApiType,
	renderer template.TemplateRenderer,
) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		reqInjector := do.New()

		do.ProvideNamedValue[dboperations.IDbOperations](reqInjector, "dbService", db)
		do.ProvideNamedValue[template.TemplateRenderer](reqInjector, "templateRenderer", renderer)
		do.ProvideNamedValue[replicated.IReplicated](reqInjector, "replicated", &replicated.MockReplicated{})
		do.ProvideNamedValue[clients.ILicense](reqInjector, "licenseClient", &clients.MockLicense{})
		do.ProvideNamedValue[constants.ApiType](reqInjector, "mode", mode)
		do.ProvideNamedValue[config.ServiceConfig](reqInjector, "config", config.ServiceConfig{
			LicenseServiceUrl: "http://licenseservice",
			OmnitruckUrl:      "https://omnitruck.chef.io",
			SupportInfra19:    true,
		})

		do.ProvideNamedValue[omnitruck.IRequestValidator](reqInjector, "validator", &omnitruck.MockRequestValidator{
			ParamsFunc: func(params *omnitruck.RequestParams, ctx omnitruck.Context) []*omnitruck.ValidationError {
				// if channel is blank and product is blank, treat as download script, allow
				if params.Channel == "" && params.Product == "" {
					return nil
				}
				// for product routes, enforce channel check
				if params.Channel != "stable" && params.Channel != "current" {
					return []*omnitruck.ValidationError{
						{
							Msg:  "Channel can only be stable or current",
							Code: fiber.StatusBadRequest,
						},
					}
				}
				return nil
			},
			ErrorMessagesFunc: func(errors []*omnitruck.ValidationError) (string, int) {
				if len(errors) == 0 {
					return "", 0
				}
				return errors[0].Msg, errors[0].Code
			},
		})

		c.Locals("reqinjector", reqInjector)
		return c.Next()
	}
}

func TestProductsHandler(t *testing.T) {
	t.Parallel()

	// Mock DB Service
	mockDb := new(dboperations.MockIDbOperations)
	mockDb.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
		return []string{"0.1.1"}, nil
	}
	mockDb.GetVersionLatestfunc = func(partitionValue string) (string, error) {
		return "0.1.1", nil
	}

	log := logrus.NewEntry(logrus.New())

	tests := []struct {
		name             string
		inject           bool
		customMiddleware func(*fiber.Ctx) error
		expectedStatus   int
		expectedContains []string
		eolParam         string
	}{
		{
			name:             "missing injector returns 500",
			inject:           false,
			expectedStatus:   http.StatusInternalServerError,
			eolParam:         "false",
			expectedContains: []string{"{\"code\":500,\"status_text\":\"Internal Server Error\",\"message\":\"Not able to process the request.\"}"},
		},
		{
			name:   "broken injector returns 500 on NewDownloadService failure",
			inject: true,
			customMiddleware: func(c *fiber.Ctx) error {
				reqInjector := do.New()
				c.Locals("reqinjector", reqInjector)
				return c.Next()
			},
			eolParam:         "false",
			expectedStatus:   http.StatusInternalServerError,
			expectedContains: []string{"Failed to create download service"},
		},
		{
			name:             "products returns success",
			inject:           true,
			customMiddleware: testInjector(mockDb, constants.Commercial, &template.MockTemplateRenderer{}),
			eolParam:         "false",
			expectedStatus:   http.StatusOK,
			expectedContains: []string{"chef"},
		},
		{
			name:             "constants.Opensource mode filters products",
			inject:           true,
			customMiddleware: testInjector(mockDb, constants.Opensource, &template.MockTemplateRenderer{}),
			eolParam:         "false",
			expectedStatus:   fiber.StatusOK,
			expectedContains: []string{"habitat"},
		},
		{
			name:             "constants.Trial mode adds enterprise product",
			inject:           true,
			customMiddleware: testInjector(mockDb, constants.Trial, &template.MockTemplateRenderer{}),
			eolParam:         "false",
			expectedStatus:   fiber.StatusOK,
			expectedContains: []string{"Chef Infra Client Enterprise"},
		},
		{
			name:             "constants.Commercial mode adds products",
			inject:           true,
			customMiddleware: testInjector(mockDb, constants.Commercial, &template.MockTemplateRenderer{}),
			eolParam:         "false",
			expectedStatus:   fiber.StatusOK,
			expectedContains: []string{"chef-360", "chef-ice", "migrate-ice"},
		},
		{
			name:             "constants.Trial mode with eol true includes Chef Infra Client Enterprise and automate-1",
			inject:           true,
			customMiddleware: testInjector(mockDb, constants.Trial, &template.MockTemplateRenderer{}),
			eolParam:         "true",
			expectedStatus:   fiber.StatusOK,
			expectedContains: []string{"Chef Infra Client Enterprise", "automate-1"},
		},
		{
			name:             "constants.Commercial mode with eol true includes automate-1",
			inject:           true,
			customMiddleware: testInjector(mockDb, constants.Commercial, &template.MockTemplateRenderer{}),
			eolParam:         "true",
			expectedStatus:   fiber.StatusOK,
			expectedContains: []string{"chef-360", "chef-ice", "migrate-ice", "automate-1"},
		},
		{
			name:             "constants.Trial mode returns formatted products",
			inject:           true,
			customMiddleware: testInjector(mockDb, constants.Trial, &template.MockTemplateRenderer{}),
			eolParam:         "false",
			expectedStatus:   fiber.StatusOK,
			expectedContains: []string{"automate:Chef Automate", "chef:Chef Infra Client (Legacy)", "chef-server:Chef Infra Server", "chef-workstation:Chef Workstation", "habitat:Chef Habitat", "inspec:InSpec", "chef-ice:Chef Infra Client Enterprise", "migrate-ice:Chef Infra Client Legacy Migration"},
		},
		{
			name:             "constants.Commercial mode returns full product list",
			inject:           true,
			customMiddleware: testInjector(mockDb, constants.Commercial, &template.MockTemplateRenderer{}),
			eolParam:         "false",
			expectedStatus:   fiber.StatusOK,
			expectedContains: []string{"automate", "chef", "chef-backend", "chef-server", "chef-workstation", "habitat", "inspec", "manage", "supermarket", "chef-360", "chef-ice", "migrate-ice"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(func(c *fiber.Ctx) error {
				c.Locals("base_url", "http://example.com")
				return c.Next()
			})

			if tt.inject {
				app.Use(tt.customMiddleware)
			}

			handler := NewDownloadsHandler(log)
			app.Get("/products", handler.ProductsHandler)

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/products?eol="+tt.eolParam, nil), 100*1000) // 100 seconds timeout
			require.NoError(t, err)
			defer resp.Body.Close()

			bodyBytes, _ := io.ReadAll(resp.Body)
			body := string(bodyBytes)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, string(body))
			for _, expected := range tt.expectedContains {
				assert.Contains(t, body, expected)
			}
		})
	}
}

func TestLatestVersionsHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestPath      string
		serverMode       constants.ApiType
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
			serverMode:       constants.Opensource,
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `"0.9.3"`,
			versions:         []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0", "1.0.0"},
		},
		{
			name:             "chef-360 success",
			requestPath:      "/stable/chef-360/versions/latest",
			serverMode:       constants.Commercial,
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `"latest"`,
			versions:         []string{"latest"},
		},
		{
			name:             "failure for chef-360 when opensource is the server type",
			requestPath:      "/stable/chef-360/versions/latest",
			serverMode:       constants.Opensource,
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400,"status_text":"Bad Request","message":"chef-360 not available for the trial and opensource"}`,
			versions:         []string{},
		},
		{
			name:             "failure for chef-360 when trial is the server type",
			requestPath:      "/stable/chef-360/versions/latest",
			serverMode:       constants.Trial,
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400,"status_text":"Bad Request","message":"chef-360 not available for the trial and opensource"}`,
			versions:         []string{},
		},
		{
			name:             "success for trial",
			requestPath:      "/stable/habitat/versions/latest",
			serverMode:       constants.Trial,
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `"1.0.0"`,
			version:          "1.0.0",
			versions:         []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0", "1.0.0"},
		},
		{
			name:             "failure validation",
			requestPath:      "/stale/automate/versions/latest",
			serverMode:       constants.Trial,
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400,"status_text":"Bad Request","message":"Channel can only be stable or current"}`,
			versions:         []string{"latest"},
		},
	}

	for _, test := range tests {
		test := test // pin variable
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("base_url", "http://example.com")
				return c.Next()
			})

			// mock db
			mockDbService := &dboperations.MockIDbOperations{
				GetVersionAllfunc: func(partitionValue string) ([]string, error) {
					return test.versions, test.versions_err
				},
				GetVersionLatestfunc: func(partitionValue string) (string, error) {
					return test.version, test.version_err
				},
				SetDbInfofunc: func(tableName string, dbModel reflect.Type) {},
			}

			log := logrus.NewEntry(logrus.New())
			h := NewDownloadsHandler(log)

			app.Use(testInjector(
				mockDbService,
				test.serverMode,
				&template.MockTemplateRenderer{},
			))

			app.Get("/:channel/:product/versions/latest", func(c *fiber.Ctx) error {
				return h.LatestVersionHandler(c)
			})

			req := httptest.NewRequest(http.MethodGet, "http://example.com"+test.requestPath, nil)
			resp, err := app.Test(req, 10_000)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedStatus, resp.StatusCode)

			if test.expectedResponse != "" {
				bodyBytes, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, test.expectedResponse, string(bodyBytes))
			}
		})
	}
}

func TestProductVersionsHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestPath      string
		serverMode       constants.ApiType
		expectedStatus   int
		expectedResponse string
		versions         []string
		versions_err     error
	}{
		{
			name:             "success for opensource",
			requestPath:      "/stable/habitat/versions/all",
			serverMode:       constants.Opensource,
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `["0.3.2", "0.7.11", "0.9.0", "0.9.3"]`,
			versions:         []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0", "1.0.0"},
			versions_err:     nil,
		},
		{
			name:             "success for chef-360",
			requestPath:      "/stable/chef-360/versions/all",
			serverMode:       constants.Commercial,
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `["latest"]`,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "failure for chef-360 when server is not commercial",
			requestPath:      "/stable/chef-360/versions/all",
			serverMode:       constants.Trial,
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"chef-360 not available for the trial and opensource", "status_text":"Bad Request"}`,
			versions:         []string{},
			versions_err:     nil,
		},
		{
			name:             "success for trial",
			requestPath:      "/stable/habitat/versions/all",
			serverMode:       constants.Trial,
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `["1.0.0"]`,
			versions:         []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0", "1.0.0"},
			versions_err:     nil,
		},
		{
			name:             "failure validation",
			requestPath:      "/stale/automate/versions/all",
			serverMode:       constants.Trial,
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Channel can only be stable or current", "status_text":"Bad Request"}`,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("base_url", "http://example.com")
				return c.Next()
			})
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return test.versions, test.versions_err
			}
			mockDbService.SetDbInfofunc = func(tableName string, dbModel reflect.Type) {
			}

			log := logrus.NewEntry(logrus.New())
			handler := NewDownloadsHandler(log)
			app.Use(testInjector(mockDbService, test.serverMode, &template.MockTemplateRenderer{}))
			app.Get("/:channel/:product/versions/all", func(c *fiber.Ctx) error {
				return handler.ProductVersionsHandler(c)
			})

			req := httptest.NewRequest(http.MethodGet, "http://example.com"+test.requestPath, nil)
			resp, err := app.Test(req, 100*1000) // 100 seconds timeout

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
		serverMode       constants.ApiType
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
			serverMode:       constants.Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=amd64&eol=false&v=latest",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"sha1":"", "sha256":"1234", "url":"http://example.com/stable/automate/download?eol=false&m=amd64&p=linux&v=latest", "version":"latest"}`,
			metadata: models.MetaData{
				Architecture:    "amd64",
				FileName:        "",
				Platform:        "linux",
				PlatformVersion: "",
				SHA1:            "",
				SHA256:          "1234",
			},
			err:          nil,
			version:      "latest",
			version_err:  nil,
			versions:     []string{"latest"},
			versions_err: nil,
		},
		{
			name:             "chef-360 success",
			serverMode:       constants.Commercial,
			requestPath:      "/stable/chef-360/metadata?p=ubuntu&pv=20.04&m=x86_64&v=latest&eol=false&license_id=viv2c0a2-111f-2caf-1fa2-1211fe1212d1",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"sha1":"", "sha256":"", "url":"http://example.com/stable/chef-360/download?eol=false&license_id=viv2c0a2-111f-2caf-1fa2-1211fe1212d1&m=x86_64&p=ubuntu&pv=20.04&v=latest", "version":"latest"}`,
			metadata: models.MetaData{
				Architecture:    "amd64",
				FileName:        "",
				Platform:        "linux",
				PlatformVersion: "",
				SHA1:            "",
				SHA256:          "",
			},
			err:          nil,
			version:      "",
			version_err:  nil,
			versions:     []string{"latest"},
			versions_err: nil,
		},
		{
			name:             "chef-360 fails on trial server",
			serverMode:       constants.Trial,
			requestPath:      "/stable/chef-360/metadata?p=ubuntu&pv=20.04&m=x86_64&v=latest&eol=false&license_id=viv2c0a2-111f-2caf-1fa2-1211fe1212d1",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"chef-360 not available for the trial and opensource", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "platform parameter missing",
			serverMode:       constants.Trial,
			requestPath:      "/stable/automate/metadata?&m=amd64&eol=false&v=latest",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Platfrom (p) params cannot be empty", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "automate parameter incorrect",
			serverMode:       constants.Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=x86&eol=false",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Product information not found. Please check the input parameters.", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              nil,
			versions:         []string{},
			versions_err:     nil,
		},
		{
			name:             "automate not latest version for trial server",
			serverMode:       constants.Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=x86_64&v=1.2",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"the requested version is not supported on the selected persona or channel", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              errors.New("ResourceNotFoundException: Requested resource not found"),
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "automate db connection error",
			serverMode:       constants.Trial,
			requestPath:      "/stable/automate/metadata?p=linux&m=x86_64&eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			err:              errors.New("ResourceNotFoundException: Requested resource not found"),
			version:          "latest",
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "opensource check success",
			serverMode:       constants.Opensource,
			requestPath:      "/stable/habitat/metadata?p=linux&m=x86_64&eol=false",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"sha1":"", "sha256":"abcd", "url":"http://example.com/stable/habitat/download?eol=false&m=x86_64&p=linux&v=0.9.3", "version":"0.9.3"}`,
			metadata: models.MetaData{
				Architecture:    "x86_64",
				FileName:        "",
				Platform:        "linux",
				PlatformVersion: "",
				SHA1:            "",
				SHA256:          "abcd",
			},
			err:          nil,
			version:      "",
			version_err:  nil,
			versions:     []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0"},
			versions_err: nil,
		},
		{
			name:             "opensource check failure",
			serverMode:       constants.Opensource,
			requestPath:      "/stable/habitat/metadata?p=linux&m=x86_64&eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching product versions", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			err:              nil,
			version:          "",
			version_err:      nil,
			versions:         []string{},
			versions_err:     errors.New("ResourceNotFoundException: Requested resource not found"),
		},
		{
			name:             "package manager parameter missing",
			serverMode:       constants.Trial,
			requestPath:      "/stable/chef-ice/metadata?p=linux&m=amd64&eol=false&v=latest",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Package Manager (pm) params cannot be empty", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "chef-ice success",
			serverMode:       constants.Trial,
			requestPath:      "/stable/chef-ice/metadata?p=linux&m=amd64&pm=deb&eol=false&v=latest",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"sha1":"", "sha256":"abcd", "url":"http://example.com/stable/chef-ice/download?eol=false&m=amd64&p=linux&pm=deb&v=latest", "version":"latest"}`,
			metadata: models.MetaData{
				Architecture:   "amd64",
				Platform:       "linux",
				PackageManager: "deb",
				SHA1:           "",
				SHA256:         "abcd",
				FileName:       "",
			},
			err:          nil,
			version:      "latest",
			version_err:  nil,
			versions:     []string{"latest"},
			versions_err: nil,
		},
		{
			name:             "chef-ice failure for opensource server",
			serverMode:       constants.Opensource,
			requestPath:      "/stable/chef-ice/metadata?p=linux&m=amd64&pm=deb&eol=false&v=latest",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"No versions found for this product/mode", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "package manager parameter missing",
			serverMode:       constants.Trial,
			requestPath:      "/stable/migrate-ice/metadata?p=linux&m=amd64&eol=false&v=latest",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Package Manager (pm) params cannot be empty", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "migrate-ice success",
			serverMode:       constants.Trial,
			requestPath:      "/stable/migrate-ice/metadata?p=linux&m=amd64&pm=deb&eol=false&v=latest",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"sha1":"", "sha256":"abcd", "url":"http://example.com/stable/migrate-ice/download?eol=false&m=amd64&p=linux&pm=deb&v=latest", "version":"latest"}`,
			metadata: models.MetaData{
				Architecture:   "amd64",
				Platform:       "linux",
				PackageManager: "deb",
				SHA1:           "",
				SHA256:         "abcd",
				FileName:       "",
			},
			err:          nil,
			version:      "latest",
			version_err:  nil,
			versions:     []string{"latest"},
			versions_err: nil,
		},
		{
			name:             "migrate-ice failure for opensource server",
			serverMode:       constants.Opensource,
			requestPath:      "/stable/migrate-ice/metadata?p=linux&m=amd64&pm=deb&eol=false&v=latest",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"No versions found for this product/mode", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			err:              nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("base_url", "http://example.com")
				return c.Next()
			})
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture, packageManager string) (*models.MetaData, error) {
				return &test.metadata, test.err
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return test.version, test.version_err
			}
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return test.versions, test.versions_err
			}
			mockDbService.SetDbInfofunc = func(tableName string, dbModel reflect.Type) {
			}

			log := logrus.NewEntry(logrus.New())
			handler := NewDownloadsHandler(log)

			app.Use(testInjector(mockDbService, test.serverMode, &template.MockTemplateRenderer{}))
			app.Get("/:channel/:product/metadata", func(c *fiber.Ctx) error {
				return handler.ProductMetadataHandler(c)
			})

			req := httptest.NewRequest(http.MethodGet, "http://example.com"+test.requestPath, nil)
			resp, err := app.Test(req, 100*1000) // 100 seconds timeout

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
		serverMode       constants.ApiType
		requestPath      string
		expectedStatus   int
		expectedResponse string
		details          interface{}
		err              error
		version          string
		version_err      error
		versions         []string
		versions_err     error
	}{
		{
			name:             "success",
			serverMode:       constants.Trial,
			requestPath:      "/stable/automate/packages?eol=false",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"linux": {"pv": {"amd64": {"sha1": "","sha256": "abcd","url": "http://example.com/stable/automate/download?eol=false&m=amd64&p=linux&v=latest","version": "latest"}}}}`,
			details: &models.ProductDetails{
				Product: "automate",
				Version: "latest",
				MetaData: []models.MetaData{
					{
						Architecture:    "amd64",
						FileName:        "",
						Platform:        "linux",
						PlatformVersion: "",
						SHA1:            "",
						SHA256:          "abcd",
					},
				},
			},
			err:          nil,
			version:      "latest",
			version_err:  nil,
			versions:     []string{"latest"},
			versions_err: nil,
		},
		{
			name:             "version is not latest",
			serverMode:       constants.Trial,
			requestPath:      "/stable/automate/packages?eol=false&v=1",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"the requested version is not supported on the selected persona or channel", "status_text":"Bad Request"}`,
			details:          &models.ProductDetails{},
			err:              nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latestt"},
			versions_err:     nil,
		},
		{
			name:             "chef-360 failure for trial server",
			serverMode:       constants.Trial,
			requestPath:      "/stable/chef-360/packages?eol=false&v=latest",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"chef-360 not available for the trial and opensource", "status_text":"Bad Request"}`,
			details:          &models.ProductDetails{},
			err:              nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "chef-360 success",
			serverMode:       constants.Commercial,
			requestPath:      "/stable/chef-360/packages?eol=false&v=latest",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"linux": {"pv": {"amd64": {"sha1": "","sha256": "","url": "http://example.com/stable/chef-360/download?eol=false&m=amd64&p=linux&v=latest","version": "latest"}}}}`,
			details: &models.ProductDetails{
				Product: "chef-360",
				Version: "latest",
				MetaData: []models.MetaData{
					{
						Architecture:    "amd64",
						FileName:        "",
						Platform:        "linux",
						PlatformVersion: "",
						SHA1:            "",
						SHA256:          "",
					},
				},
			},
			err:          nil,
			version:      "latest",
			version_err:  nil,
			versions:     []string{"latest"},
			versions_err: nil,
		},
		{
			name:             "db connection  error",
			serverMode:       constants.Trial,
			requestPath:      "/stable/automate/packages?eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			details:          &models.ProductDetails{},
			err:              nil,
			version:          "",
			version_err:      errors.New("ResourceNotFoundException: Requested resource not found"),
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "opensource check success",
			serverMode:       constants.Opensource,
			requestPath:      "/stable/habitat/packages?eol=false",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"linux": {"pv": {"x86_64": {"sha1":"", "sha256":"abcd", "url":"http://example.com/stable/habitat/download?eol=false&m=x86_64&p=linux&v=0.9.3", "version":"0.9.3"}}}}`,
			details: &models.ProductDetails{
				Product: "habitat",
				Version: "0.9.3",
				MetaData: []models.MetaData{{
					Architecture:    "x86_64",
					FileName:        "",
					Platform:        "linux",
					PlatformVersion: "",
					SHA1:            "",
					SHA256:          "abcd",
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
			serverMode:       constants.Opensource,
			requestPath:      "/stable/habitat/packages?eol=false",
			expectedStatus:   fiber.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching product versions", "status_text":"Internal Server Error"}`,
			details:          &models.ProductDetails{},
			err:              nil,
			version:          "",
			version_err:      nil,
			versions:         []string{},
			versions_err:     errors.New("ResourceNotFoundException: Requested resource not found"),
		},
		{
			name:             "empty metadate info",
			serverMode:       constants.Trial,
			requestPath:      "/stable/habitat/packages?eol=false",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Product information not found. Please check the input parameters.", "status_text":"Bad Request"}`,
			details: &models.ProductDetails{
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
			serverMode:       constants.Opensource,
			requestPath:      "/stable/habitat/packages?eol=false&v=0.3.2",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"linux": {"pv": {"x86_64": {"sha1":"", "sha256":"abcd", "url":"http://example.com/stable/habitat/download?eol=false&m=x86_64&p=linux&v=0.3.2", "version":"0.3.2"}}}}`,
			details: &models.ProductDetails{
				Product: "habitat",
				Version: "0.3.2",
				MetaData: []models.MetaData{{
					Architecture:    "x86_64",
					FileName:        "",
					Platform:        "linux",
					PlatformVersion: "",
					SHA1:            "",
					SHA256:          "abcd",
				}},
			},
			err:          nil,
			version:      "",
			version_err:  nil,
			versions:     []string{"0.9.3", "0.3.2", "0.7.11", "0.9.0"},
			versions_err: nil,
		},
		{
			name:           "chef-ice product success",
			serverMode:     constants.Commercial,
			requestPath:    "/stable/chef-ice/packages?eol=false&license_id=tmns-e88dafdb-06e1-4676-908f-87503da14c4d-3413&v=19.7.17",
			expectedStatus: fiber.StatusOK,
			expectedResponse: `{
						"linux": {
							"x86_64": {
								"deb": {
									"sha1": "dcf75b37bb80128af4657501bfd41eac52820191",
									"sha256": "2c501d02b16d67e9d5a28578b95f8d3155bed940ee4946229213f41a2e8b798e",
									"url": "http://example.com/stable/chef-ice/download?eol=false&license_id=tmns-e88dafdb-06e1-4676-908f-87503da14c4d-3413&m=x86_64&p=linux&pm=deb&v=19.7.17",
									"version": "19.7.17"
								}
							}
						}
					}`,
			details: &models.PackageDetails{
				Product: "chef-ice",
				Version: "19.7.17",
				Metadata: map[string]models.Platform{
					"linux": {
						"x86_64": {
							"deb": models.PackageType{
								Filename: "chef-ice_19.7.17_amd64.deb",
								SHA1:     "dcf75b37bb80128af4657501bfd41eac52820191",
								SHA256:   "2c501d02b16d67e9d5a28578b95f8d3155bed940ee4946229213f41a2e8b798e",
							},
						},
					},
				},
			},
			err:          nil,
			version:      "19.7.17",
			version_err:  nil,
			versions:     []string{"19.7.17"},
			versions_err: nil,
		},
		{
			name:           "migrate-ice product success",
			serverMode:     constants.Commercial,
			requestPath:    "/stable/migrate-ice/packages?eol=false&license_id=tmns-e88dafdb-06e1-4676-908f-87503da14c4d-3413&v=19.0.1",
			expectedStatus: fiber.StatusOK,
			expectedResponse: `{
				"linux": {
					"x86_64": {
						"deb": {
							"sha1": "dcf75b37bb80128af4657501bfd41eac52820191",
							"sha256": "2c501d02b16d67e9d5a28578b95f8d3155bed940ee4946229213f41a2e8b798e",
							"url": "http://example.com/stable/migrate-ice/download?eol=false&license_id=tmns-e88dafdb-06e1-4676-908f-87503da14c4d-3413&m=x86_64&p=linux&pm=deb&v=19.0.1",
							"version": "19.0.1"
						}
					}
				}
			}`,
			details: &models.PackageDetails{
				Product: "migrate-ice",
				Version: "19.0.1",
				Metadata: map[string]models.Platform{
					"linux": {
						"x86_64": {
							"deb": models.PackageType{
								Filename: "migration-tool_19.0.1_amd64.deb",
								SHA1:     "dcf75b37bb80128af4657501bfd41eac52820191",
								SHA256:   "2c501d02b16d67e9d5a28578b95f8d3155bed940ee4946229213f41a2e8b798e",
							},
						},
					},
				},
			},
			err:          nil,
			version:      "19.0.1",
			version_err:  nil,
			versions:     []string{"19.0.1"},
			versions_err: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("base_url", "http://example.com")
				return c.Next()
			})
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetPackagesfunc = func(partitionValue, sortValue string) (interface{}, error) {
				return test.details, test.err
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return test.version, test.version_err
			}
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return test.versions, test.versions_err
			}
			mockDbService.SetDbInfofunc = func(tableName string, dbModel reflect.Type) {
			}
			log := logrus.NewEntry(logrus.New())
			handler := NewDownloadsHandler(log)
			app.Use(testInjector(mockDbService, test.serverMode, &template.MockTemplateRenderer{}))
			app.Get("/:channel/:product/packages", func(c *fiber.Ctx) error {
				return handler.ProductPackagesHandler(c)
			})

			req := httptest.NewRequest(http.MethodGet, "http://example.com"+test.requestPath, nil)
			resp, err := app.Test(req, 100*1000) // 100 seconds timeout

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
		serverMode       constants.ApiType
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
			serverMode:       constants.Trial,
			requestPath:      "/current/automate/fileName?p=linux&pv=16.04&m=x86_64&v=latest",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"fileName":"automate_4.7.52-1_amd64.deb"}`,
			metadata:         models.MetaData{FileName: "automate_4.7.52-1_amd64.deb"},
			metadata_err:     nil,
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "HABITAT_SUCCESS",
			serverMode:       constants.Opensource,
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
			versions:         []string{"latest"},
		},
		{
			name:             "failure db connection while fetching latest version",
			serverMode:       constants.Trial,
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=1.6.652",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching product versions", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			metadata_err:     nil,
			version:          "",
			versions_err:     errors.New("Unable to get latest version of habitat"),
			versions:         []string{"1.6.652"},
		},
		{
			name:             "failure channel is not stable/current",
			serverMode:       constants.Trial,
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
			serverMode:       constants.Trial,
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=1.6.652",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			metadata_err:     errors.New("Error while fetching file name"),
			version_err:      nil,
			version:          "1.6.652",
			versions:         []string{"1.6.652", "1.6.651"},
		},
		{
			name:             "chef-360 as a product is given",
			serverMode:       constants.Commercial,
			expectedStatus:   http.StatusOK,
			requestPath:      "/current/chef-360/fileName?p=linux&pv=20.04&m=x86_64&v=latest&license_id=viv2c0a2-111f-2caf-1fa2-1211fe1212d1",
			expectedResponse: `{"fileName":"chef-360.tar.gz"}`,
			metadata:         models.MetaData{},
			metadata_err:     nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
		},
		{
			name:             "chef-360 trial is not supported",
			serverMode:       constants.Trial,
			expectedStatus:   http.StatusBadRequest,
			requestPath:      "/current/chef-360/fileName?p=linux&pv=20.04&m=x86_64&v=1.2&license_id=viv2c0a2-111f-2caf-1fa2-1211fe1212d1",
			expectedResponse: `{"code":400, "message":"chef-360 not available for the trial and opensource", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			metadata_err:     nil,
			version:          "latest",
			version_err:      nil,
		},
		{
			name:             "haitat opensource verion not supported ",
			serverMode:       constants.Opensource,
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=0.79.0",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"the requested version is not supported on the selected persona or channel", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			metadata_err:     errors.New("Error while fetching file name"),
			version_err:      nil,
			version:          "",
			versions:         []string{"0.79.0", "0.78.0"},
			versions_err:     nil,
		},
		{
			name:             "haitat opensource version fetching error",
			serverMode:       constants.Opensource,
			requestPath:      "/current/habitat/fileName?p=linux&pv=16.04&m=x86_64&v=0.79.0",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching product versions", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			metadata_err:     errors.New("Error while fetching file name"),
			version_err:      nil,
			version:          "",
			versions:         []string{},
			versions_err:     errors.New("Unable to get all versions of habitat"),
		},
		{
			name:             "chef-ice SUCCESS",
			serverMode:       constants.Commercial,
			requestPath:      "/current/chef-ice/fileName?p=windows&pv=pv&m=x86_64&pm=msi&v=latest&license_id=viv2c0a2-111f-2caf-1fa2-1211fe1212d1",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"fileName":"chef-19.0.1-windows-x86_64.msi"}`,
			metadata:         models.MetaData{FileName: "chef-19.0.1-windows-x86_64.msi"},
			metadata_err:     nil,
			version_err:      nil,
			version:          "",
			versions:         []string{"19.0.1"},
			versions_err:     nil,
		},
		{
			name:             "chef-ice failure for db error",
			requestPath:      "/current/chef-ice/fileName?p=windows&pv=pv&m=x86_64&pm=msi&v=latest&license_id=viv2c0a2-111f-2caf-1fa2-1211fe1212d1",
			serverMode:       constants.Commercial,
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Error while fetching the information for the product from DB.", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			metadata_err:     nil,
			version_err:      errors.New("Unable to get latest version of chef-ice"),
			versions:         []string{"latest"},
		},
		{
			name:             "chef-ice not supported for opensource server",
			serverMode:       constants.Opensource,
			requestPath:      "/current/chef-ice/fileName?p=windows&pv=pv&m=x86_64&pm=msi&v=latest",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"No versions found for this product/mode", "status_text":"Bad Request"}`,
			metadata:         models.MetaData{},
			metadata_err:     nil,
			version_err:      nil,
			version:          "",
			versions:         []string{""},
			versions_err:     nil,
		},
		{
			name:             "chef-ice parameter missing",
			requestPath:      "/current/chef-ice/fileName?p=windows&pv=pv&m=x86_64&v=latest&license_id=viv2c0a2-111f-2caf-1fa2-1211fe1212d1",
			serverMode:       constants.Commercial,
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500, "message":"Package Manager (pm) params cannot be empty", "status_text":"Internal Server Error"}`,
			metadata:         models.MetaData{},
			version_err:      nil,
			versions:         []string{"19.0.1"},
		},
		{
			name:             "chef-ice fileName success",
			serverMode:       constants.Commercial,
			requestPath:      "/current/chef-ice/fileName?p=linux&pv=20.04&m=x86_64&pm=deb&v=latest",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"fileName":"chef-ice_1.0.0_amd64.deb"}`,
			metadata:         models.MetaData{FileName: "chef-ice_1.0.0_amd64.deb"},
			metadata_err:     nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
		{
			name:             "migrate-ice fileName success",
			serverMode:       constants.Commercial,
			requestPath:      "/current/migrate-ice/fileName?p=linux&pv=20.04&m=x86_64&pm=deb&v=latest",
			expectedStatus:   http.StatusOK,
			expectedResponse: `{"fileName":"migrate-ice_1.0.0_amd64.deb"}`,
			metadata:         models.MetaData{FileName: "migrate-ice_1.0.0_amd64.deb"},
			metadata_err:     nil,
			version:          "latest",
			version_err:      nil,
			versions:         []string{"latest"},
			versions_err:     nil,
		},
	}

	for _, test := range tests {
		timeout := 1 * time.Minute
		t.Run(test.name, func(t *testing.T) {
			done := make(chan struct{})
			go func() {
				defer close(done)
				app := fiber.New()
				app.Use(func(c *fiber.Ctx) error {
					c.Locals("base_url", "http://example.com")
					return c.Next()
				})
				mockDbService := new(dboperations.MockIDbOperations)

				mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture, packageManager string) (*models.MetaData, error) {
					return &test.metadata, test.metadata_err
				}
				mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
					return test.version, test.version_err
				}
				mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
					return test.versions, test.versions_err
				}
				mockDbService.SetDbInfofunc = func(tableName string, dbModel reflect.Type) {
				}

				log := logrus.NewEntry(logrus.New())
				handler := NewDownloadsHandler(log)
				app.Use(testInjector(mockDbService, test.serverMode, &template.MockTemplateRenderer{}))
				app.Get("/:channel/:product/fileName", func(c *fiber.Ctx) error {
					return handler.FileNameHandler(c)
				})

				req := httptest.NewRequest(http.MethodGet, "http://example.com"+test.requestPath, nil)
				resp, err := app.Test(req, 100*1000) // 100 seconds timeout

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
		serverMode       constants.ApiType
		mockTemplate     func(baseUrl string, params *omnitruck.RequestParams, filepath string) (string, error)
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("base_url", "http://example.com")
				return c.Next()
			})
			mockTemplate := new(template.MockTemplateRenderer)
			mockTemplate.GetScriptfunc = test.mockTemplate
			log := logrus.NewEntry(logrus.New())
			handler := NewDownloadsHandler(log)

			app.Use(testInjector(&dboperations.MockIDbOperations{}, test.serverMode, mockTemplate))
			app.Get("/install.sh", func(c *fiber.Ctx) error {
				return handler.DownloadLinuxScript(c)
			})

			req := httptest.NewRequest(http.MethodGet, test.requestPath, nil)
			resp, err := app.Test(req, 100*1000) // 100 seconds timeout

			assert.NoError(t, err)
			assert.Equal(t, test.expectedStatus, resp.StatusCode)
		})
	}
}

func TestDownloadWindowsScriptHandler(t *testing.T) {
	tests := []struct {
		name             string
		serverMode       constants.ApiType
		mockTemplate     func(baseUrl string, params *omnitruck.RequestParams, filepath string) (string, error)
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("base_url", "http://example.com")
				return c.Next()
			})
			mockTemplate := new(template.MockTemplateRenderer)
			mockTemplate.GetScriptfunc = test.mockTemplate
			log := logrus.NewEntry(logrus.New())
			handler := NewDownloadsHandler(log)

			app.Use(testInjector(&dboperations.MockIDbOperations{}, test.serverMode, mockTemplate))

			app.Get("/install.ps1", func(c *fiber.Ctx) error {
				return handler.DownloadWindowsScript(c)
			})

			req := httptest.NewRequest(http.MethodGet, test.requestPath, nil)
			resp, err := app.Test(req, 100*1000) // 100 seconds timeout
			assert.NoError(t, err)
			assert.Equal(t, test.expectedStatus, resp.StatusCode)
		})
	}
}

func TestPackageManagersHandler(t *testing.T) {
	tests := []struct {
		name             string
		mockData         []string
		mockErr          error
		expectedStatus   int
		expectedResponse string
		mode             constants.ApiType
	}{
		{
			name:             "Success - package managers fetched",
			mockData:         []string{"deb", "tar", "rpm"},
			mockErr:          nil,
			expectedStatus:   http.StatusOK,
			expectedResponse: `["deb","tar","rpm"]`,
			mode:             constants.Commercial,
		},
		{
			name:             "Error - DB call fails",
			mockData:         nil,
			mockErr:          errors.New("db failure"),
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: `{"code":500,"message":"Error while fetching the information for the product from DB.","status_text":"Internal Server Error"}`,
			mode:             constants.Commercial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetPackageManagersfunc = func() ([]string, error) {
				return tt.mockData, tt.mockErr
			}

			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				c.Locals("base_url", "http://example.com")
				return c.Next()
			})
			log := logrus.NewEntry(logrus.New())
			handler := NewDownloadsHandler(log)

			app.Use(testInjector(mockDbService, tt.mode, &template.MockTemplateRenderer{}))
			app.Get("/package-managers", func(c *fiber.Ctx) error {
				return handler.PackageManagersHandler(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/package-managers", nil)
			resp, err := app.Test(req, 100*1000) // 100 seconds timeout
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			assert.JSONEq(t, tt.expectedResponse, string(bodyBytes))
		})
	}
}

// TestProductDownloadHandler_MandatoryFlags tests only the mandatory flag validations for ProductDownloadHandler
func TestProductDownloadHandler(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	handler := NewDownloadsHandler(log)

	tests := []struct {
		name             string
		requestPath      string
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "missing platform",
			requestPath:      "/current/chef/download?pv=20.04&m=x86_64&pm=tar",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Platfrom (p) params cannot be empty", "status_text":"Bad Request"}`,
		},
		{
			name:             "missing platform version",
			requestPath:      "/current/chef/download?p=ubuntu&m=x86_64&pm=tar",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Platform Version (pv) params cannot be empty", "status_text":"Bad Request"}`,
		},
		{
			name:             "missing architecture",
			requestPath:      "/current/chef/download?p=ubuntu&pv=20.04&pm=tar",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Architecture (m) params cannot be empty", "status_text":"Bad Request"}`,
		},
		{
			name:             "package manager parameter missing",
			requestPath:      "/stable/chef-ice/download?p=linux&m=amd64&eol=false&v=latest",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Package Manager (pm) params cannot be empty", "status_text":"Bad Request"}`,
		},
		{
			name:             "package manager parameter missing",
			requestPath:      "/stable/migrate-ice/download?p=linux&m=amd64&eol=false&v=latest",
			expectedStatus:   fiber.StatusBadRequest,
			expectedResponse: `{"code":400, "message":"Package Manager (pm) params cannot be empty", "status_text":"Bad Request"}`,
		},
		{
			name:             "package manager auto add for automate",
			requestPath:      "/stable/automate/download?p=linux&m=amd64&eol=false&v=latest",
			expectedStatus:   fiber.StatusOK,
			expectedResponse: `{"code":400, "message":"Package Manager (pm) params cannot be empty", "status_text":"Bad Request"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			// Set up DownloadService with necessary mocks (see other tests for pattern)
			mockDbService := new(dboperations.MockIDbOperations)
			mockDbService.GetMetaDatafunc = func(partitionValue, sortValue, platform, platformVersion, architecture, packageManager string) (*models.MetaData, error) {
				return &models.MetaData{}, nil
			}
			mockDbService.GetVersionLatestfunc = func(partitionValue string) (string, error) {
				return "latest", nil
			}
			mockDbService.GetVersionAllfunc = func(partitionValue string) ([]string, error) {
				return []string{"latest"}, nil
			}
			mockDbService.SetDbInfofunc = func(tableName string, dbModel reflect.Type) {}

			app.Use(testInjector(mockDbService, constants.Commercial, &template.MockTemplateRenderer{}))
			app.Get("/:channel/:product/download", func(c *fiber.Ctx) error {
				return handler.ProductDownloadHandler(c)
			})
			req := httptest.NewRequest(http.MethodGet, "http://example.com"+test.requestPath, nil)
			resp, err := app.Test(req, 100*1000)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedStatus, resp.StatusCode)
			if test.expectedStatus != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				assert.JSONEq(t, test.expectedResponse, string(bodyBytes))
			}
		})
	}
}

func TestPlatformsHandler(t *testing.T) {
	t.Parallel()

	log := logrus.NewEntry(logrus.New())

	tests := []struct {
		name             string
		inject           bool
		customMiddleware func(*fiber.Ctx) error
		expectedStatus   int
		expectedContains string
	}{
		{
			name:             "missing injector returns 500",
			inject:           false,
			expectedStatus:   http.StatusInternalServerError,
			expectedContains: "Not able to process the request.", // match actual handler
		},
		{
			name:   "broken injector returns 500 on NewDownloadService failure",
			inject: true,
			customMiddleware: func(c *fiber.Ctx) error {
				reqInjector := do.New()
				c.Locals("reqinjector", reqInjector)
				return c.Next()
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedContains: "Failed to create download service",
		},
		{
			name:   "platforms returns success",
			inject: true,
			customMiddleware: testInjector(
				&dboperations.MockIDbOperations{},
				constants.Commercial,
				&template.MockTemplateRenderer{},
			),
			expectedStatus:   http.StatusOK,
			expectedContains: "{",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(func(c *fiber.Ctx) error {
				c.Locals("base_url", "http://example.com")
				return c.Next()
			})

			if tt.inject {
				app.Use(tt.customMiddleware)
			}

			handler := NewDownloadsHandler(log)
			app.Get("/platforms", handler.PlatformsHandler)

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/platforms", nil), 2000)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, string(body))
			assert.Contains(t, string(body), tt.expectedContains)
		})
	}
}

func TestArchitecturesHandler(t *testing.T) {
	t.Parallel()

	log := logrus.NewEntry(logrus.New())

	tests := []struct {
		name             string
		inject           bool
		customMiddleware func(*fiber.Ctx) error
		expectedStatus   int
		expectedContains string
	}{
		{
			name:             "missing injector returns 500",
			inject:           false,
			expectedStatus:   http.StatusInternalServerError,
			expectedContains: "Not able to process the request.",
		},
		{
			name:   "broken injector returns 500 on NewDownloadService failure",
			inject: true,
			customMiddleware: func(c *fiber.Ctx) error {
				reqInjector := do.New()
				c.Locals("reqinjector", reqInjector)
				return c.Next()
			},
			expectedStatus:   http.StatusInternalServerError,
			expectedContains: "Failed to create download service",
		},
		{
			name:   "architectures returns success",
			inject: true,
			customMiddleware: testInjector(
				&dboperations.MockIDbOperations{},
				constants.Commercial,
				&template.MockTemplateRenderer{},
			),
			expectedStatus:   http.StatusOK,
			expectedContains: "[",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(func(c *fiber.Ctx) error {
				c.Locals("base_url", "http://example.com")
				return c.Next()
			})

			if tt.inject {
				app.Use(tt.customMiddleware)
			}

			handler := NewDownloadsHandler(log)
			app.Get("/architectures", handler.ArchitecturesHandler)

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/architectures", nil), 2000)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, string(body))
			assert.Contains(t, string(body), tt.expectedContains)
		})
	}
}

func TestRelatedProductsHandler_MissingInjector(t *testing.T) {
	t.Parallel()
	log := logrus.NewEntry(logrus.New())

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("base_url", "http://example.com")
		return c.Next()
	})

	handler := NewDownloadsHandler(log)
	app.Get("/relatedProducts", handler.RelatedProductsHandler)

	req := httptest.NewRequest(http.MethodGet, "/relatedProducts?bom=foobar", nil)
	resp, err := app.Test(req, 10_000)
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, string(body))
	assert.Contains(t, string(body), "Not able to process the request.")
}

func TestRelatedProductsHandler_BrokenInjector(t *testing.T) {
	t.Parallel()
	log := logrus.NewEntry(logrus.New())

	app := fiber.New()

	injector := do.New()
	// validator provided so ValidateRequest does not panic
	do.ProvideNamedValue[omnitruck.IRequestValidator](injector, "validator", &omnitruck.MockRequestValidator{
		ParamsFunc: func(params *omnitruck.RequestParams, ctx omnitruck.Context) []*omnitruck.ValidationError {
			return nil
		},
		ErrorMessagesFunc: func(errors []*omnitruck.ValidationError) (string, int) {
			return "", 0
		},
	})
	// do not register other dependencies
	// so NewDownloadService will fail

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("base_url", "http://example.com")
		c.Locals("reqinjector", injector)
		return c.Next()
	})

	handler := NewDownloadsHandler(log)
	app.Get("/relatedProducts", handler.RelatedProductsHandler)

	req := httptest.NewRequest(http.MethodGet, "/relatedProducts?bom=foobar", nil)
	resp, err := app.Test(req, 10_000)
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, string(body))
	assert.Contains(t, string(body), "Failed to create download service")
}
