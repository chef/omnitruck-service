package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"

	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/dboperations"
	"github.com/chef/omnitruck-service/models"
	"github.com/chef/omnitruck-service/utils/template"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildInjector(templateRenderer template.TemplateRenderer, omnitruckURL string) *do.Injector {
	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{
		GetVersionAllfunc: func(partitionValue string) ([]string, error) {
			return []string{"1.0.0", "2.0.0"}, nil
		},
	})
	do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", templateRenderer)
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{
		LicenseServiceUrl: "http://license-service",
		OmnitruckUrl:      omnitruckURL,
		SupportInfra19:    true,
	})
	return injector
}

func TestNewDownloadService_MissingDbService(t *testing.T) {
	injector := do.New()
	do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", &template.MockTemplateRenderer{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	log := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, log, map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewDownloadService_MissingTemplateRenderer(t *testing.T) {
	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	log := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, log, map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewDownloadService_MissingReplicated(t *testing.T) {
	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", &template.MockTemplateRenderer{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	log := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, log, map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewDownloadService_MissingLicenseClient(t *testing.T) {
	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", &template.MockTemplateRenderer{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	log := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, log, map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewDownloadService_MissingMode(t *testing.T) {
	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", &template.MockTemplateRenderer{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	log := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, log, map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewDownloadService_MissingConfig(t *testing.T) {
	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", &template.MockTemplateRenderer{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	log := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, log, map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestNewDownloadService_Success(t *testing.T) {
	mockTemplate := &template.MockTemplateRenderer{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			return "mock script", nil
		},
	}
	injector := buildInjector(mockTemplate, "https://omnitruck.chef.io")

	log := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, log, map[string]interface{}{"license_id": "123"})
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestNewDownloadService_MissingDependency(t *testing.T) {
	injector := do.New()
	log := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, log, map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestOmnitruck(t *testing.T) {
	mockTemplate := &template.MockTemplateRenderer{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			return "mock script", nil
		},
	}
	injector := buildInjector(mockTemplate, "https://omnitruck.chef.io")

	log := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, log, map[string]interface{}{})
	require.NoError(t, err)

	client := svc.Omnitruck()
	assert.NotNil(t, client)
}

func TestPlatformServices(t *testing.T) {
	mockTemplate := &template.MockTemplateRenderer{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			return "mock script", nil
		},
	}
	injector := buildInjector(mockTemplate, "https://omnitruck.chef.io")

	log := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, log, map[string]interface{}{})
	require.NoError(t, err)

	plat := svc.PlatformServices()
	assert.NotNil(t, plat)
}

func TestReplicatedService(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	svc := DownloadService{}
	rep := svc.ReplicatedService(config.ReplicatedConfig{}, log)
	assert.NotNil(t, rep)
}

func TestGetLinuxScript(t *testing.T) {
	tests := []struct {
		name           string
		mode           constants.ApiType
		locals         map[string]interface{}
		params         *omnitruck.RequestParams
		omnitruckUrl   string
		mockResponse   string
		mockStatusCode int
	}{
		{
			name:           "commercial mode with license_id",
			mode:           constants.Commercial,
			locals:         map[string]interface{}{"base_url": "http://x"},
			params:         &omnitruck.RequestParams{LicenseId: "test-license"},
			mockResponse:   "#!/bin/bash\n# License ID provided via context\nlicense_id='test-license'\ninstall script",
			mockStatusCode: 200,
		},
		{
			name:           "trial mode with license_id",
			mode:           constants.Trial,
			locals:         map[string]interface{}{"base_url": "http://x"},
			params:         &omnitruck.RequestParams{LicenseId: "trial-license"},
			mockResponse:   "#!/bin/bash\n# License ID provided via context\nlicense_id='trial-license'\ninstall script",
			mockStatusCode: 200,
		},
		{
			name:           "opensource mode without license_id",
			mode:           constants.Opensource,
			locals:         map[string]interface{}{"base_url": "http://x"},
			params:         &omnitruck.RequestParams{},
			mockResponse:   "#!/bin/bash\ninstall script",
			mockStatusCode: 200,
		},
		{
			name:           "with base_url parameter",
			mode:           constants.Opensource,
			locals:         map[string]interface{}{"base_url": "http://x"},
			params:         &omnitruck.RequestParams{BaseUrl: "https://custom.chef.io"},
			mockResponse:   "#!/bin/bash\nbase_api_url=\"https://custom.chef.io\"\ninstall script",
			mockStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/install.sh", r.URL.Path)
				if tt.params.LicenseId != "" {
					assert.Equal(t, tt.params.LicenseId, r.URL.Query().Get("license_id"))
				}
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			mockTemplate := &template.MockTemplateRenderer{}
			injector := buildInjector(mockTemplate, server.URL)
			log := logrus.NewEntry(logrus.New())

			svc, _ := NewDownloadService(injector, log, tt.locals)
			svc.setMode(tt.mode)
			script, req := svc.GetLinuxScript(tt.params)

			assert.True(t, req.Ok)
			assert.Equal(t, fiber.StatusOK, req.Code)
			assert.NotEmpty(t, script)
			assert.Contains(t, script, "install")
		})
	}
}

