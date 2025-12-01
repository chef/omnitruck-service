package strategy_test

import (
	"reflect"
	"testing"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/internal/strategy"
	"github.com/chef/omnitruck-service/models"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSelectProductStrategy(t *testing.T) {
	mockDynamo := &omnitruck.MockDynamoServices{}
	mockPlatform := &omnitruck.PlatformServices{}
	mockOmnitruck := &omnitruck.Omnitruck{}
	mockLog := log.NewEntry(log.New())

	deps := &strategy.ProductStrategyDeps{
		DynamoService:     mockDynamo,
		PlatformService:   mockPlatform,
		OmnitruckService:  mockOmnitruck,
		Log:               mockLog,
		Replicated:        nil,
		LicenseClient:     nil,
		LicenseServiceUrl: "http://mock-license",
		Mode:              constants.Opensource,
		Config: config.ServiceConfig{
			MetadataDetailsTable:       "mock_metadata_table",
			PackageDetailsCurrentTable: "mock_current_table",
			PackageDetailsStableTable:  "mock_stable_table",
			AWSConfig:                  config.AWSConfig{},
			SupportInfra19:             true,
		},
	}

	tests := []struct {
		name         string
		product      string
		channel      string
		expectedType reflect.Type
		expectDbInfo *struct {
			table string
			model reflect.Type
		}
	}{
		{
			name:         "automate product",
			product:      constants.AUTOMATE_PRODUCT,
			channel:      "",
			expectedType: reflect.TypeOf(&strategy.ProductDynamoStrategy{}),
			expectDbInfo: &struct {
				table string
				model reflect.Type
			}{
				table: deps.Config.MetadataDetailsTable,
				model: reflect.TypeOf(models.ProductDetails{}),
			},
		},
		{
			name:         "platform service product",
			product:      constants.PLATFORM_SERVICE_PRODUCT,
			channel:      "",
			expectedType: reflect.TypeOf(&strategy.PlatformServiceStrategy{}),
		},
		{
			name:         "infra product current channel",
			product:      constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT,
			channel:      constants.CURRENT_CHANNEL,
			expectedType: reflect.TypeOf(&strategy.InfraProductStrategy{}),
			expectDbInfo: &struct {
				table string
				model reflect.Type
			}{
				table: deps.Config.PackageDetailsCurrentTable,
				model: reflect.TypeOf(models.PackageDetails{}),
			},
		},
		{
			name:         "infra product stable channel",
			product:      constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT,
			channel:      constants.STABLE_CHANNEL,
			expectedType: reflect.TypeOf(&strategy.InfraProductStrategy{}),
			expectDbInfo: &struct {
				table string
				model reflect.Type
			}{
				table: deps.Config.PackageDetailsStableTable,
				model: reflect.TypeOf(models.PackageDetails{}),
			},
		},
		{
			name:         "default product",
			product:      "something-else",
			channel:      "",
			expectedType: reflect.TypeOf(&strategy.DefaultProductStrategy{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strat := strategy.SelectProductStrategy(tt.product, tt.channel, deps)
			assert.Equal(t, tt.expectedType, reflect.TypeOf(strat))

			if tt.expectDbInfo != nil {
				assert.NotEmpty(t, mockDynamo.SetDbInfoCalledWith, "expected SetDbInfo to be called")
				lastCall := mockDynamo.SetDbInfoCalledWith[len(mockDynamo.SetDbInfoCalledWith)-1]
				assert.Equal(t, tt.expectDbInfo.table, lastCall.Table)
				assert.Equal(t, tt.expectDbInfo.model, lastCall.Model)
			}
		})
	}
}

func TestSelectProductStrategy_SupportInfra19(t *testing.T) {
	mockDynamo := &omnitruck.MockDynamoServices{}
	mockPlatform := &omnitruck.PlatformServices{}
	mockOmnitruck := &omnitruck.Omnitruck{}
	mockLog := log.NewEntry(log.New())

	tests := []struct {
		name           string
		product        string
		channel        string
		supportInfra19 bool
		expectedType   reflect.Type
		expectDbInfo   bool
	}{
		{
			name:           "infra product with SupportInfra19=true should return InfraProductStrategy",
			product:        constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT,
			channel:        constants.CURRENT_CHANNEL,
			supportInfra19: true,
			expectedType:   reflect.TypeOf(&strategy.InfraProductStrategy{}),
			expectDbInfo:   true,
		},
		{
			name:           "infra product with SupportInfra19=false should return DefaultProductStrategy",
			product:        constants.CHEF_INFRA_CLIENT_ENTERPRISE_PRODUCT,
			channel:        constants.CURRENT_CHANNEL,
			supportInfra19: false,
			expectedType:   reflect.TypeOf(&strategy.DefaultProductStrategy{}),
			expectDbInfo:   false,
		},
		{
			name:           "migrate ice with SupportInfra19=true should return InfraProductStrategy",
			product:        constants.MIGRATE_ICE,
			channel:        constants.STABLE_CHANNEL,
			supportInfra19: true,
			expectedType:   reflect.TypeOf(&strategy.InfraProductStrategy{}),
			expectDbInfo:   true,
		},
		{
			name:           "migrate ice with SupportInfra19=false should return DefaultProductStrategy",
			product:        constants.MIGRATE_ICE,
			channel:        constants.STABLE_CHANNEL,
			supportInfra19: false,
			expectedType:   reflect.TypeOf(&strategy.DefaultProductStrategy{}),
			expectDbInfo:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			mockDynamo.SetDbInfoCalledWith = nil

			deps := &strategy.ProductStrategyDeps{
				DynamoService:     mockDynamo,
				PlatformService:   mockPlatform,
				OmnitruckService:  mockOmnitruck,
				Log:               mockLog,
				Replicated:        nil,
				LicenseClient:     nil,
				LicenseServiceUrl: "http://mock-license",
				Mode:              constants.Opensource,
				Config: config.ServiceConfig{
					MetadataDetailsTable:       "mock_metadata_table",
					PackageDetailsCurrentTable: "mock_current_table",
					PackageDetailsStableTable:  "mock_stable_table",
					AWSConfig:                  config.AWSConfig{},
					SupportInfra19:             tt.supportInfra19,
				},
			}

			strat := strategy.SelectProductStrategy(tt.product, tt.channel, deps)
			assert.Equal(t, tt.expectedType, reflect.TypeOf(strat))

			if tt.expectDbInfo {
				assert.NotEmpty(t, mockDynamo.SetDbInfoCalledWith, "expected SetDbInfo to be called when SupportInfra19=true")
			} else {
				assert.Empty(t, mockDynamo.SetDbInfoCalledWith, "expected SetDbInfo NOT to be called when SupportInfra19=false")
			}
		})
	}
}
