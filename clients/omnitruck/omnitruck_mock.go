package omnitruck
import "github.com/chef/omnitruck-service/clients"

type MockOmnitruck struct {
	LatestVersionFunc   func(params *RequestParams) *clients.Request
	ProductVersionsFunc func(params *RequestParams) *clients.Request
	ProductPackagesFunc func(params *RequestParams) *clients.Request
	ProductMetadataFunc func(params *RequestParams) *clients.Request
	ProductDownloadFunc func(params *RequestParams) *clients.Request
}

func (m *MockOmnitruck) LatestVersion(params *RequestParams) *clients.Request {
	if m.LatestVersionFunc != nil {
		return m.LatestVersionFunc(params)
	}
	return nil
}
func (m *MockOmnitruck) ProductVersions(params *RequestParams) *clients.Request {
	if m.ProductVersionsFunc != nil {
		return m.ProductVersionsFunc(params)
	}
	return nil
}
func (m *MockOmnitruck) ProductPackages(params *RequestParams) *clients.Request {
	if m.ProductPackagesFunc != nil {
		return m.ProductPackagesFunc(params)
	}
	return nil
}
func (m *MockOmnitruck) ProductMetadata(params *RequestParams) *clients.Request {
	if m.ProductMetadataFunc != nil {
		return m.ProductMetadataFunc(params)
	}
	return nil
}
func (m *MockOmnitruck) ProductDownload(params *RequestParams) *clients.Request {
	if m.ProductDownloadFunc != nil {
		return m.ProductDownloadFunc(params)
	}
	return nil
}