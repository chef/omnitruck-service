package services_test

import (
	"fmt"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/dboperations"
	"github.com/chef/omnitruck-service/internal/services"
	"github.com/chef/omnitruck-service/models"
	"github.com/chef/omnitruck-service/utils/template"
	"github.com/gofiber/fiber/v2"
	"github.com/samber/do"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildInjector() *do.Injector {
	injector := do.New()

	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			return "script", nil
		},
	})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{
		LicenseServiceUrl: "http://license-service",
	})
	return injector
}
func TestNewDownloadService_MissingDbService(t *testing.T) {
    injector := do.New()
    do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
    do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
    do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
    do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
    do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

    log := logrus.NewEntry(logrus.New())
    svc, err := services.NewDownloadService(injector, log, map[string]interface{}{})
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
    svc, err := services.NewDownloadService(injector, log, map[string]interface{}{})
    assert.Error(t, err)
    assert.Nil(t, svc)
}

func TestNewDownloadService_MissingReplicated(t *testing.T) {
    injector := do.New()
    do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
    do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
    do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
    do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
    do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

    log := logrus.NewEntry(logrus.New())
    svc, err := services.NewDownloadService(injector, log, map[string]interface{}{})
    assert.Error(t, err)
    assert.Nil(t, svc)
}

func TestNewDownloadService_MissingLicenseClient(t *testing.T) {
    injector := do.New()
    do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
    do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
    do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
    do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
    do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

    log := logrus.NewEntry(logrus.New())
    svc, err := services.NewDownloadService(injector, log, map[string]interface{}{})
    assert.Error(t, err)
    assert.Nil(t, svc)
}

func TestNewDownloadService_MissingMode(t *testing.T) {
    injector := do.New()
    do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
    do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
    do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
    do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
    do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

    log := logrus.NewEntry(logrus.New())
    svc, err := services.NewDownloadService(injector, log, map[string]interface{}{})
    assert.Error(t, err)
    assert.Nil(t, svc)
}

func TestNewDownloadService_MissingConfig(t *testing.T) {
    injector := do.New()
    do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
    do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
    do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
    do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
    do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
    log := logrus.NewEntry(logrus.New())
    svc, err := services.NewDownloadService(injector, log, map[string]interface{}{})
    assert.Error(t, err)
    assert.Nil(t, svc)
}


func TestNewDownloadService_Success(t *testing.T) {
	injector := buildInjector()
	log := logrus.NewEntry(logrus.New())
	svc, err := services.NewDownloadService(injector, log, map[string]interface{}{"license_id": "123"})
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestNewDownloadService_MissingDependency(t *testing.T) {
	injector := do.New()
	log := logrus.NewEntry(logrus.New())
	svc, err := services.NewDownloadService(injector, log, map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, svc)
}

func TestOmnitruck(t *testing.T) {
	injector := buildInjector()
	log := logrus.NewEntry(logrus.New())
	svc, _ := services.NewDownloadService(injector, log, map[string]interface{}{})
	client := svc.Omnitruck()
	assert.NotNil(t, client)
}

func TestPlatformServices(t *testing.T) {
	injector := buildInjector()
	log := logrus.NewEntry(logrus.New())
	svc, _ := services.NewDownloadService(injector, log, map[string]interface{}{})
	plat := svc.PlatformServices()
	assert.NotNil(t, plat)
}

func TestReplicatedService(t *testing.T) {
	log := logrus.NewEntry(logrus.New())
	svc := &services.DownloadService{}
	rep := svc.ReplicatedService(config.ReplicatedConfig{}, log)
	assert.NotNil(t, rep)
}

func TestGetLinuxScript(t *testing.T) {
	injector := buildInjector()
	log := logrus.NewEntry(logrus.New())
	svc, _ := services.NewDownloadService(injector, log, map[string]interface{}{"base_url": "http://x"})
	params := &omnitruck.RequestParams{}
	script, req := svc.GetLinuxScript(params)
	assert.True(t, req.Ok)
	assert.Equal(t, "script", script)
}

func TestGetWindowsScript(t *testing.T) {
	injector := buildInjector()
	log := logrus.NewEntry(logrus.New())
	svc, _ := services.NewDownloadService(injector, log, map[string]interface{}{"base_url": "http://x"})
	params := &omnitruck.RequestParams{}
	script, req := svc.GetWindowsScript(params)
	assert.True(t, req.Ok)
	assert.Equal(t, "script", script)
}

func TestProductDownload_FailsOnEmptyVersions(t *testing.T) {
	injector := buildInjector()
	log := logrus.NewEntry(logrus.New())
	svc, _ := services.NewDownloadService(injector, log, map[string]interface{}{"base_url": "http://x"})
	params := &omnitruck.RequestParams{Product: "chef", Channel: "stable"}
	_, _, _, msg, code, _ := svc.ProductDownload(params, &fiber.Ctx{})
	assert.NotEqual(t, fiber.StatusOK, code)
	assert.NotEmpty(t, msg)
}

func TestDownloadService_Products(t *testing.T) {
	log := logrus.NewEntry(logrus.New())

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

			do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{
				GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
					return "script", nil
				},
			})

			do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
			do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
			do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
			do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

			locals := map[string]interface{}{
				"license_id": "dummy-license",
				"base_url":   "http://example.com",
			}

			svc, err := services.NewDownloadService(injector, log, locals)
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

	logEntry := log.NewEntry(log.New())

	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{
		GetRelatedProductsfunc: func(partitionValue string) (*models.RelatedProducts, error) {
			return &models.RelatedProducts{Products: map[string]string{"test": "test"}}, nil
		},
	})
	do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	locals := map[string]interface{}{"license_id": "123"}
	svc, err := services.NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	data, req := svc.Platforms()

	assert.True(t, req.Ok, "expected request to be OK")
	assert.NotEmpty(t, data, "expected platform list from Omnitruck to be non-empty")
}

