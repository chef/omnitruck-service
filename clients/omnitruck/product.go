package omnitruck

import (
	version "github.com/hashicorp/go-version"
)

type Product struct {
	Name              string
	ProductName       string
	SupportedVersion  version.Constraints
	OpensourceVersion version.Constraints
}

func NewConstraint(i string) version.Constraints {
	c, _ := version.NewConstraint(i)
	return c
}

var supportedProducts = map[string]Product{
	"automate": {
		Name:              "automate",
		ProductName:       "Chef Automate",
		SupportedVersion:  NewConstraint(">= 0"),
		OpensourceVersion: NewConstraint(">= 0"),
	},
	"chef": {
		Name:              "chef",
		ProductName:       "Chef Infra Client",
		SupportedVersion:  NewConstraint(">= 16.0.0"),
		OpensourceVersion: NewConstraint("<= 14.15.6"),
	},
	"chef-backend": {
		Name:             "chef-backend",
		ProductName:      "Chef Backend",
		SupportedVersion: NewConstraint(">= 3.0.0"),
	},
	"chef-server": {
		Name:              "chef-server",
		ProductName:       "Chef Infra Server",
		SupportedVersion:  NewConstraint(">= 14.0.0"),
		OpensourceVersion: NewConstraint("<= 12.19.31"),
	},
	"chef-workstation": {
		Name:              "chef-workstation",
		ProductName:       "Chef Workstation",
		SupportedVersion:  NewConstraint(">= 21.0.0"),
		OpensourceVersion: NewConstraint("<= 0.4.2"),
	},
	"habitat": {
		Name:              "habitat",
		ProductName:       "Chef Habitat",
		SupportedVersion:  NewConstraint(">= 0"),
		OpensourceVersion: NewConstraint("< 0.79.0"),
	},
	"inspec": {
		Name:              "inspec",
		ProductName:       "InSpec",
		SupportedVersion:  NewConstraint(">= 4.0.0"),
		OpensourceVersion: NewConstraint("<= 4.3.2"),
	},
	"manage": {
		Name:             "manage",
		ProductName:      "Chef Manage",
		SupportedVersion: NewConstraint(">= 2.5.0"),
	},
	"supermarket": {
		Name:              "supermarket",
		ProductName:       "Chef Supermarket",
		SupportedVersion:  NewConstraint(">= 5.0.0"),
		OpensourceVersion: NewConstraint("<= 5.1.63"),
	},
	"desktop": {
		Name:              "desktop",
		ProductName:       "",
		SupportedVersion:  NewConstraint(">= 0"),
		OpensourceVersion: NewConstraint("<= 14.15.6"),
	},
	"chef-ice": {
		Name:             "chef-ice",
		ProductName:      "Chef Infra Client Enterprise",
		SupportedVersion: NewConstraint(">= 0"),
	},
	"migration-tool": {
		Name:             "migration-tool",
		ProductName:      "Migration Tool",
		SupportedVersion: NewConstraint(">= 0"),
	},
}

func SupportedVersion(product string) string {
	p, ok := supportedProducts[product]
	if ok {
		return p.SupportedVersion.String()
	}
	return ""
}

func EolProductName(name string) bool {
	p, ok := supportedProducts[name]
	return !ok || p.SupportedVersion == nil
}

func EolProductVersion(product string, v ProductVersion) bool {
	// Latest should never be EOL
	if v == "latest" || len(v) == 0 {
		return false
	}
	// If we can't find the product in our list then it's no EOL
	p, ok := supportedProducts[product]
	if !ok {
		return false
	}

	v1, err := version.NewVersion(string(v))
	if err != nil {
		return false
	}

	if p.SupportedVersion != nil {
		return !p.SupportedVersion.Check(v1)
	}

	return false
}

func OsProductName(name string) bool {
	p, ok := supportedProducts[name]
	return ok && p.OpensourceVersion != nil
}

func OsProductVersion(name string, v ProductVersion) bool {
	// If we can't find it in our list then it's not Opensource
	p, ok := supportedProducts[name]
	if !ok || p.OpensourceVersion == nil {
		return false
	}

	v1, _ := version.NewVersion(string(v))
	return p.OpensourceVersion.Check(v1.Core())
}

func ProductDisplayName(data ItemList) ItemList {
	for i, val := range data {
		p, ok := supportedProducts[val]
		if !ok {
			data[i] = val
		}
		data[i] = val + ":" + p.ProductName
	}
	return data
}

func ProductsForFreeTrial(name string) bool {
	if name == "supermarket" || name == "manage" || name == "chef-backend" {
		return true
	}
	return false
}
