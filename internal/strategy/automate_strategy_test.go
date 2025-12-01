package strategy_test

import (
	"errors"
	"testing"

	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/internal/strategy"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestProductDynamoStrategy_GetLatestVersion(t *testing.T) {
	tests := []struct {
		name        string
		mockService func() *omnitruck.MockDynamoServices
		expectOK    bool
		expectVer   string
		expectCode  int
	}{
		{
			name: "successfully gets latest version",
			mockService: func() *omnitruck.MockDynamoServices {
				return &omnitruck.MockDynamoServices{
					VersionLatestFunc: func(p *omnitruck.RequestParams) (omnitruck.ProductVersion, error) {
						return "2.0.0", nil
					},
				}
			},
			expectOK:   true,
			expectVer:  "2.0.0",
			expectCode: 0,
		},
		{
			name: "error fetching latest version",
			mockService: func() *omnitruck.MockDynamoServices {
				return &omnitruck.MockDynamoServices{
					VersionLatestFunc: func(p *omnitruck.RequestParams) (omnitruck.ProductVersion, error) {
						return "", errors.New("db error")
					},
				}
			},
			expectOK:   false,
			expectVer:  "",
			expectCode: 500,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := &strategy.ProductDynamoStrategy{
				DynamoService: tt.mockService(),
				Log:           log.NewEntry(log.New()),
			}
			params := &omnitruck.RequestParams{Product: "automate"}

			version, req := s.GetLatestVersion(params)

			assert.Equal(t, tt.expectOK, req.Ok)
			assert.Equal(t, tt.expectVer, string(version))
			assert.Equal(t, tt.expectCode, req.Code)

			if !req.Ok {
				t.Logf("Failure Request: %+v", req)
			}
		})
	}
}

func TestProductDynamoStrategy_GetAllVersions(t *testing.T) {
	tests := []struct {
		name        string
		mockService func() *omnitruck.MockDynamoServices
		expectOK    bool
		expectLen   int
	}{
		{
			name: "successfully fetches all versions",
			mockService: func() *omnitruck.MockDynamoServices {
				return &omnitruck.MockDynamoServices{
					VersionAllFunc: func(p *omnitruck.RequestParams) ([]omnitruck.ProductVersion, error) {
						return []omnitruck.ProductVersion{"1.0.0", "2.0.0"}, nil
					},
				}
			},
			expectOK:  true,
			expectLen: 2,
		},
		{
			name: "error fetching all versions",
			mockService: func() *omnitruck.MockDynamoServices {
				return &omnitruck.MockDynamoServices{
					VersionAllFunc: func(p *omnitruck.RequestParams) ([]omnitruck.ProductVersion, error) {
						return nil, errors.New("db error")
					},
				}
			},
			expectOK:  false,
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := &strategy.ProductDynamoStrategy{
				DynamoService: tt.mockService(),
				Log:           log.NewEntry(log.New()),
			}

			params := &omnitruck.RequestParams{Product: "automate"}
			versions, req := s.GetAllVersions(params)

			assert.Equal(t, tt.expectOK, req.Ok)
			if tt.expectOK {
				assert.Len(t, versions, tt.expectLen)
			} else {
				assert.Nil(t, versions)
			}
		})
	}
}

