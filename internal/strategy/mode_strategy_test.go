package strategy_test

import (
	"testing"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/internal/strategy"
	"github.com/stretchr/testify/assert"
)

func TestCommercialModeStrategy_FilterProducts(t *testing.T) {
	commercial := &strategy.CommercialModeStrategy{}
	input := omnitruck.ItemList{"chef", "automate"}
	result := commercial.FilterProducts(input, false)
	assert.Contains(t, result, constants.PLATFORM_SERVICE_PRODUCT)
}

func TestCommercialModeStrategy_FilterVersions(t *testing.T) {
	commercial := &strategy.CommercialModeStrategy{}
	versions := []omnitruck.ProductVersion{"1.2.3", "2.0.0"}
	result := commercial.FilterVersions(versions, "chef")
	assert.NotNil(t, result)
}

func TestOpensourceModeStrategy_FilterProducts(t *testing.T) {
	open := &strategy.OpensourceModeStrategy{}
	input := omnitruck.ItemList{"chef", "automate"}
	result := open.FilterProducts(input, false)
	assert.NotNil(t, result)
}

func TestOpensourceModeStrategy_FilterVersions(t *testing.T) {
	open := &strategy.OpensourceModeStrategy{}
	versions := []omnitruck.ProductVersion{"1.2.3", "2.0.0"}
	result := open.FilterVersions(versions, "chef")
	assert.NotNil(t, result)

	automate := open.FilterVersions(versions, constants.AUTOMATE_PRODUCT)
	assert.Equal(t, versions, automate)
}

func TestTrialModeStrategy_FilterProducts(t *testing.T) {
	trial := &strategy.TrialModeStrategy{}
	input := omnitruck.ItemList{"chef", "automate"}
	result := trial.FilterProducts(input, false)
	assert.NotNil(t, result)
}

func TestTrialModeStrategy_FilterVersions(t *testing.T) {
	trial := &strategy.TrialModeStrategy{}
	versions := []omnitruck.ProductVersion{"1.0.0", "2.0.0", "3.0.0"}
	result := trial.FilterVersions(versions, "chef")
	assert.Equal(t, []omnitruck.ProductVersion{"3.0.0"}, result)

	empty := trial.FilterVersions([]omnitruck.ProductVersion{}, "chef")
	assert.Empty(t, empty)
}

func TestSelectModeStrategy(t *testing.T) {
	assert.IsType(t, &strategy.OpensourceModeStrategy{}, strategy.SelectModeStrategy(constants.Opensource))
	assert.IsType(t, &strategy.TrialModeStrategy{}, strategy.SelectModeStrategy(constants.Trial))
	assert.IsType(t, &strategy.CommercialModeStrategy{}, strategy.SelectModeStrategy(constants.Commercial))
}
