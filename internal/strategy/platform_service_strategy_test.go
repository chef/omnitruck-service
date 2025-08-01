package strategy_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/chef/omnitruck-service/clients"
	"github.com/chef/omnitruck-service/clients/omnitruck"
	"github.com/chef/omnitruck-service/clients/omnitruck/replicated"
	"github.com/chef/omnitruck-service/constants"
	"github.com/chef/omnitruck-service/models"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/chef/omnitruck-service/internal/strategy"
	"github.com/stretchr/testify/assert"
)

func TestPlatformServiceStrategy_GetLatestVersion(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn omnitruck.ProductVersion
		mockErr    error
		expected   string
		ok         bool
		message    string
	}{
		{
			name:       "success returns latest version",
			mockReturn: "3.1.0",
			expected:   "3.1.0",
			ok:         true,
		},
		{
			name:    "error from platform service",
			mockErr: fiber.NewError(http.StatusInternalServerError, "db error"),
			ok:      false,
			message: "db error",
		},
	}

	for _, tt := range tests {
		tt := tt // pin
		t.Run(tt.name, func(t *testing.T) {
			mockPlatform := &omnitruck.MockPlatformServices{
				PlatformVersionLatestFunc: func(params *omnitruck.RequestParams, mode int) (omnitruck.ProductVersion, error) {
					return tt.mockReturn, tt.mockErr
				},
			}

			s := &strategy.PlatformServiceStrategy{
				PlatformService: mockPlatform,
				Log:             log.NewEntry(log.New()),
				Mode:            constants.Commercial,
			}

			version, req := s.GetLatestVersion(&omnitruck.RequestParams{})

			if tt.ok {
				assert.Equal(t, tt.expected, string(version))
				assert.True(t, req.Ok)
			} else {
				assert.Empty(t, version)
				assert.False(t, req.Ok)
				assert.Contains(t, req.Message, tt.message)
			}
		})
	}
}

func TestPlatformServiceStrategy_GetAllVersions(t *testing.T) {
	tests := []struct {
		name        string
		mockReturn  []omnitruck.ProductVersion
		mockErr     error
		expectedLen int
		ok          bool
		message     string
	}{
		{
			name:        "success returns all versions",
			mockReturn:  []omnitruck.ProductVersion{"1.0.0", "2.0.0"},
			expectedLen: 2,
			ok:          true,
		},
		{
			name:    "error from platform service",
			mockErr: fiber.NewError(http.StatusInternalServerError, "db error"),
			ok:      false,
			message: "db error",
		},
	}

	for _, tt := range tests {
		tt := tt // pin loop var
		t.Run(tt.name, func(t *testing.T) {
			mockPlatform := &omnitruck.MockPlatformServices{
				PlatformVersionsAllFunc: func(params *omnitruck.RequestParams, mode int) ([]omnitruck.ProductVersion, error) {
					return tt.mockReturn, tt.mockErr
				},
			}

			s := &strategy.PlatformServiceStrategy{
				PlatformService: mockPlatform,
				Log:             log.NewEntry(log.New()),
				Mode:            constants.Commercial,
			}

			versions, req := s.GetAllVersions(&omnitruck.RequestParams{})

			assert.Equal(t, tt.ok, req.Ok)

			if tt.ok {
				assert.Len(t, versions, tt.expectedLen)
			} else {
				assert.Nil(t, versions)
				assert.Contains(t, req.Message, tt.message)
			}
		})
	}
}

func TestPlatformServiceStrategy_GetPackages(t *testing.T) {
	mockPlatform := &omnitruck.MockPlatformServices{
		PlatformPackagesFunc: func(params *omnitruck.RequestParams, mode int) (omnitruck.PackageList, error) {
			return omnitruck.PackageList{}, nil
		},
	}
	s := &strategy.PlatformServiceStrategy{
		PlatformService: mockPlatform,
	}
	_, err := s.GetPackages(&omnitruck.RequestParams{})
	assert.NoError(t, err)
}

