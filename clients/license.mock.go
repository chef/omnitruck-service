package clients

import (
	"net/http"
)

type MockLicense struct {
	GetFunc                        func(url string) *Request
	ValidateFunc                   func(id, licenseServiceUrl string, data *Response) *Request
	GetReplicatedCustomerEmailFunc func(licenseId, licenseServiceUrl string, data *Response) *Request
	IsTrialFunc                    func(l string) bool
	IsFreeFunc                     func(l string) bool
}

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return &http.Response{}, nil
}

func NewMockLicenseClient() ILicense {
	return &License{
		client: &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{}, nil
			},
		},
	}
}

func (m *MockLicense) GetReplicatedCustomerEmail(licenseId, licenseServiceUrl string, data *Response) *Request {
	return m.GetReplicatedCustomerEmailFunc(licenseId, licenseServiceUrl, data)
}

func (m *MockLicense) Get(url string) *Request {
	return m.GetFunc(url)
}

func (m *MockLicense) Validate(id, licenseServiceUrl string, data *Response) *Request {
	return m.ValidateFunc(id, licenseServiceUrl, data)
}

func (m *MockLicense) IsTrial(l string) bool {
	return m.IsTrialFunc(l)
}

func (m *MockLicense) IsFree(l string) bool {
	return m.IsFreeFunc(l)
}
