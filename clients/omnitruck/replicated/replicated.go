package replicated

import "github.com/chef/omnitruck-service/clients/omnitruck"

type IReplicated interface {
	SearchCustomersByEmail(email string, requestId string) (customers []Customer, err error)
	PlatformVersionsAll(req *omnitruck.RequestParams, serverMode int) ([]omnitruck.ProductVersion, error)
	PlatformVersionLatest(req *omnitruck.RequestParams, serverMode int) (omnitruck.ProductVersion, error)
	PlatformMetadata(req *omnitruck.RequestParams, serverMode int) (omnitruck.PackageMetadata, error)
	PlatformPackages(req *omnitruck.RequestParams, serverMode int) (omnitruck.PackageList, error)
	PlatformFilename(req *omnitruck.RequestParams, serverMode int) (string, error)
}