func TestDownloadService_Architectures(t *testing.T) {
	t.Parallel()

	logEntry := log.NewEntry(log.New())

	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	locals := map[string]interface{}{"license_id": "123"}
	svc, err := services.NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	data, req := svc.Architectures()

	assert.True(t, req.Ok, "expected request to be OK")
	assert.NotEmpty(t, data, "expected architectures list from Omnitruck to be non-empty")
}
func TestDownloadService_LatestVersion(t *testing.T) {
	t.Parallel()

	logEntry := log.NewEntry(log.New())

	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	locals := map[string]interface{}{"license_id": "123"}

	svc, err := services.NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

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

	logEntry := log.NewEntry(log.New())

	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	locals := map[string]interface{}{"license_id": "123"}

	svc, err := services.NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

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

	logEntry := log.NewEntry(log.New())
	locals := map[string]interface{}{"base_url": "http://example.com"}

	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	svc, err := services.NewDownloadService(injector, logEntry, locals)
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

	logEntry := log.NewEntry(log.New())
	locals := map[string]interface{}{"base_url": "http://example.com"}

	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	svc, err := services.NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	tests := []struct {
		name          string
		params        *omnitruck.RequestParams
		expectSuccess bool
		expectCode    int
	}{
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
			name: "error when GetMetadata fails",
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
			data, req := svc.ProductMetadata(tt.params)

			if tt.expectSuccess {
				assert.True(t, req.Ok, "expected ok true")
				assert.Equal(t, tt.expectCode, req.Code)
				assert.NotEmpty(t, data.Url)
			} else {
				assert.False(t, req.Ok, "expected ok false")
				assert.Equal(t, tt.expectCode, req.Code)
				assert.Empty(t, data.Url)
			}
		})
	}
}

