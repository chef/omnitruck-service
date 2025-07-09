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
