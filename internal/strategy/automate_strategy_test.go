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
	mock := &omnitruck.MockDynamoServices{
		VersionLatestFunc: func(p *omnitruck.RequestParams) (omnitruck.ProductVersion, error) {
			return "2.0.0", nil
		},
	}

	s := &strategy.ProductDynamoStrategy{
		DynamoService: mock,
		Log:           log.NewEntry(log.New()),
	}

	params := &omnitruck.RequestParams{Product: "automate"}
	version, req := s.GetLatestVersion(params)
	assert.True(t, req.Ok)
	assert.Equal(t, "2.0.0", string(version))
}

func TestProductDynamoStrategy_GetLatestVersion_Error(t *testing.T) {
	mock := &omnitruck.MockDynamoServices{
		VersionLatestFunc: func(p *omnitruck.RequestParams) (omnitruck.ProductVersion, error) {
			return "", errors.New("db error")
		},
	}

	s := &strategy.ProductDynamoStrategy{
		DynamoService: mock,
		Log:           log.NewEntry(log.New()),
	}

	params := &omnitruck.RequestParams{Product: "automate"}
	_, req := s.GetLatestVersion(params)

	assert.False(t, req.Ok)
	assert.Equal(t, 500, req.Code)
	assert.Empty(t, req.Message)
	t.Logf("Request: %+v", req)
}

func TestProductDynamoStrategy_GetAllVersions(t *testing.T) {
	mock := &omnitruck.MockDynamoServices{
		VersionAllFunc: func(p *omnitruck.RequestParams) ([]omnitruck.ProductVersion, error) {
			return []omnitruck.ProductVersion{"1.0.0", "2.0.0"}, nil
		},
	}

	s := &strategy.ProductDynamoStrategy{
		DynamoService: mock,
		Log:           log.NewEntry(log.New()),
	}

	params := &omnitruck.RequestParams{Product: "automate"}
	data, req := s.GetAllVersions(params)
	assert.True(t, req.Ok)
	assert.Equal(t, 2, len(data))
}

func TestProductDynamoStrategy_GetAllVersions_Error(t *testing.T) {
	mock := &omnitruck.MockDynamoServices{
		VersionAllFunc: func(p *omnitruck.RequestParams) ([]omnitruck.ProductVersion, error) {
			return nil, errors.New("db error")
		},
	}

	s := &strategy.ProductDynamoStrategy{
		DynamoService: mock,
		Log:           log.NewEntry(log.New()),
	}

	params := &omnitruck.RequestParams{Product: "automate"}
	data, req := s.GetAllVersions(params)
	assert.False(t, req.Ok)
	assert.Nil(t, data)
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
	mock := &omnitruck.MockDynamoServices{
		ProductMetadataFunc: func(p *omnitruck.RequestParams) (omnitruck.PackageMetadata, error) {
			return omnitruck.PackageMetadata{Version: "2.0.0"}, nil
		},
	}

	s := &strategy.ProductDynamoStrategy{
		DynamoService: mock,
		Log:           log.NewEntry(log.New()),
	}
	params := &omnitruck.RequestParams{Product: "automate"}
	data, req := s.GetMetadata(params)
	assert.True(t, req.Ok)
	assert.Equal(t, "2.0.0", data.Version)
}

func TestProductDynamoStrategy_GetMetadata_Error(t *testing.T) {
	mock := &omnitruck.MockDynamoServices{
		ProductMetadataFunc: func(p *omnitruck.RequestParams) (omnitruck.PackageMetadata, error) {
			return omnitruck.PackageMetadata{}, errors.New("db error")
		},
	}

	s := &strategy.ProductDynamoStrategy{
		DynamoService: mock,
		Log:           log.NewEntry(log.New()),
	}

	params := &omnitruck.RequestParams{Product: "automate"}
	_, req := s.GetMetadata(params)

	assert.False(t, req.Ok)
	assert.Equal(t, 500, req.Code)
	assert.Empty(t, req.Message)
	t.Logf("Request: %+v", req)
}

func TestProductDynamoStrategy_Download(t *testing.T) {
	mock := &omnitruck.MockDynamoServices{
		ProductDownloadFunc: func(p *omnitruck.RequestParams) (string, error) {
			return "http://example.com", nil
		},
	}
	s := &strategy.ProductDynamoStrategy{
		DynamoService: mock,
		Log:           log.NewEntry(log.New()),
	}
	params := &omnitruck.RequestParams{Product: "automate"}
	url, rc, hdr, msg, code, err := s.Download(params)
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com", url)
	assert.Nil(t, rc)
	assert.Nil(t, hdr)
	assert.Empty(t, msg)
	assert.Equal(t, 0, code)
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