func TestGetWindowsScript(t *testing.T) {
	tests := []struct {
		name           string
		mode           constants.ApiType
		locals         map[string]interface{}
		params         *omnitruck.RequestParams
		omnitruckUrl   string
		mockResponse   string
		mockStatusCode int
	}{
		{
			name:           "commercial mode with license_id",
			mode:           constants.Commercial,
			locals:         map[string]interface{}{"base_url": "http://x"},
			params:         &omnitruck.RequestParams{LicenseId: "test-license"},
			mockResponse:   "# PowerShell install script\n# License ID provided via context - adding to install command\ninstall -license_id 'test-license'",
			mockStatusCode: 200,
		},
		{
			name:           "trial mode with license_id",
			mode:           constants.Trial,
			locals:         map[string]interface{}{"base_url": "http://x"},
			params:         &omnitruck.RequestParams{LicenseId: "trial-license"},
			mockResponse:   "# PowerShell install script\n# License ID provided via context - adding to install command\ninstall -license_id 'trial-license'",
			mockStatusCode: 200,
		},
		{
			name:           "opensource mode without license_id",
			mode:           constants.Opensource,
			locals:         map[string]interface{}{"base_url": "http://x"},
			params:         &omnitruck.RequestParams{},
			mockResponse:   "# PowerShell install script",
			mockStatusCode: 200,
		},
		{
			name:           "with base_url parameter",
			mode:           constants.Opensource,
			locals:         map[string]interface{}{"base_url": "http://x"},
			params:         &omnitruck.RequestParams{BaseUrl: "https://custom.chef.io"},
			mockResponse:   "# PowerShell install script\n$base_server_uri = \"https://custom.chef.io\"",
			mockStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/install.ps1", r.URL.Path)
				if tt.params.LicenseId != "" {
					assert.Equal(t, tt.params.LicenseId, r.URL.Query().Get("license_id"))
				}
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			mockTemplate := &template.MockTemplateRenderer{}
			injector := buildInjector(mockTemplate, server.URL)
			log := logrus.NewEntry(logrus.New())

			svc, _ := NewDownloadService(injector, log, tt.locals)
			svc.setMode(tt.mode)
			script, req := svc.GetWindowsScript(tt.params)

			assert.True(t, req.Ok)
			assert.Equal(t, fiber.StatusOK, req.Code)
			assert.NotEmpty(t, script)
			assert.Contains(t, script, "install")
		})
	}
}

