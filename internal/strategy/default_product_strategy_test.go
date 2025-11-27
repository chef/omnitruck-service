package strategy_test

import (
	"testing"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/internal/strategy"
	"github.com/stretchr/testify/assert"
)

func TestDefaultProductStrategy_GetPackages(t *testing.T) {
	tests := []struct {
		name     string
		ok       bool
		code     int
		message  string
		hasError bool
	}{
		{
			name:     "success",
			ok:       true,
			hasError: false,
		},
		{
			name:     "failure",
			ok:       false,
			code:     500,
			message:  "bad request",
			hasError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &omnitruck.MockOmnitruck{
				ProductPackagesFunc: func(_ *omnitruck.RequestParams) *clients.Request {
					return &clients.Request{Ok: tt.ok, Code: tt.code, Message: tt.message, Body: []byte(`{}`)}
				},
			}
			s := &strategy.DefaultProductStrategy{OmnitruckService: mock}
			pkgs, err := s.GetPackages(&omnitruck.RequestParams{})
			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, pkgs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pkgs)
			}
		})
	}
}

func TestDefaultProductStrategy_Download(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		ok        bool
		code      int
		message   string
		licenseId string
		expected  string
		hasError  bool
	}{
		{
			name:     "success",
			body:     `{"url":"http://example.com","version":"2.0.0"}`,
			ok:       true,
			expected: "http://example.com",
		},
		{
			name:      "success with licenseId",
			body:      `{"url":"http://example.com/package.rpm","version":"2.0.0"}`,
			ok:        true,
			licenseId: "test-license-123",
			expected:  "http://example.com/package.rpm?licenseId=test-license-123",
		},
		{
			name:      "success with empty licenseId",
			body:      `{"url":"http://example.com/file.deb","version":"2.0.0"}`,
			ok:        true,
			licenseId: "",
			expected:  "http://example.com/file.deb",
		},
		{
			name:     "failure",
			body:     ``,
			ok:       false,
			code:     500,
			message:  "failed download",
			hasError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &omnitruck.MockOmnitruck{
				ProductDownloadFunc: func(_ *omnitruck.RequestParams) *clients.Request {
					return &clients.Request{Ok: tt.ok, Code: tt.code, Message: tt.message, Body: []byte(tt.body)}
				},
			}
			s := &strategy.DefaultProductStrategy{OmnitruckService: mock}
			params := &omnitruck.RequestParams{LicenseId: tt.licenseId}
			url, _, _, msg, code, err := s.Download(params)
			assert.Equal(t, tt.expected, url)
			assert.Equal(t, tt.message, msg)
			assert.Equal(t, tt.code, code)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultProductStrategy_GetFileName(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		ok       bool
		code     int
		message  string
		expected string
		hasError bool
	}{
		{
			name:     "success",
			body:     `{"url":"http://example.com/installer.pkg","version":"2.0.0"}`,
			ok:       true,
			expected: "installer.pkg",
		},
		{
			name:     "failure",
			body:     ``,
			ok:       false,
			code:     404,
			message:  "not found",
			hasError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &omnitruck.MockOmnitruck{
				ProductMetadataFunc: func(_ *omnitruck.RequestParams) *clients.Request {
					return &clients.Request{Ok: tt.ok, Code: tt.code, Message: tt.message, Body: []byte(tt.body)}
				},
			}
			s := &strategy.DefaultProductStrategy{OmnitruckService: mock}
			out, err := s.GetFileName(&omnitruck.RequestParams{})
			assert.Equal(t, tt.expected, out)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultProductStrategy_GetMetadata(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		ok         bool
		expected   string
		expectFail bool
	}{
		{
			name:     "success",
			body:     `{"version":"1.2.3"}`,
			ok:       true,
			expected: "1.2.3",
		},
		{
			name:       "failure",
			body:       ``,
			ok:         false,
			expected:   "",
			expectFail: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &omnitruck.MockOmnitruck{
				ProductMetadataFunc: func(_ *omnitruck.RequestParams) *clients.Request {
					return &clients.Request{Ok: tt.ok, Body: []byte(tt.body)}
				},
			}
			s := &strategy.DefaultProductStrategy{OmnitruckService: mock}
			meta, req := s.GetMetadata(&omnitruck.RequestParams{})

			assert.Equal(t, tt.ok, req.Ok)
			assert.Equal(t, tt.expected, meta.Version)
		})
	}
}
func TestDefaultProductStrategy_GetLatestVersion(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse string
		mockOk       bool
		expected     string
		expectedOk   bool
	}{
		{
			name:         "success",
			mockResponse: `"1.2.3"`,
			mockOk:       true,
			expected:     "1.2.3",
			expectedOk:   true,
		},
		{
			name:         "failure",
			mockResponse: "",
			mockOk:       false,
			expected:     "",
			expectedOk:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &omnitruck.MockOmnitruck{
				LatestVersionFunc: func(_ *omnitruck.RequestParams) *clients.Request {
					return &clients.Request{Ok: tt.mockOk, Body: []byte(tt.mockResponse)}
				},
			}
			s := &strategy.DefaultProductStrategy{OmnitruckService: mock}
			out, req := s.GetLatestVersion(&omnitruck.RequestParams{})
			assert.Equal(t, tt.expectedOk, req.Ok)
			assert.Equal(t, tt.expected, string(out))
		})
	}
}

func TestDefaultProductStrategy_GetAllVersions(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		ok       bool
		expected int
	}{
		{
			name:     "success",
			body:     `["1.2.3", "2.0.0"]`,
			ok:       true,
			expected: 2,
		},
		{
			name:     "failure",
			body:     `[]`,
			ok:       false,
			expected: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &omnitruck.MockOmnitruck{
				ProductVersionsFunc: func(_ *omnitruck.RequestParams) *clients.Request {
					return &clients.Request{Ok: tt.ok, Body: []byte(tt.body)}
				},
			}
			s := &strategy.DefaultProductStrategy{OmnitruckService: mock}
			out, req := s.GetAllVersions(&omnitruck.RequestParams{})
			assert.Equal(t, tt.ok, req.Ok)
			assert.Len(t, out, tt.expected)
		})
	}
}

func TestDefaultProductStrategy_UpdatePackages(t *testing.T) {
	s := &strategy.DefaultProductStrategy{}
	params := &omnitruck.RequestParams{}
	list := omnitruck.PackageList{
		"linux": {
			"20.04": {
				"x86_64": omnitruck.PackageMetadata{
					Version: "1.2.3",
				},
			},
		},
	}

	s.UpdatePackages(&list, params, "http://baseurl")
	assert.Contains(t, list["linux"]["20.04"]["x86_64"].Url, "http://baseurl")
}