func TestDownloadService_RelatedProducts(t *testing.T) {
	t.Parallel()

	logEntry := log.NewEntry(log.New())
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
	do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	svc, err := services.NewDownloadService(injector, logEntry, locals)
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

func TestDownloadService_GetFileName(t *testing.T) {
	t.Parallel()

	logEntry := log.NewEntry(log.New())
	locals := map[string]interface{}{"base_url": "http://example.com"}

	injector := do.New()
	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	svc, err := services.NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	tests := []struct {
		name          string
		params        *omnitruck.RequestParams
		expectSuccess bool
		expectCode    int
		overrideMode  constants.ApiType 
	}{
		{
			name: "success returns filename",
			params: &omnitruck.RequestParams{
				Product:         "chef",
				Channel:         "stable",
				Platform:        "ubuntu",
				PlatformVersion: "20.04",
				Architecture:    "x86_64",
			},
			expectSuccess: true,
			expectCode:    fiber.StatusOK,
		},
		{
			name: "error on invalid product (GetAllVersions fails)",
			params: &omnitruck.RequestParams{
				Product: "invalid-product",
				Channel: "stable",
			},
			expectSuccess: false,
			expectCode:    fiber.StatusBadRequest,
		},
		{
			name: "error on GetFileName fails with internal error",
			params: &omnitruck.RequestParams{
				Product:         "chef",
				Channel:         "stable",
				Platform:        "m",
				PlatformVersion: "1",
				Architecture:    "x86_64",
			},
			expectSuccess: false,
			expectCode:    fiber.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.overrideMode != constants.ApiType(0) {
				svc.SetMode(tt.overrideMode)
			}

			svc, err := services.NewDownloadService(injector, logEntry, locals)
			require.NoError(t, err)

			fileName, req := svc.GetFileName(tt.params)

			if tt.expectSuccess {
				assert.True(t, req.Ok, "expected ok true")
				assert.Equal(t, tt.expectCode, req.Code)
				assert.NotEmpty(t, fileName)
			} else {
				assert.False(t, req.Ok, "expected ok false")
				assert.Equal(t, tt.expectCode, req.Code)
				assert.Empty(t, fileName)
			}
		})
	}
}

func TestDownloadService_GetScripts(t *testing.T) {
	t.Parallel()

	logEntry := log.NewEntry(log.New())
	locals := map[string]interface{}{"base_url": "http://example.com"}

	injector := do.New()

	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", &dboperations.MockIDbOperations{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	mockTemplate := &template.MockTemplateRennder{
		GetScriptfunc: func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
			if params.Product == "fail" {
				return "", fmt.Errorf("forced script error")
			}
			return "echo 'install chef'", nil
		},
	}
	do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", mockTemplate)

	svc, err := services.NewDownloadService(injector, logEntry, locals)
	require.NoError(t, err)

	tests := []struct {
		name          string
		params        *omnitruck.RequestParams
		isLinux       bool
		expectSuccess bool
		expectCode    int
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
			name: "error generating linux script",
			params: &omnitruck.RequestParams{
				Product: "fail",
				Channel: "stable",
			},
			isLinux:       true,
			expectSuccess: false,
			expectCode:    fiber.StatusInternalServerError,
		},
		{
			name: "error generating windows script",
			params: &omnitruck.RequestParams{
				Product: "fail",
				Channel: "stable",
			},
			isLinux:       false,
			expectSuccess: false,
			expectCode:    fiber.StatusInternalServerError,
		},
		{
			name: "opensource mode removes license id",
			params: &omnitruck.RequestParams{
				Product:   "chef",
				Channel:   "stable",
				LicenseId: "should-be-cleared",
			},
			isLinux:       true,
			expectSuccess: true,
			expectCode:    fiber.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "opensource mode removes license id" {
				svc.SetMode(constants.Opensource)
			} else {
				svc.SetMode(constants.Commercial)
			}

			var script string
			var req *clients.Request
			if tt.isLinux {
				script, req = svc.GetLinuxScript(tt.params)
			} else {
				script, req = svc.GetWindowsScript(tt.params)
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

			if tt.name == "opensource mode removes license id" {
				assert.Empty(t, tt.params.LicenseId, "expected license id to be cleared for opensource")
			}
		})
	}
}

func TestDownloadService_GetPackageManagers(t *testing.T) {
	t.Parallel()

	logEntry := log.NewEntry(log.New())
	locals := map[string]interface{}{"base_url": "http://example.com"}

	injector := do.New()

	mockDb := &dboperations.MockIDbOperations{
		GetPackageManagersfunc: func() ([]string, error) {
			return []string{"yum", "apt"}, nil
		},
	}

	do.ProvideNamedValue[dboperations.IDbOperations](injector, "dbService", mockDb)
	do.ProvideNamedValue[template.TemplateRender](injector, "templateRenderer", &template.MockTemplateRennder{})
	do.ProvideNamedValue[replicated.IReplicated](injector, "replicated", &replicated.MockReplicated{})
	do.ProvideNamedValue[clients.ILicense](injector, "licenseClient", &clients.MockLicense{})
	do.ProvideNamedValue[constants.ApiType](injector, "mode", constants.Commercial)
	do.ProvideNamedValue[config.ServiceConfig](injector, "config", config.ServiceConfig{})

	svc, err := services.NewDownloadService(injector, logEntry, locals)
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