func TestProductDownload_FailsOnEmptyVersions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/versions/all")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`not found`))
	}))
	defer ts.Close()

	mockTemplate := &template.MockTemplateRenderer{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			return "script", nil
		},
	}
	injector := buildInjector(mockTemplate, "https://omnitruck.chef.io")

	do.OverrideNamed[config.ServiceConfig](injector, "config", func(i *do.Injector) (config.ServiceConfig, error) {
		return config.ServiceConfig{
			LicenseServiceUrl: "http://license-service",
			OmnitruckUrl:      ts.URL,
		}, nil
	})

	log := logrus.NewEntry(logrus.New())
	svc, _ := NewDownloadService(injector, log, map[string]interface{}{"base_url": "http://x"})
	params := &omnitruck.RequestParams{Product: "chef", Channel: "stable"}
	_, _, _, msg, code, _ := svc.ProductDownload(params, &fiber.Ctx{})

	assert.NotEqual(t, fiber.StatusOK, code)
	assert.NotEmpty(t, msg)
}

func TestDownloadService_Products(t *testing.T) {
	log := logrus.NewEntry(logrus.New())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/products", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["chef", "chef-server"]`))
	}))
	defer ts.Close()

	tests := []struct {
		name     string
		eolParam string
	}{
		{
			name:     "products with eol=false",
			eolParam: "false",
		},
		{
			name:     "products with eol=true",
			eolParam: "true",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			injector := do.New()

			do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{
				GetPackagesfunc: func(partitionValue, sortValue string) (interface{}, error) {
					return nil, nil
				},
				GetVersionAllfunc: func(partitionValue string) ([]string, error) {
					return []string{"1.0.0"}, nil
				},
				GetRelatedProductsfunc: func(partitionValue string) (*models.RelatedProducts, error) {
					return &models.RelatedProducts{
						Bom: "example-bom",
						Products: map[string]string{
							"chef": "Chef Infra",
						},
					}, nil
				},
				GetPackageManagersfunc: func() ([]string, error) {
					return []string{"apt"}, nil
				},
				SetDbInfofunc: func(tableName string, dbModel reflect.Type) {},
			})

			do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", &template.MockTemplateRenderer{
				GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
					return "script", nil
				},
			})

			do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
			do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
			do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)

			do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{
				OmnitruckUrl: ts.URL,
			})

			locals := map[string]interface{}{
				"license_id": "dummy-license",
				"base_url":   "http://example.com",
			}

			svc, err := NewDownloadService(injector, log, locals)
			assert.NoError(t, err)

			params := &omnitruck.RequestParams{
				Eol: tt.eolParam,
			}

			data, req := svc.Products(params)

			assert.NotNil(t, req)
			assert.True(t, req.Ok)
			assert.Equal(t, 200, req.Code)
			assert.NotNil(t, data)
		})
	}
}

func TestDownloadService_Platforms(t *testing.T) {
	t.Parallel()

	logEntry := logrus.NewEntry(logrus.New())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/platforms", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ubuntu":"Ubuntu","debian":"Debian"}`))
	}))
	defer ts.Close()

	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{
		GetRelatedProductsfunc: func(partitionValue string) (*models.RelatedProducts, error) {
			return &models.RelatedProducts{Products: map[string]string{"test": "test"}}, nil
		},
	})
	do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", &template.MockTemplateRenderer{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{
		OmnitruckUrl:   ts.URL,
		SupportInfra19: true,
	})

	locals := map[string]interface{}{"license_id": "123"}
	svc, err := NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	data, req := svc.Platforms()

	assert.True(t, req.Ok, "expected request to be OK")
	assert.NotEmpty(t, data, "expected platform list from Omnitruck to be non-empty")
}

func TestDownloadService_Architectures(t *testing.T) {
	t.Parallel()

	logEntry := logrus.NewEntry(logrus.New())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/architectures", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`["x86_64","arm64"]`))
	}))
	defer ts.Close()

	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", &template.MockTemplateRenderer{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{
		OmnitruckUrl: ts.URL,
	})

	locals := map[string]interface{}{"license_id": "123"}
	svc, err := NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	data, req := svc.Architectures()

	assert.True(t, req.Ok, "expected request to be OK")
	assert.NotEmpty(t, data, "expected architectures list from Omnitruck to be non-empty")
}
func TestDownloadService_LatestVersion(t *testing.T) {
	t.Parallel()

	logEntry := logrus.NewEntry(logrus.New())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "versions/all") {
			w.Header().Set("Content-Type", "application/json")
			if strings.Contains(r.URL.Path, "invalid-product") {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`["1.0.0","2.0.0"]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	mockTemplate := &template.MockTemplateRenderer{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			return "mock script", nil
		},
	}
	injector := buildInjector(mockTemplate, ts.URL)

	locals := map[string]interface{}{"license_id": "123"}
	svc, err := NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	svc.setMode(constants.Opensource)

	tests := []struct {
		name          string
		params        *omnitruck.RequestParams
		expectSuccess bool
	}{
		{
			name: "success",
			params: &omnitruck.RequestParams{
				Product: "chef",
				Channel: "stable",
			},
			expectSuccess: true,
		},
		{
			name: "error on invalid product",
			params: &omnitruck.RequestParams{
				Product: "invalid-product",
				Channel: "stable",
			},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			data, req := svc.LatestVersion(tt.params)
			if tt.expectSuccess {
				assert.True(t, req.Ok, "expected LatestVersion to succeed")
				assert.NotEmpty(t, data, "expected a latest version")
			} else {
				assert.False(t, req.Ok, "expected LatestVersion to fail")
				assert.Empty(t, data, "expected no latest version")
			}
		})
	}
}