func TestPlatformServiceStrategy_GetMetadata(t *testing.T) {
	tests := []struct {
		name     string
		mockMeta omnitruck.PackageMetadata
		mockErr  error
		expected string
		ok       bool
		contains string
	}{
		{
			name:     "success returns metadata with version",
			mockMeta: omnitruck.PackageMetadata{Version: "3.2.1"},
			expected: "3.2.1",
			ok:       true,
		},
		{
			name:     "error from platform service",
			mockErr:  fiber.NewError(http.StatusInternalServerError, "db error"),
			ok:       false,
			contains: "db error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockPlatform := &omnitruck.MockPlatformServices{
				PlatformMetadataFunc: func(params *omnitruck.RequestParams, mode int) (omnitruck.PackageMetadata, error) {
					return tt.mockMeta, tt.mockErr
				},
			}

			s := &strategy.PlatformServiceStrategy{
				PlatformService: mockPlatform,
				Log:             log.NewEntry(log.New()),
				Mode:            constants.Commercial,
			}

			meta, req := s.GetMetadata(&omnitruck.RequestParams{})

			if tt.ok {
				assert.Equal(t, tt.expected, meta.Version)
				assert.True(t, req.Ok)
			} else {
				assert.Empty(t, meta.Version)
				assert.False(t, req.Ok)
				assert.Contains(t, req.Message, tt.contains)
			}
		})
	}
}

func TestPlatformServiceStrategy_Download(t *testing.T) {
	tests := []struct {
		name                string
		mode                constants.ApiType
		mockLicenseClient   clients.ILicense
		mockReplicated      replicated.IReplicated
		locals              map[string]interface{}
		expectedCode        int
		expectedMsg         string
		expectedDisposition string
		expectError         bool
	}{
		{
			name:         "open source mode should reject download",
			mode:         constants.Opensource,
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "Platform Service does not support download in Open Source mode",
			expectError:  true,
		},
		{
			name: "commercial mode should return file",
			mode: constants.Commercial,
			mockLicenseClient: &clients.MockLicense{
				GetReplicatedCustomerEmailFunc: func(licenseID, url string, resp *clients.Response) *clients.Request {
					return &clients.Request{
						Ok:      true,
						Code:    200,
						Message: "success",
						Body:    []byte(`{"replicatedEmail":"test@example.com","status_code":200,"message":"success"}`),
					}
				},
			},
			mockReplicated: &replicated.MockReplicated{
				SearchCustomersByEmailFunc: func(email, requestId string) ([]models.Customer, error) {
					return []models.Customer{{InstallationId: "install123"}}, nil
				},
				GetDowloadUrlFunc: func(customer models.Customer, requestId string) (string, error) {
					return "http://example.com/download", nil
				},
				DownloadFromReplicatedFunc: func(url, requestId, auth string) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewReader([]byte("ok"))),
						Header:     http.Header{"Content-Disposition": []string{constants.PLATFORM_SERVICE_CONTENT_DISPOSITION}},
					}, nil
				},
			},
			locals: map[string]interface{}{
				"requestid": "req123",
			},
			expectedCode:        200,
			expectedMsg:         "",
			expectedDisposition: constants.PLATFORM_SERVICE_CONTENT_DISPOSITION,
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &strategy.PlatformServiceStrategy{
				Mode:              tt.mode,
				LicenseClient:     tt.mockLicenseClient,
				Replicated:        tt.mockReplicated,
				LicenseServiceUrl: "http://licenseservice",
				Log:               log.NewEntry(log.New()),
			}

			if tt.locals != nil {
				s.Locals = tt.locals
			}

			url, respBody, header, msg, code, err := s.Download(&omnitruck.RequestParams{LicenseId: "lic123"})

			assert.Equal(t, tt.expectedCode, code)
			assert.Equal(t, tt.expectedMsg, msg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, respBody)
				assert.Equal(t, "", url)
				assert.Equal(t, tt.expectedDisposition, header.Get("Content-Disposition"))
			}
		})
	}
}

func TestPlatformServiceStrategy_GetFileName(t *testing.T) {
	mockPlatform := &omnitruck.MockPlatformServices{
		PlatformFilenameFunc: func(params *omnitruck.RequestParams, mode int) (string, error) {
			return "platformservice.tar.gz", nil
		},
	}
	s := &strategy.PlatformServiceStrategy{
		PlatformService: mockPlatform,
		Mode:            constants.Commercial,
	}
	filename, err := s.GetFileName(&omnitruck.RequestParams{})
	assert.NoError(t, err)
	assert.Equal(t, "platformservice.tar.gz", filename)
}

