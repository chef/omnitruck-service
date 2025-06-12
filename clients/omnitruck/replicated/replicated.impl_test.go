package replicated_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/config"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/logger"
	"github.com/chef/omnitruck-service/models"
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

func TestReplicatedImpl_GetDowloadUrl(t *testing.T) {
	type fields struct {
		ReplicatedConfig config.ReplicatedConfig
		Client           replicated.HTTPClient
		Logger           logger.Logger
	}
	type args struct {
		customer  models.Customer
		requestId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantUrl string
		wantErr bool
	}{
		{
			name: "No channel error",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{customer: models.Customer{
				ID:       "cust123",
				Airgap:   true,
				Channels: []models.Channel{},
			}},
			wantUrl: "",
			wantErr: true,
		},
		{
			name: "Empty slug value",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{customer: models.Customer{
				ID:     "cust123",
				Airgap: true,
				Channels: []models.Channel{
					{ID: "channel123", AppSlug: "app123"},
				},
			}},
			wantUrl: "",
			wantErr: true,
		},
		{
			name: "Non Airgap",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{customer: models.Customer{
				ID:     "cust123",
				Airgap: false,
				Channels: []models.Channel{
					{ID: "channel123", AppSlug: "app123", ChannelSlug: "ch123"},
				},
			}},
			wantUrl: constants.REPLICATED_DOWNLOAD_URL + "/app123/ch123",
			wantErr: false,
		},
		{
			name: "Airgap",
			fields: fields{
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{customer: models.Customer{
				ID:     "cust123",
				Airgap: true,
				Channels: []models.Channel{
					{ID: "channel123", AppSlug: "app123", ChannelSlug: "ch123"},
				},
			}},
			wantUrl: constants.REPLICATED_DOWNLOAD_URL + "/app123/ch123?airgap=true",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &replicated.ReplicatedImpl{
				ReplicatedConfig: tt.fields.ReplicatedConfig,
				Client:           tt.fields.Client,
				Logger:           tt.fields.Logger,
			}
			gotUrl, err := r.GetDowloadUrl(tt.args.customer, tt.args.requestId)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, gotUrl)
			} else {
				assert.NotEmpty(t, gotUrl)
				assert.Equal(t, tt.wantUrl, gotUrl)
				assert.Nil(t, err)
			}

		})
	}
}

func TestReplicatedImpl_DownloadFromReplicated(t *testing.T) {
	type fields struct {
		ReplicatedConfig config.ReplicatedConfig
		Client           replicated.HTTPClient
		Logger           logger.Logger
	}
	type args struct {
		url           string
		requestid     string
		authorization string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRes *http.Response
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				Client: &MockClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						responseBody := io.NopCloser(bytes.NewReader([]byte(`{"message": "success"}`)))
						return &http.Response{
							StatusCode: 200,
							Body:       responseBody,
						}, nil
					},
				},
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				url:           "https://www.example.com",
				requestid:     "req123",
				authorization: "123",
			},
			wantRes: &http.Response{StatusCode: 200},
			wantErr: false,
		},
		{
			name: "Error",
			fields: fields{
				Client: &MockClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						return nil, fmt.Errorf("Error")
					},
				},
				ReplicatedConfig: config.ReplicatedConfig{},
				Logger:           logger.NewLogrusStandardLogger(),
			},
			args: args{
				url:           "https://www.example.com",
				requestid:     "req123",
				authorization: "123",
			},
			wantRes: nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := replicated.ReplicatedImpl{
				ReplicatedConfig: tt.fields.ReplicatedConfig,
				Client:           tt.fields.Client,
				Logger:           tt.fields.Logger,
			}
			gotRes, err := r.DownloadFromReplicated(tt.args.url, tt.args.requestid, tt.args.authorization)

			if tt.wantErr {
				assert.Nil(t, gotRes)
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.wantRes.StatusCode, gotRes.StatusCode)
			}

		})
	}
}