func TestDownloadService_ProductVersions(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/stable/inspec/versions/all":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`["4.0.0", "5.0.0"]`))
		case "/stable/invalid-product/versions/all":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`not found`))
		default:
			http.NotFound(w, r)
		}
	}))

	mockTemplate := &template.MockTemplateRenderer{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			return "mock script", nil
		},
	}

	injector := buildInjector(mockTemplate, ts.URL)

	logEntry := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, logEntry, map[string]interface{}{
		"license_id": "test",
	})
	require.NoError(t, err)

	tests := []struct {
		name          string
		params        *omnitruck.RequestParams
		expectSuccess bool
	}{
		{
			name: "success",
			params: &omnitruck.RequestParams{
				Product: "inspec",
				Channel: "stable",
			},
			expectSuccess: true,
		},
		{
			name: "error on invalid product",
			params: &omnitruck.RequestParams{
				Product: "invalid-product",
				Channel: "stable",
			},
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			data, req := svc.ProductVersions(tt.params)
			if tt.expectSuccess {
				assert.True(t, req.Ok, "expected ProductVersions to succeed")
				assert.NotEmpty(t, data, "expected non-empty versions")
			} else {
				assert.False(t, req.Ok, "expected ProductVersions to fail")
				assert.Empty(t, data, "expected no versions")
			}
		})
	}
}

func TestDownloadService_ProductPackages(t *testing.T) {
	t.Parallel()

	logEntry := logrus.NewEntry(logrus.New())
	locals := map[string]interface{}{"base_url": "http://example.com"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "invalid-product"):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`not found`))
		case strings.Contains(r.URL.Path, "/versions/all"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`["16.0.0", "17.0.0"]`))
		case strings.Contains(r.URL.Path, "/packages"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"chef":{"16.0.0":{"x86_64":{"url":"http://example.com/download","version":"16.0.0"}}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`not found`))
		}
	}))
	defer ts.Close()

	// mock template renderer
	mockTemplate := &template.MockTemplateRenderer{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			return "mock script", nil
		},
	}

	injector := buildInjector(mockTemplate, ts.URL)

	svc, err := NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	tests := []struct {
		name          string
		params        *omnitruck.RequestParams
		expectSuccess bool
		expectCode    int
	}{
		{
			name: "success returns product packages",
			params: &omnitruck.RequestParams{
				Product: "chef",
				Channel: "stable",
				Version: "16.0.0",
			},
			expectSuccess: true,
			expectCode:    fiber.StatusOK,
		},
		{
			name: "error on invalid product (getFilteredVersions fails)",
			params: &omnitruck.RequestParams{
				Product: "invalid-product",
				Channel: "stable",
			},
			expectSuccess: false,
			expectCode:    fiber.StatusBadRequest,
		},
		{
			name: "error on ValidateOrSetVersion fails",
			params: &omnitruck.RequestParams{
				Product: "chef",
				Channel: "stable",
				Version: "invalid-version",
			},
			expectSuccess: false,
			expectCode:    fiber.StatusBadRequest,
		},
		{
			name: "error on GetPackages fails (nonexistent version)",
			params: &omnitruck.RequestParams{
				Product: "chef",
				Channel: "stable",
				Version: "0.0.0",
			},
			expectSuccess: false,
			expectCode:    fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			data, req := svc.ProductPackages(tt.params)

			if tt.expectSuccess {
				assert.True(t, req.Ok, "expected ok true")
				assert.Equal(t, tt.expectCode, req.Code)
				assert.NotNil(t, data)
			} else {
				assert.False(t, req.Ok, "expected ok false")
				assert.Equal(t, tt.expectCode, req.Code)
				assert.Nil(t, data)
			}
		})
	}
}

