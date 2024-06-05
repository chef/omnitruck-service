package replicated_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/logger"
	"github.com/stretchr/testify/assert"
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (mc MockClient) Do(req *http.Request) (*http.Response, error) {
	return mc.DoFunc(req)
}

func TestSearchCustomerByEmailSuccess(t *testing.T) {
	repImp := replicated.ReplicatedImpl{
		Client: &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`{
					"query": "email:test@progress.com",
					"total_hits": 1,
					"customers": [
						{
							"id": "2eDnZGGGSjwJN91WZh74A",
							"teamId": "6IopgEQswci9pZWVGGNHU4NyoRbaWe7d",
							"name": "[DEV] Test",
							"email": "test@progress.com",
							"customId": ""
						}
					]
				}`)))
				return &http.Response{
					StatusCode: 200,
					Body:       responseBody,
				}, nil
			},
		},
		ReplicatedConfig: config.ReplicatedConfig{},
		Logger:           logger.NewLogrusStandardLogger(),
	}
	customers, err := repImp.SearchCustomersByEmail("s-no@progress.com", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(customers))

}

func TestSearchCustomerByEmailError(t *testing.T) {
	repImp := replicated.ReplicatedImpl{
		Client: &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("catastrophic failure")
			},
		},
		ReplicatedConfig: config.ReplicatedConfig{},
		Logger:           logger.NewLogrusStandardLogger(),
	}
	_, err := repImp.SearchCustomersByEmail("s-no@progress.com", "")
	assert.Error(t, err)
}

func TestSearchCustomerByEmail500(t *testing.T) {
	repImp := replicated.ReplicatedImpl{
		Client: &MockClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				responseBody := io.NopCloser(bytes.NewReader([]byte(`{
					"message": "error occurred"
				}`)))
				return &http.Response{
					StatusCode: 500,
					Body:       responseBody,
				}, nil
			},
		},
		ReplicatedConfig: config.ReplicatedConfig{},
		Logger:           logger.NewLogrusStandardLogger(),
	}
	_, err := repImp.SearchCustomersByEmail("s-no@progress.com", "")
	assert.Error(t, err)
}

// func TestReplicatedImpl_SearchCustomersByEmail(t *testing.T) {
// 	rc := config.ReplicatedConfig{
// 		URL:   "https://api.replicated.com/vendor/v3",
// 		Token: "e69364a39827225d1042609c4e461e336093305e8a7f58cd47da97107075738d",
// 		AppID: "2dbKte6a9ecfZo6Mn0KTjRvDak4",
// 	}
// 	r := replicated.NewReplicatedImpl(rc, logger.NewLogrusStandardLogger())
// 	customers, err := r.SearchCustomersByEmail("george.westwater@progress.com", "abc123")
// 	assert.NoError(t, err)
// 	assert.Equal(t, 1, len(customers))
// }
