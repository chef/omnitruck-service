package strategy

import (
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/constants"
)

type ModeStrategy interface {
	FilterProducts(data omnitruck.ItemList, eol bool) omnitruck.ItemList
	FilterVersions(data []omnitruck.ProductVersion, product string) []omnitruck.ProductVersion
}

type CommercialModeStrategy struct{}

type OpensourceModeStrategy struct{}

type TrialModeStrategy struct{}

func (s *CommercialModeStrategy) FilterProducts(data omnitruck.ItemList, eol bool) omnitruck.ItemList {
	if !eol {
		data = omnitruck.FilterList(data, omnitruck.EolProductName)
	}
	return append(data, constants.PLATFORM_SERVICE_PRODUCT)
}

func (s *CommercialModeStrategy) FilterVersions(data []omnitruck.ProductVersion, product string) []omnitruck.ProductVersion {
	return omnitruck.FilterProductList(data, product, omnitruck.EolProductVersion)
}

func (s *OpensourceModeStrategy) FilterProducts(data omnitruck.ItemList, eol bool) omnitruck.ItemList {
	return omnitruck.SelectList(data, omnitruck.OsProductName)
}

func (s *OpensourceModeStrategy) FilterVersions(data []omnitruck.ProductVersion, product string) []omnitruck.ProductVersion {
	if product != constants.AUTOMATE_PRODUCT {
		return omnitruck.FilterList(data, func(v omnitruck.ProductVersion) bool {
			return !omnitruck.OsProductVersion(product, v)
		})
	}
	return data
}

func (s *TrialModeStrategy) FilterProducts(data omnitruck.ItemList, eol bool) omnitruck.ItemList {
	if !eol {
		data = omnitruck.FilterList(data, omnitruck.EolProductName)
	}
	data = omnitruck.FilterProductsForFreeTrial(data, omnitruck.ProductsForFreeTrial)
	return omnitruck.ProductDisplayName(data)
}

func (s *TrialModeStrategy) FilterVersions(data []omnitruck.ProductVersion, product string) []omnitruck.ProductVersion {
	if len(data) == 0 {
		return data
	}
	return []omnitruck.ProductVersion{data[len(data)-1]}
}

func SelectModeStrategy(mode constants.ApiType) ModeStrategy {
	switch mode {
	case constants.Opensource:
		return &OpensourceModeStrategy{}
	case constants.Trial:
		return &TrialModeStrategy{}
	default:
		return &CommercialModeStrategy{}
	}
}