func TestDownloadService_ProductMetadata(t *testing.T) {
	t.Parallel()

	logEntry := logrus.NewEntry(logrus.New())
	locals := map[string]interface{}{"base_url": "http://example.com"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/metadata"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"sha1":"fake-sha1",
				"sha256":"fake-sha256",
				"url":"http://original-download",
				"version":"16.0.0"
			}`))
		case strings.Contains(r.URL.Path, "invalid-product"):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`not found`))
		case strings.Contains(r.URL.Path, "/versions/all"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`["16.0.0","17.0.0"]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	mockTemplate := &template.MockTemplateRenderer{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			return "mock script", nil
		},
	}

	injector := buildInjector(mockTemplate, ts.URL)

	svc, err := NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	tests := []struct {
		name          string
		params        *omnitruck.RequestParams
		expectSuccess bool
		expectCode    int
	}{
		{
			name: "success returns metadata",
			params: &omnitruck.RequestParams{
				Product:         "chef",
				Channel:         "stable",
				Version:         "16.0.0",
				Platform:        "ubuntu",
				PlatformVersion: "20.04",
				Architecture:    "x86_64",
			},
			expectSuccess: true,
			expectCode:    fiber.StatusOK,
		},
		{
			name: "invalid product returns error",
			params: &omnitruck.RequestParams{
				Product:         "invalid-product",
				Channel:         "stable",
				Version:         "1.0.0",
				Platform:        "ubuntu",
				PlatformVersion: "20.04",
				Architecture:    "x86_64",
			},
			expectSuccess: false,
			expectCode:    fiber.StatusBadRequest,
		},
		{
			name: "invalid version returns error",
			params: &omnitruck.RequestParams{
				Product:         "chef",
				Channel:         "stable",
				Version:         "invalid-version",
				Platform:        "ubuntu",
				PlatformVersion: "20.04",
				Architecture:    "x86_64",
			},
			expectSuccess: false,
			expectCode:    fiber.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			data, req := svc.ProductMetadata(tt.params)
			if tt.expectSuccess {
				assert.True(t, req.Ok, "expected ProductMetadata to succeed")
				assert.Equal(t, tt.expectCode, req.Code)
				assert.NotEmpty(t, data.Url, "expected a remapped download URL")
				expectedUrl := "http://example.com/stable/chef/download?m=x86_64&p=ubuntu&pv=20.04&v=16.0.0"
				assert.Equal(t, expectedUrl, data.Url, "should remap to local download")
			} else {
				assert.False(t, req.Ok, "expected ProductMetadata to fail")
				assert.Equal(t, tt.expectCode, req.Code)
			}
		})
	}
}
func TestDownloadService_RelatedProducts(t *testing.T) {
	t.Parallel()

	logEntry := logrus.NewEntry(logrus.New())
	locals := map[string]interface{}{"base_url": "http://example.com"}

	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{
		GetRelatedProductsfunc: func(partitionValue string) (*models.RelatedProducts, error) {
			if partitionValue == "valid-bom" {
				return &models.RelatedProducts{
					Bom:      "valid-bom",
					Products: map[string]string{"chef": "Chef Infra Client"},
				}, nil
			}
			return nil, fmt.Errorf("forced dynamodb error")
		},
	})
	do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", &template.MockTemplateRenderer{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	svc, err := NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	tests := []struct {
		name          string
		params        *omnitruck.RequestParams
		expectSuccess bool
		expectCode    int
	}{
		{
			name: "success returns related products",
			params: &omnitruck.RequestParams{
				BOM: "valid-bom",
			},
			expectSuccess: true,
			expectCode:    fiber.StatusOK,
		},
		{
			name: "error when dynamo returns error",
			params: &omnitruck.RequestParams{
				BOM: "invalid-bom",
			},
			expectSuccess: false,
			expectCode:    fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			data, req := svc.RelatedProducts(tt.params)

			if tt.expectSuccess {
				assert.True(t, req.Ok, "expected ok true")
				assert.Equal(t, tt.expectCode, req.Code)
				assert.NotNil(t, data)
				assert.Contains(t, data["relatedProducts"], "chef")
			} else {
				assert.False(t, req.Ok, "expected ok false")
				assert.Equal(t, tt.expectCode, req.Code)
				assert.Nil(t, data)
			}
		})
	}
}

