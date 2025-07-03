package strategy_test

import (
	"testing"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/internal/strategy"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockOmnitruck struct {
	LatestVersionFunc   func(params *omnitruck.RequestParams) *clients.Request
	ProductVersionsFunc func(params *omnitruck.RequestParams) *clients.Request
	ProductPackagesFunc func(params *omnitruck.RequestParams) *clients.Request
	ProductMetadataFunc func(params *omnitruck.RequestParams) *clients.Request
	ProductDownloadFunc func(params *omnitruck.RequestParams) *clients.Request
}

func (m *mockOmnitruck) LatestVersion(params *omnitruck.RequestParams) *clients.Request {
	return m.LatestVersionFunc(params)
}
func (m *mockOmnitruck) ProductVersions(params *omnitruck.RequestParams) *clients.Request {
	return m.ProductVersionsFunc(params)
}
func (m *mockOmnitruck) ProductPackages(params *omnitruck.RequestParams) *clients.Request {
	return m.ProductPackagesFunc(params)
}
func (m *mockOmnitruck) ProductMetadata(params *omnitruck.RequestParams) *clients.Request {
	return m.ProductMetadataFunc(params)
}
func (m *mockOmnitruck) ProductDownload(params *omnitruck.RequestParams) *clients.Request {
	return m.ProductDownloadFunc(params)
}

func TestDefaultProductStrategy(t *testing.T) {
	params := &omnitruck.RequestParams{
		Product: "chef",
		Channel: "stable",
	}

	t.Run("happy paths", func(t *testing.T) {
		mockOmni := &mockOmnitruck{
			LatestVersionFunc: func(params *omnitruck.RequestParams) *clients.Request {
				return &clients.Request{Ok: true, Body: []byte(`"1.2.3"`)}
			},
			ProductVersionsFunc: func(params *omnitruck.RequestParams) *clients.Request {
				return &clients.Request{Ok: true, Body: []byte(`["1.2.3","2.0.0"]`)}
			},
			ProductPackagesFunc: func(params *omnitruck.RequestParams) *clients.Request {
				return &clients.Request{Ok: true, Body: []byte(`{}`)}
			},
			ProductMetadataFunc: func(params *omnitruck.RequestParams) *clients.Request {
				return &clients.Request{Ok: true, Body: []byte(`{"url":"http://example.com","version":"1.2.3"}`)}
			},
			ProductDownloadFunc: func(params *omnitruck.RequestParams) *clients.Request {
				return &clients.Request{Ok: true, Body: []byte(`{"url":"http://example.com","version":"1.2.3"}`)}
			},
		}

		strat := &strategy.DefaultProductStrategy{
			OmnitruckService: mockOmni,
			Log:              log.NewEntry(log.New()),
		}

		version, req := strat.GetLatestVersion(params)
		assert.True(t, req.Ok)
		assert.Equal(t, "1.2.3", string(version))

		versions, req := strat.GetAllVersions(params)
		assert.True(t, req.Ok)
		assert.Len(t, versions, 2)

		pkgs, err := strat.GetPackages(params)
		assert.NoError(t, err)
		assert.NotNil(t, pkgs)

		meta, req := strat.GetMetadata(params)
		assert.True(t, req.Ok)
		assert.Equal(t, "1.2.3", meta.Version)

		url, rc, hdr, msg, code, err := strat.Download(params)
		assert.Equal(t, "http://example.com", url)
		assert.Nil(t, rc)
		assert.Nil(t, hdr)
		assert.Empty(t, msg)
		assert.Equal(t, 0, code)
		assert.NoError(t, err)

		fileName, err := strat.GetFileName(params)
		assert.NoError(t, err)
		assert.Equal(t, "example.com", fileName)

		list := omnitruck.PackageList{
			"linux": {
				"20.04": {
					"x86_64": omnitruck.PackageMetadata{
						Version: "1.2.3",
					},
				},
			},
		}
		strat.UpdatePackages(&list, params, "http://baseurl")
		assert.Contains(t, list["linux"]["20.04"]["x86_64"].Url, "http://baseurl")
	})

	t.Run("error paths", func(t *testing.T) {
		mockOmni := &mockOmnitruck{
			ProductPackagesFunc: func(params *omnitruck.RequestParams) *clients.Request {
				return &clients.Request{Ok: false, Code: 500, Message: "bad packages"}
			},
			ProductMetadataFunc: func(params *omnitruck.RequestParams) *clients.Request {
				return &clients.Request{Ok: false, Code: 500, Message: "bad metadata"}
			},
			ProductDownloadFunc: func(params *omnitruck.RequestParams) *clients.Request {
				return &clients.Request{Ok: false, Code: 500, Message: "bad download"}
			},
		}
		strat := &strategy.DefaultProductStrategy{
			OmnitruckService: mockOmni,
			Log:              log.NewEntry(log.New()),
		}

		_, err := strat.GetPackages(params)
		assert.Error(t, err)

		_, req := strat.GetMetadata(params)
		assert.False(t, req.Ok)

		url, rc, hdr, msg, code, err := strat.Download(params)
		assert.Empty(t, url)
		assert.Nil(t, rc)
		assert.Nil(t, hdr)
		assert.Equal(t, "bad download", msg)
		assert.Equal(t, 500, code)
		assert.Error(t, err)

		_, err = strat.GetFileName(params)
		assert.Error(t, err)
	})
}
