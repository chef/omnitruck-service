package strategy_test

import (
	"testing"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/internal/strategy"
	"github.com/stretchr/testify/assert"
)

func TestModeStrategies_FilterProducts(t *testing.T) {
	tests := []struct {
		name       string
		strategy   strategy.ModeStrategy
		assertFunc func(t *testing.T, result omnitruck.ItemList)
	}{
		{
			name:     "Commercial mode includes platform-service",
			strategy: &strategy.CommercialModeStrategy{},
			assertFunc: func(t *testing.T, result omnitruck.ItemList) {
				assert.Contains(t, result, constants.PLATFORM_SERVICE_PRODUCT)
			},
		},
		{
			name:     "Opensource mode returns non-nil list",
			strategy: &strategy.OpensourceModeStrategy{},
			assertFunc: func(t *testing.T, result omnitruck.ItemList) {
				assert.NotNil(t, result)
			},
		},
		{
			name:     "Trial mode returns non-nil list",
			strategy: &strategy.TrialModeStrategy{},
			assertFunc: func(t *testing.T, result omnitruck.ItemList) {
				assert.NotNil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := omnitruck.ItemList{"chef", "automate"}
			result := tt.strategy.FilterProducts(input, false)
			tt.assertFunc(t, result)
		})
	}
}
func TestModeStrategies_FilterVersions(t *testing.T) {
	tests := []struct {
		name       string
		strategy   strategy.ModeStrategy
		versions   []omnitruck.ProductVersion
		product    string
		eol        string
		assertFunc func(t *testing.T, result []omnitruck.ProductVersion)
	}{
		{
			name:     "Commercial mode returns non-nil version list",
			strategy: &strategy.CommercialModeStrategy{},
			versions: []omnitruck.ProductVersion{"1.2.3", "17.0.0"},
			product:  "chef",
			eol:      "false",
			assertFunc: func(t *testing.T, result []omnitruck.ProductVersion) {
				assert.Equal(t, []omnitruck.ProductVersion{"17.0.0"}, result)
			},
		},
		{
			name:     "Commercial mode returns non-nil version list",
			strategy: &strategy.CommercialModeStrategy{},
			versions: []omnitruck.ProductVersion{"1.2.3", "17.0.0"},
			product:  "chef",
			eol:      "true",
			assertFunc: func(t *testing.T, result []omnitruck.ProductVersion) {
				assert.Equal(t, []omnitruck.ProductVersion{"1.2.3", "17.0.0"}, result)
			},
		},
		{
			name:     "Opensource mode returns all versions for 'chef'",
			strategy: &strategy.OpensourceModeStrategy{},
			versions: []omnitruck.ProductVersion{"14.2.3", "17.0.0"},
			product:  "chef",
			eol:      "false",
			assertFunc: func(t *testing.T, result []omnitruck.ProductVersion) {
				assert.Equal(t, []omnitruck.ProductVersion{"14.2.3"}, result)
			},
		},
		{
			name:     "Opensource mode returns all versions for 'automate'",
			strategy: &strategy.OpensourceModeStrategy{},
			versions: []omnitruck.ProductVersion{"1.2.3", "2.0.0"},
			product:  constants.AUTOMATE_PRODUCT,
			eol:      "false",
			assertFunc: func(t *testing.T, result []omnitruck.ProductVersion) {
				assert.Equal(t, []omnitruck.ProductVersion{"1.2.3", "2.0.0"}, result)
			},
		},
		{
			name:     "Trial mode returns only the latest version",
			strategy: &strategy.TrialModeStrategy{},
			versions: []omnitruck.ProductVersion{"1.0.0", "2.0.0", "3.0.0"},
			product:  "chef",
			eol:      "false",
			assertFunc: func(t *testing.T, result []omnitruck.ProductVersion) {
				assert.Equal(t, []omnitruck.ProductVersion{"3.0.0"}, result)
			},
		},
		{
			name:     "Trial mode with empty input returns empty list",
			strategy: &strategy.TrialModeStrategy{},
			versions: []omnitruck.ProductVersion{},
			product:  "chef",
			eol:      "false",
			assertFunc: func(t *testing.T, result []omnitruck.ProductVersion) {
				assert.Empty(t, result)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			result := tt.strategy.FilterVersions(tt.versions, tt.product, tt.eol)
			tt.assertFunc(t, result)
		})
	}
}
func TestSelectModeStrategy(t *testing.T) {
	assert.IsType(t, &strategy.OpensourceModeStrategy{}, strategy.SelectModeStrategy(constants.Opensource))
	assert.IsType(t, &strategy.TrialModeStrategy{}, strategy.SelectModeStrategy(constants.Trial))
	assert.IsType(t, &strategy.CommercialModeStrategy{}, strategy.SelectModeStrategy(constants.Commercial))
}