func TestDownloadService_RelatedProducts_SupportInfra19(t *testing.T) {
	t.Parallel()

	logEntry := logrus.NewEntry(logrus.New())
	locals := map[string]interface{}{"base_url": "http://example.com"}

	tests := []struct {
		name               string
		supportInfra19     bool
		expectedProducts   []string
		unexpectedProducts []string
	}{
		{
			name:               "SupportInfra19 true - includes infra products",
			supportInfra19:     true,
			expectedProducts:   []string{"chef", "chef-ice", "migrate-ice"},
			unexpectedProducts: []string{},
		},
		{
			name:               "SupportInfra19 false - excludes infra products",
			supportInfra19:     false,
			expectedProducts:   []string{"chef"},
			unexpectedProducts: []string{"chef-ice", "migrate-ice"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			injector := do.New()
			do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{
				GetRelatedProductsfunc: func(partitionValue string) (*models.RelatedProducts, error) {
					return &models.RelatedProducts{
						Bom: "test-bom",
						Products: map[string]string{
							"chef":        "Chef Infra Client",
							"chef-ice":    "Chef Infra Client Enterprise",
							"migrate-ice": "Migrate ICE",
							"automate":    "Chef Automate",
						},
					}, nil
				},
			})
			do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", &template.MockTemplateRenderer{})
			do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
			do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
			do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
			do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{
				SupportInfra19: tt.supportInfra19,
			})

			svc, err := NewDownloadService(injector, logEntry, locals)
			require.NoError(t, err)

			params := &omnitruck.RequestParams{
				BOM: "test-bom",
			}

			data, req := svc.RelatedProducts(params)

			assert.True(t, req.Ok, "expected ok true")
			assert.Equal(t, fiber.StatusOK, req.Code)
			assert.NotNil(t, data)

			products, ok := data["relatedProducts"].(map[string]string)
			assert.True(t, ok, "expected relatedProducts to be map[string]string")

			// Check expected products are present
			for _, product := range tt.expectedProducts {
				assert.Contains(t, products, product, "expected product %s to be present", product)
			}

			// Check unexpected products are not present
			for _, product := range tt.unexpectedProducts {
				assert.NotContains(t, products, product, "expected product %s to be excluded", product)
			}
		})
	}
}

func TestDownloadService_GetFileName_InvalidProduct(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/stable/invalid-product/versions/all":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`not found`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	mockTemplate := &template.MockTemplateRenderer{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			return "mock script", nil
		},
	}
	injector := buildInjector(mockTemplate, ts.URL)

	logEntry := logrus.NewEntry(logrus.New())
	svc, err := NewDownloadService(injector, logEntry, map[string]interface{}{
		"license_id": "test",
	})
	require.NoError(t, err)

	params := &omnitruck.RequestParams{
		Product: "invalid-product",
		Channel: "stable",
	}

	fileName, req := svc.GetFileName(params)

	assert.False(t, req.Ok, "expected ok false for invalid product")
	assert.Equal(t, fiber.StatusBadRequest, req.Code, "expected 400 bad request for invalid product")
	assert.Empty(t, fileName, "expected empty filename for invalid product")
}