func TestPlatformServiceStrategy_UpdatePackages(t *testing.T) {
	baseUrl := "http://example.com"
	params := &omnitruck.RequestParams{
		Product:      "platform-service",
		Version:      "1.2.3",
		Platform:     "linux",
		Architecture: "amd64",
	}

	packageList := omnitruck.PackageList{
		"linux": omnitruck.PlatformVersionList{
			"pv": omnitruck.ArchList{
				"amd64": omnitruck.PackageMetadata{
					Version: "1.2.3",
					Url:     "",
				},
			},
		},
	}

	s := &strategy.PlatformServiceStrategy{}

	s.UpdatePackages(&packageList, params, baseUrl)

	updated := packageList["linux"]["pv"]["amd64"]
	assert.Contains(t, updated.Url, baseUrl)
	assert.Equal(t, "1.2.3", updated.Version)
	assert.Equal(t, "linux", params.Platform)
	assert.Equal(t, "amd64", params.Architecture)
}

func TestPlatformServiceStrategy_DownloadChefPlatform_Errors(t *testing.T) {
	tests := []struct {
		name                  string
		requestOk             bool
		requestBody           []byte
		requestCode           int
		replicatedStatusCode  int
		replicatedMessage     string
		searchCustomersError  error
		customers             []models.Customer
		getDownloadUrlError   error
		downloadReplicatedErr error
		expectedMsg           string
		expectedCode          int
	}{
		{
			name:         "request not ok",
			requestOk:    false,
			requestCode:  400,
			expectedMsg:  "request failed",
			expectedCode: 400,
		},
		{
			name:         "invalid json in body",
			requestOk:    true,
			requestCode:  200,
			requestBody:  []byte("invalid"),
			expectedMsg:  constants.UNMARSHAL_ERR_MSG,
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:                 "replicated email response status code not 200",
			requestOk:            true,
			requestCode:          200,
			requestBody:          []byte(`{"replicatedEmail":"test@example.com","status_code":400,"message":"failed"}`),
			replicatedStatusCode: 400,
			replicatedMessage:    "failed",
			expectedMsg:          "failed",
			expectedCode:         400,
		},
		{
			name:                 "error searching customers by email",
			requestOk:            true,
			requestCode:          200,
			requestBody:          []byte(`{"replicatedEmail":"test@example.com","status_code":200}`),
			replicatedStatusCode: 200,
			searchCustomersError: fmt.Errorf("some error"),
			expectedMsg:          constants.REPLICATED_CUSTOMER_ERROR,
			expectedCode:         http.StatusInternalServerError,
		},
		{
			name:                 "no replicated customers found",
			requestOk:            true,
			requestCode:          200,
			requestBody:          []byte(`{"replicatedEmail":"test@example.com","status_code":200}`),
			replicatedStatusCode: 200,
			customers:            []models.Customer{},
			expectedMsg:          constants.REPLICATED_CUSTOMER_ERROR,
			expectedCode:         http.StatusInternalServerError,
		},
		{
			name:                 "error from GetDownloadUrl",
			requestOk:            true,
			requestCode:          200,
			requestBody:          []byte(`{"replicatedEmail":"test@example.com","status_code":200}`),
			replicatedStatusCode: 200,
			customers:            []models.Customer{{InstallationId: "id123"}},
			getDownloadUrlError:  fmt.Errorf("url error"),
			expectedMsg:          constants.REPLICATED_DOWNLOAD_ERROR,
			expectedCode:         http.StatusInternalServerError,
		},
		{
			name:                  "error from DownloadFromReplicated",
			requestOk:             true,
			requestCode:           200,
			requestBody:           []byte(`{"replicatedEmail":"test@example.com","status_code":200}`),
			replicatedStatusCode:  200,
			customers:             []models.Customer{{InstallationId: "id123"}},
			downloadReplicatedErr: fmt.Errorf("download error"),
			expectedMsg:           constants.REPLICATED_DOWNLOAD_ERROR,
			expectedCode:          http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLicense := &clients.MockLicense{
				GetReplicatedCustomerEmailFunc: func(licenseID, url string, resp *clients.Response) *clients.Request {
					return &clients.Request{
						Ok:      tt.requestOk,
						Code:    tt.requestCode,
						Message: "request failed",
						Body:    tt.requestBody,
					}
				},
			}
			mockReplicated := &replicated.MockReplicated{
				SearchCustomersByEmailFunc: func(email, reqID string) ([]models.Customer, error) {
					return tt.customers, tt.searchCustomersError
				},
				GetDowloadUrlFunc: func(customer models.Customer, reqID string) (string, error) {
					return "http://example.com", tt.getDownloadUrlError
				},
				DownloadFromReplicatedFunc: func(url, reqID, auth string) (*http.Response, error) {
					if tt.downloadReplicatedErr != nil {
						return nil, tt.downloadReplicatedErr
					}
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewReader([]byte("ok"))),
						Header:     http.Header{},
					}, nil
				},
			}

			s := &strategy.PlatformServiceStrategy{
				LicenseClient:     mockLicense,
				Replicated:        mockReplicated,
				LicenseServiceUrl: "http://licenseservice",
				Log:               log.NewEntry(log.New()),
				Locals:            map[string]interface{}{"requestid": "req123"},
			}

			_, _, _, msg, code, err := s.DownloadChefPlatform(&omnitruck.RequestParams{LicenseId: "lic123"})
			assert.Error(t, err)
			assert.Equal(t, tt.expectedMsg, msg)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}

