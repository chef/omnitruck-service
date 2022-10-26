package filters

import (
	"github.com/coreos/go-semver/semver"
)

type Product struct {
	Name                string
	MinSupportedVersion *semver.Version
	MaxOsVersion        *semver.Version
}

var supportedProducts = map[string]Product{
	"chef": {
		Name:                "chef",
		MinSupportedVersion: semver.New("16.0.0"),
		MaxOsVersion:        semver.New("14.15.6"),
	},
	"chef-backend": {
		Name:                "chef-backend",
		MinSupportedVersion: semver.New("3.0.0"),
	},
	"chef-server": {
		Name:                "chef-server",
		MinSupportedVersion: semver.New("14.0.0"),
		MaxOsVersion:        semver.New("12.19.31"),
	},
	"chef-workstation": {
		Name:                "chef-workstation",
		MinSupportedVersion: semver.New("21.0.0"),
		MaxOsVersion:        semver.New("0.4.2"),
	},
	"inspec": {
		Name:                "inspec",
		MinSupportedVersion: semver.New("4.0.0"),
		MaxOsVersion:        semver.New("4.3.2"),
	},
	"manage": {
		Name:                "manage",
		MinSupportedVersion: semver.New("2.5.0"),
	},
	"supermarket": {
		Name:                "supermarket",
		MinSupportedVersion: semver.New("5.0.0"),
		MaxOsVersion:        semver.New("5.1.44"),
	},
}

func EolProduct(name string) bool {
	_, ok := supportedProducts[name]
	return ok
}

func OsProduct(name string) bool {
	p, ok := supportedProducts[name]
	if ok && p.MaxOsVersion != nil {
		return true
	}

	return false
}