func TestDownloadService_GetScripts(t *testing.T) {
	t.Parallel()

	logEntry := logrus.NewEntry(logrus.New())
	locals := map[string]interface{}{"base_url": "http://example.com"}

	mockTemplate := &template.MockTemplateRenderer{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			if params.Product == "fail" {
				return "", fmt.Errorf("forced script error")
			}
			return "echo 'install chef'", nil
		},
	}

	injector := buildInjector(mockTemplate, "https://omnitruck.chef.io")

	tests := []struct {
		name          string
		params        *omnitruck.RequestParams
		isLinux       bool
		expectSuccess bool
		expectCode    int
		overrideMode  constants.ApiType
	}{
		{
			name: "success linux script",
			params: &omnitruck.RequestParams{
				Product: "chef",
				Channel: "stable",
			},
			isLinux:       true,
			expectSuccess: true,
			expectCode:    fiber.StatusOK,
		},
		{
			name: "success windows script",
			params: &omnitruck.RequestParams{
				Product: "chef",
				Channel: "stable",
			},
			isLinux:       false,
			expectSuccess: true,
			expectCode:    fiber.StatusOK,
		},
		{
			name: "opensource mode with license id",
			params: &omnitruck.RequestParams{
				Product:   "chef",
				Channel:   "stable",
				LicenseId: "test-license",
			},
			isLinux:       true,
			expectSuccess: true,
			expectCode:    fiber.StatusOK,
			overrideMode:  constants.Opensource,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			paramsCopy := *tt.params

			svc, err := NewDownloadService(injector, logEntry, locals)
			require.NoError(t, err)

			if tt.overrideMode != constants.ApiType(0) {
				svc.setMode(tt.overrideMode)
			} else {
				svc.setMode(constants.Commercial)
			}

			var script string
			var req *clients.Request
			if tt.isLinux {
				script, req = svc.GetLinuxScript(&paramsCopy)
			} else {
				script, req = svc.GetWindowsScript(&paramsCopy)
			}

			if tt.expectSuccess {
				assert.True(t, req.Ok)
				assert.Equal(t, tt.expectCode, req.Code)
				assert.Contains(t, script, "install")
			} else {
				assert.False(t, req.Ok)
				assert.Equal(t, tt.expectCode, req.Code)
				assert.Empty(t, script)
			}
		})
	}
}

func TestDownloadService_GetPackageManagers(t *testing.T) {
	t.Parallel()

	logEntry := logrus.NewEntry(logrus.New())
	locals := map[string]interface{}{"base_url": "http://example.com"}

	injector := do.New()

	mockDb := &dboperations.MockIDbOperations{
		GetPackageManagersfunc: func() ([]string, error) {
			return []string{"yum", "apt"}, nil
		},
	}

	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", mockDb)
	do.ProvideNamedValue[template.TemplateRenderer](injector, "templateRenderer", &template.MockTemplateRenderer{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	svc, err := NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	t.Run("success returns package managers", func(t *testing.T) {
		data, req := svc.GetPackageManagers()
		assert.True(t, req.Ok)
		assert.Equal(t, fiber.StatusOK, req.Code)
		assert.NotNil(t, data)
		assert.Contains(t, data, "yum")
	})

	t.Run("error returns failure response", func(t *testing.T) {
		mockDb.GetPackageManagersfunc = func() ([]string, error) {
			return nil, fmt.Errorf("dynamo down")
		}

		data, req := svc.GetPackageManagers()
		assert.False(t, req.Ok)
		assert.Equal(t, fiber.StatusInternalServerError, req.Code)
		assert.Nil(t, data)
		assert.Equal(t, "Error while fetching the information for the product from DB.", req.Message)
	})

}