func TestProductDynamoStrategy_GetPackages(t *testing.T) {
	mock := &omnitruck.MockDynamoServices{
		ProductPackagesFunc: func(p *omnitruck.RequestParams) (omnitruck.PackageList, error) {
			return omnitruck.PackageList{}, nil
		},
	}
	s := &strategy.ProductDynamoStrategy{
		DynamoService: mock,
		Log:           log.NewEntry(log.New()),
	}
	params := &omnitruck.RequestParams{Product: "automate"}
	data, err := s.GetPackages(params)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

func TestProductDynamoStrategy_GetMetadata(t *testing.T) {
	tests := []struct {
		name        string
		mockService func() *omnitruck.MockDynamoServices
		expectOK    bool
		expectVer   string
	}{
		{
			name: "returns metadata successfully",
			mockService: func() *omnitruck.MockDynamoServices {
				return &omnitruck.MockDynamoServices{
					ProductMetadataFunc: func(p *omnitruck.RequestParams) (omnitruck.PackageMetadata, error) {
						return omnitruck.PackageMetadata{Version: "2.0.0"}, nil
					},
				}
			},
			expectOK:  true,
			expectVer: "2.0.0",
		},
		{
			name: "error retrieving metadata",
			mockService: func() *omnitruck.MockDynamoServices {
				return &omnitruck.MockDynamoServices{
					ProductMetadataFunc: func(p *omnitruck.RequestParams) (omnitruck.PackageMetadata, error) {
						return omnitruck.PackageMetadata{}, errors.New("db error")
					},
				}
			},
			expectOK:  false,
			expectVer: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s := &strategy.ProductDynamoStrategy{
				DynamoService: tt.mockService(),
				Log:           log.NewEntry(log.New()),
			}

			params := &omnitruck.RequestParams{Product: "automate"}
			meta, req := s.GetMetadata(params)

			assert.Equal(t, tt.expectOK, req.Ok)
			assert.Equal(t, tt.expectVer, meta.Version)
			if !tt.expectOK {
				assert.Equal(t, 500, req.Code)
			}
		})
	}
}

func TestProductDynamoStrategy_Download(t *testing.T) {
	tests := []struct {
		name        string
		params      *omnitruck.RequestParams
		mockURL     string
		expectedURL string
	}{
		{
			name:        "without licenseId",
			params:      &omnitruck.RequestParams{Product: "automate"},
			mockURL:     "http://example.com",
			expectedURL: "http://example.com",
		},
		{
			name:        "with licenseId",
			params:      &omnitruck.RequestParams{Product: "automate", LicenseId: "test-license-123"},
			mockURL:     "http://example.com/package.rpm",
			expectedURL: "http://example.com/package.rpm?licenseId=test-license-123",
		},
		{
			name:        "with empty licenseId",
			params:      &omnitruck.RequestParams{Product: "habitat", LicenseId: ""},
			mockURL:     "http://example.com/hab.tar.gz",
			expectedURL: "http://example.com/hab.tar.gz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &omnitruck.MockDynamoServices{
				ProductDownloadFunc: func(p *omnitruck.RequestParams) (string, error) {
					return tt.mockURL, nil
				},
			}
			s := &strategy.ProductDynamoStrategy{
				DynamoService: mock,
				Log:           log.NewEntry(log.New()),
			}
			url, rc, hdr, msg, code, err := s.Download(tt.params)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedURL, url)
			assert.Nil(t, rc)
			assert.Nil(t, hdr)
			assert.Empty(t, msg)
			assert.Equal(t, 0, code)
		})
	}
}

func TestProductDynamoStrategy_GetFileName(t *testing.T) {
	mock := &omnitruck.MockDynamoServices{
		GetFilenameFunc: func(p *omnitruck.RequestParams) (string, error) {
			return "test.deb", nil
		},
	}
	s := &strategy.ProductDynamoStrategy{
		DynamoService: mock,
		Log:           log.NewEntry(log.New()),
	}
	params := &omnitruck.RequestParams{Product: "automate"}
	name, err := s.GetFileName(params)
	assert.NoError(t, err)
	assert.Equal(t, "test.deb", name)
}

func TestProductDynamoStrategy_UpdatePackages(t *testing.T) {
	s := &strategy.ProductDynamoStrategy{}
	list := omnitruck.PackageList{
		"linux": {
			"20.04": {
				"x86_64": omnitruck.PackageMetadata{Version: "2.0.0"},
			},
		},
	}
	params := &omnitruck.RequestParams{
		Product:      "automate",
		Channel:      "stable",
		Architecture: "x86_64",
	}
	s.UpdatePackages(&list, params, "http://download")
	pkg := list["linux"]["20.04"]["x86_64"]
	assert.Contains(t, pkg.Url, "http://download")
}