func TestApiService_downloadChefPlatform(t *testing.T) {
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Locals("license_id", "lic01")
	c.Locals("requestid", "req01")
	type fields struct {
		Replicated        replicated.IReplicated
		LicenseClient     clients.ILicense
		mockUnmarshalFunc func(data []byte, v any) error
	}
	type args struct {
		params *omnitruck.RequestParams
		c      *fiber.Ctx
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				Replicated: replicated.MockReplicated{
					GetDowloadUrlFunc: func(customer models.Customer, requestId string) (url string, err error) {
						return "https://replicated.app/embedded/app/beta/channel/chef-360", nil
					},
					DownloadFromReplicatedFunc: func(url, requestId, authorization string) (res *http.Response, err error) {
						mockResponse := http.Response{Status: "200", Body: ioutil.NopCloser(bytes.NewBufferString(("This is body"))), Header: make(http.Header)}
						mockResponse.Header.Set("Content-Type", "application/json")
						return &mockResponse, nil
					},
					SearchCustomersByEmailFunc: func(email, requestId string) (customers []models.Customer, err error) {
						return []models.Customer{
							{
								ID:             "2eDnZGGGSjwJNOoWkRr91WZh74A",
								TeamID:         "6IopgEQswci9pZWVGGNHU4NyoRbaWe7d",
								Name:           "[DEV] George Westwater",
								Email:          "george.westwater@progress.com",
								InstallationId: "2eDnZGaJ7x912cC4CQ2U9TRtMbf",
								Airgap:         false,
								Channels: []models.Channel{
									{
										ID:          "2eBqOYbRRv1T0qcIafb9wb0Hvyx",
										AppID:       "2dbKte6a9ecfZo6Mn0KTjRvDak4",
										AppSlug:     "chef-360",
										AppName:     "Chef 360",
										ChannelSlug: "rc",
										Name:        "RC",
									},
								},
							},
						}, nil
					},
				},
				LicenseClient: &clients.MockLicense{
					GetReplicatedCustomerEmailFunc: func(licenseId, licenseServiceUrl string, data *clients.Response) *clients.Request {
						return &clients.Request{
							Code: 200,
							Ok:   true,
							Body: []byte(`{
								"replicatedEmail": "abc@gmail.com",
								"message": "OK",
								"status_code": 200
							  }`),
						}
					},
				},
			},
			args: args{
				params: &omnitruck.RequestParams{
					LicenseId: "abc123",
				},
				c: c,
			},
			wantErr: false,
		},
		{
			name: "License Client Err",
			fields: fields{
				LicenseClient: &clients.MockLicense{
					GetReplicatedCustomerEmailFunc: func(licenseId, licenseServiceUrl string, data *clients.Response) *clients.Request {
						return &clients.Request{
							Code: 200,
							Ok:   false,
						}
					},
				},
			},
			args: args{
				params: &omnitruck.RequestParams{
					LicenseId: "abc123",
				},
				c: c,
			},
			wantErr: true,
		},
		{
			name: "Search Customer error",
			fields: fields{
				Replicated: replicated.MockReplicated{
					GetDowloadUrlFunc: func(customer models.Customer, requestId string) (url string, err error) {
						return "https://replicated.app/embedded/app/beta/channel/chef-360", nil
					},
					DownloadFromReplicatedFunc: func(url, requestId, authorization string) (res *http.Response, err error) {
						return &http.Response{Status: "200", Body: ioutil.NopCloser(bytes.NewBufferString(("This is body")))}, nil
					},
					SearchCustomersByEmailFunc: func(email, requestId string) (customers []models.Customer, err error) {
						return nil, errors.New("Error fetching customer")
					},
				},
				LicenseClient: &clients.MockLicense{
					GetReplicatedCustomerEmailFunc: func(licenseId, licenseServiceUrl string, data *clients.Response) *clients.Request {
						return &clients.Request{
							Code: 200,
							Ok:   true,
							Body: []byte(`{
								"replicatedEmail": "abc@gmail.com",
								"message": "OK",
								"status_code": 200
							  }`),
						}
					},
				},
			},
			args: args{
				params: &omnitruck.RequestParams{
					LicenseId: "abc123",
				},
				c: c,
			},
			wantErr: true,
		},
		{
			name: "0 customers found",
			fields: fields{
				Replicated: replicated.MockReplicated{
					GetDowloadUrlFunc: func(customer models.Customer, requestId string) (url string, err error) {
						return "https://replicated.app/embedded/app/beta/channel/chef-360", nil
					},
					DownloadFromReplicatedFunc: func(url, requestId, authorization string) (res *http.Response, err error) {
						return &http.Response{Status: "200", Body: ioutil.NopCloser(bytes.NewBufferString(("This is body")))}, nil
					},
					SearchCustomersByEmailFunc: func(email, requestId string) (customers []models.Customer, err error) {
						return []models.Customer{}, nil
					},
				},
				LicenseClient: &clients.MockLicense{
					GetReplicatedCustomerEmailFunc: func(licenseId, licenseServiceUrl string, data *clients.Response) *clients.Request {
						return &clients.Request{
							Code: 200,
							Ok:   true,
							Body: []byte(`{
								"replicatedEmail": "abc@gmail.com",
								"message": "OK",
								"status_code": 200
							  }`),
						}
					},
				},
			},
			args: args{
				params: &omnitruck.RequestParams{
					LicenseId: "abc123",
				},
				c: c,
			},
			wantErr: true,
		},
		{
			name: "GetUrl Err",
			fields: fields{
				Replicated: replicated.MockReplicated{
					GetDowloadUrlFunc: func(customer models.Customer, requestId string) (url string, err error) {
						return "", errors.New("error getting download url")
					},
					DownloadFromReplicatedFunc: func(url, requestId, authorization string) (res *http.Response, err error) {
						return &http.Response{Status: "200", Body: ioutil.NopCloser(bytes.NewBufferString(("This is body")))}, nil
					},
					SearchCustomersByEmailFunc: func(email, requestId string) (customers []models.Customer, err error) {
						return []models.Customer{
							{
								ID:             "2eDnZGGGSjwJNOoWkRr91WZh74A",
								TeamID:         "6IopgEQswci9pZWVGGNHU4NyoRbaWe7d",
								Name:           "[DEV] George Westwater",
								Email:          "george.westwater@progress.com",
								InstallationId: "2eDnZGaJ7x912cC4CQ2U9TRtMbf",
								Airgap:         false,
								Channels: []models.Channel{
									{
										ID:          "2eBqOYbRRv1T0qcIafb9wb0Hvyx",
										AppID:       "2dbKte6a9ecfZo6Mn0KTjRvDak4",
										AppSlug:     "chef-360",
										AppName:     "Chef 360",
										ChannelSlug: "rc",
										Name:        "RC",
									},
								},
							},
						}, nil
					},
				},
				LicenseClient: &clients.MockLicense{
					GetReplicatedCustomerEmailFunc: func(licenseId, licenseServiceUrl string, data *clients.Response) *clients.Request {
						return &clients.Request{
							Code: 200,
							Ok:   true,
							Body: []byte(`{
								"replicatedEmail": "abc@gmail.com",
								"message": "OK",
								"status_code": 200
							  }`),
						}
					},
				},
			},
			args: args{
				params: &omnitruck.RequestParams{
					LicenseId: "abc123",
				},
				c: c,
			},
			wantErr: true,
		},
		{
			name: "Download error",
			fields: fields{
				Replicated: replicated.MockReplicated{
					GetDowloadUrlFunc: func(customer models.Customer, requestId string) (url string, err error) {
						return "https://replicated.app/embedded/app/beta/channel/chef-360", nil
					},
					DownloadFromReplicatedFunc: func(url, requestId, authorization string) (res *http.Response, err error) {
						return nil, errors.New("error downloading")
					},
					SearchCustomersByEmailFunc: func(email, requestId string) (customers []models.Customer, err error) {
						return []models.Customer{
							{
								ID:             "2eDnZGGGSjwJNOoWkRr91WZh74A",
								TeamID:         "6IopgEQswci9pZWVGGNHU4NyoRbaWe7d",
								Name:           "[DEV] George Westwater",
								Email:          "george.westwater@progress.com",
								InstallationId: "2eDnZGaJ7x912cC4CQ2U9TRtMbf",
								Airgap:         false,
								Channels: []models.Channel{
									{
										ID:          "2eBqOYbRRv1T0qcIafb9wb0Hvyx",
										AppID:       "2dbKte6a9ecfZo6Mn0KTjRvDak4",
										AppSlug:     "chef-360",
										AppName:     "Chef 360",
										ChannelSlug: "rc",
										Name:        "RC",
									},
								},
							},
						}, nil
					},
				},
				LicenseClient: &clients.MockLicense{
					GetReplicatedCustomerEmailFunc: func(licenseId, licenseServiceUrl string, data *clients.Response) *clients.Request {
						return &clients.Request{
							Code: 200,
							Ok:   true,
							Body: []byte(`{
								"replicatedEmail": "abc@gmail.com",
								"message": "OK",
								"status_code": 200
							  }`),
						}
					},
				},
			},
			args: args{
				params: &omnitruck.RequestParams{
					LicenseId: "abc123",
				},
				c: c,
			},
			wantErr: true,
		},
		{
			name: "Unmarshal Error",
			fields: fields{
				Replicated: replicated.MockReplicated{
					GetDowloadUrlFunc: func(customer models.Customer, requestId string) (url string, err error) {
						return "https://replicated.app/embedded/app/beta/channel/chef-360", nil
					},
					DownloadFromReplicatedFunc: func(url, requestId, authorization string) (res *http.Response, err error) {
						mockResponse := http.Response{Status: "200", Body: ioutil.NopCloser(bytes.NewBufferString(("This is body"))), Header: make(http.Header)}
						mockResponse.Header.Set("Content-Type", "application/json")
						return &mockResponse, nil
					},
					SearchCustomersByEmailFunc: func(email, requestId string) (customers []models.Customer, err error) {
						return []models.Customer{
							{
								ID:             "2eDnZGGGSjwJNOoWkRr91WZh74A",
								TeamID:         "6IopgEQswci9pZWVGGNHU4NyoRbaWe7d",
								Name:           "[DEV] George Westwater",
								Email:          "george.westwater@progress.com",
								InstallationId: "2eDnZGaJ7x912cC4CQ2U9TRtMbf",
								Airgap:         false,
								Channels: []models.Channel{
									{
										ID:          "2eBqOYbRRv1T0qcIafb9wb0Hvyx",
										AppID:       "2dbKte6a9ecfZo6Mn0KTjRvDak4",
										AppSlug:     "chef-360",
										AppName:     "Chef 360",
										ChannelSlug: "rc",
										Name:        "RC",
									},
								},
							},
						}, nil
					},
				},
				LicenseClient: &clients.MockLicense{
					GetReplicatedCustomerEmailFunc: func(licenseId, licenseServiceUrl string, data *clients.Response) *clients.Request {
						return &clients.Request{
							Code: 200,
							Ok:   true,
							Body: []byte(`{
								"replicatedEmail": "abc@gmail.com",
								"message": "OK",
								"status_code": 200
							  }`),
						}
					},
				},
				mockUnmarshalFunc: func(data []byte, v any) error {
					return errors.New("error occurred")
				},
			},
			args: args{
				params: &omnitruck.RequestParams{
					LicenseId: "abc123",
				},
				c: c,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up custom unmarshal/ioCopy if provided
			if tt.fields.mockUnmarshalFunc != nil {
				strategy.JsonUnmarshal = tt.fields.mockUnmarshalFunc
			} else {
				strategy.JsonUnmarshal = func(data []byte, v any) error {
					return json.Unmarshal(data, &v)
				}
			}
			s := &strategy.PlatformServiceStrategy{
				Replicated:    tt.fields.Replicated,
				LicenseClient: tt.fields.LicenseClient,
				Log:           log.NewEntry(log.New()),
				Locals:        map[string]interface{}{"requestid": "req01"},
			}

			_, _, _, msg, code, err := s.DownloadChefPlatform(tt.args.params)
			if tt.wantErr {
				assert.NotNil(t, err, "expected error but got nil")
			} else {
				assert.Nil(t, err, "expected no error but got one: %v", err)
			}

			_ = msg // Optionally assert msg/code if needed
			_ = code
		})
	}
}
