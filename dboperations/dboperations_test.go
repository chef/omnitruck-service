package dboperations

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/chef/omnitruck-service/models"
	"github.com/stretchr/testify/assert"
)

type MDB struct {
	GetItemfunc func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	Scanfunc    func(*dynamodb.ScanInput) (*dynamodb.ScanOutput, error)
}

func (mdb *MDB) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return mdb.GetItemfunc(input)
}

func (mdb *MDB) Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	return mdb.Scanfunc(input)
}

func TestGetPackagesSuccess(t *testing.T) {
	type args struct {
		partitionValue string
		sortValue      string
	}
	tests := []struct {
		name     string
		args     args
		model    interface{}
		mockItem map[string]*dynamodb.AttributeValue
		want     interface{}
	}{
		{
			name:  "Success with ProductDetails",
			args:  args{"automate", "4.3.9"},
			model: models.ProductDetails{},
			mockItem: map[string]*dynamodb.AttributeValue{
				"product": {S: aws.String("automate")},
				"version": {S: aws.String("4.3.9")},
			},
			want: &models.ProductDetails{Product: "automate", Version: "4.3.9"},
		},
		{
			name:  "Success with PackageDetails",
			args:  args{"chef-ice", "19.1.27"},
			model: models.PackageDetails{},
			mockItem: map[string]*dynamodb.AttributeValue{
				"product": {S: aws.String("chef-ice")},
				"version": {S: aws.String("19.1.27")},
				"metadata": {
					M: map[string]*dynamodb.AttributeValue{}, // simulate structure
				},
			},
			want: &models.PackageDetails{
				Product:  "chef-ice",
				Version:  "19.1.27",
				Metadata: map[string]models.Platform{}, // fixed
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &DbOperationsService{
				db: &MDB{
					GetItemfunc: func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
						return &dynamodb.GetItemOutput{Item: tt.mockItem}, nil
					},
				},
				dbModelType: reflect.TypeOf(tt.model),
			}
			got, err := ser.GetPackages(tt.args.partitionValue, tt.args.sortValue)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
func TestGetPackagesFailure(t *testing.T) {
	type args struct {
		partitionValue string
		sortValue      string
	}
	tests := []struct {
		name        string
		args        args
		model       interface{}
		mockItem    map[string]*dynamodb.AttributeValue
		expectError string
	}{
		{
			name:        "Failure in DB fetch",
			args:        args{"automate", "4.3.9"},
			model:       models.ProductDetails{},
			expectError: "ResourceNotFoundException: Requested resource not found",
		},
		{
			name:  "Unknown type fallback",
			args:  args{"chef", "1.0.0"},
			model: struct{ Foo string }{}, // unknown type
			mockItem: map[string]*dynamodb.AttributeValue{
				"foo": {S: aws.String("bar")},
			},
			expectError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MDB{}
			if tt.mockItem != nil {
				mockDB.GetItemfunc = func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
					return &dynamodb.GetItemOutput{Item: tt.mockItem}, nil
				}
			} else {
				mockDB.GetItemfunc = func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
					return nil, &dynamodb.ResourceNotFoundException{
						Message_: aws.String("Requested resource not found"),
					}
				}
			}

			ser := &DbOperationsService{
				db:          mockDB,
				dbModelType: reflect.TypeOf(tt.model),
			}

			got, err := ser.GetPackages(tt.args.partitionValue, tt.args.sortValue)

			if tt.expectError != "" {
				assert.EqualError(t, err, tt.expectError)
				assert.Nil(t, got)
			} else {
				assert.Nil(t, got)
				assert.NoError(t, err)
			}
		})
	}
}

const (
	version4054 = "4.0.54"
	version4091 = "4.0.91"
)

func TestGetVersionAllSuccess(t *testing.T) {
	type args struct {
		partitionValue string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "SuccessFull",
			args: args{
				partitionValue: "autoamte",
			},
			want: []string{
				version4054,
				version4091,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &DbOperationsService{
				db: &MDB{
					Scanfunc: func(si *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
						return &dynamodb.ScanOutput{
							Items: []map[string]*dynamodb.AttributeValue{
								{
									"product": {S: aws.String("automate")},
									"version": {S: aws.String(version4054)},
								},
								{
									"product": {S: aws.String("automate")},
									"version": {S: aws.String(version4091)},
								},
							},
						}, nil
					},
				},
				dbModelType: reflect.TypeOf(models.ProductDetails{}),
			}
			got, _ := ser.GetVersionAll(tt.args.partitionValue)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestGetVersionAllFailure(t *testing.T) {
	type args struct {
		partitionValue string
		tableName      string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr error
	}{
		{
			name: "Failure in reading the DataBase",
			args: args{
				partitionValue: "autoamte",
				tableName:      "test-table",
			},
			want:    nil,
			wantErr: errors.New("ReplicaNotFoundException: Requested resource not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &DbOperationsService{
				db: &MDB{
					Scanfunc: func(si *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
						return nil, &dynamodb.ReplicaNotFoundException{
							Message_: aws.String("Requested resource not found"),
						}
					},
				},
				dbModelType: reflect.TypeOf(models.ProductDetails{}),
			}
			got, err := ser.GetVersionAll(tt.args.partitionValue)
			assert.Equal(t, got, tt.want)
			assert.Equal(t, err.Error(), tt.wantErr.Error())
		})
	}
}

func TestGetMetaDataSuccess(t *testing.T) {
	type args struct {
		partitionValue  string
		sortValue       string
		platform        string
		platformVersion string
		architecture    string
		packageManager  string
	}
	tests := []struct {
		name         string
		args         args
		want         *models.MetaData
		wantErr      bool
		dynamodbResp dynamodb.GetItemOutput
		dbModelType  reflect.Type
	}{
		{
			name: "Success for automate",
			args: args{
				partitionValue:  "automate",
				sortValue:       "4.0.54",
				platform:        "amazon",
				platformVersion: "2",
				architecture:    "arch64",
			},
			want: &models.MetaData{
				Architecture:    "arch64",
				Platform:        "amazon",
				PlatformVersion: "2",
				SHA1:            "SHA1arch64",
				SHA256:          "SHA256arch64",
			},
			wantErr: false,
			dynamodbResp: dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					"product": {S: aws.String("automate")},
					"version": {S: aws.String("4.3.9")},
					"metaData": {L: []*dynamodb.AttributeValue{
						{
							M: map[string]*dynamodb.AttributeValue{
								"architecture": {S: aws.String(("arch64"))},
								"platform":     {S: aws.String("amazon")},
								"sha1":         {S: aws.String("SHA1arch64")},
								"sha256":       {S: aws.String("SHA256arch64")},
							},
						},
						{
							M: map[string]*dynamodb.AttributeValue{
								"architecture": {S: aws.String(("x86_64"))},
								"platform":     {S: aws.String("windows")},
								"sha1":         {S: aws.String("SHA1x86_64")},
								"sha256":       {S: aws.String("SHA256x86_64")},
							},
						},
					}},
				},
			},
			dbModelType: reflect.TypeOf(models.ProductDetails{}),
		},
		{
			name: "Success for chef-ice",
			args: args{
				partitionValue:  "chef-ice",
				sortValue:       "19.1.2",
				platform:        "linux",
				platformVersion: "",
				architecture:    "x86_64",
				packageManager:  "deb",
			},
			want: &models.MetaData{
				Architecture:    "x86_64",
				Platform:        "linux",
				PlatformVersion: "",
				SHA1:            "SHA1x86_64",
				SHA256:          "SHA256x86_64",
				FileName:        "chef_19.1.27-1_amd64.deb",
				PackageManager:  "deb",
			},
			wantErr: false,
			dynamodbResp: dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					"product": {S: aws.String("chef-ice")},
					"version": {S: aws.String("19.1.2")},
					"metadata": {M: map[string]*dynamodb.AttributeValue{
						"linux": {M: map[string]*dynamodb.AttributeValue{
							"x86_64": {M: map[string]*dynamodb.AttributeValue{
								"deb": {M: map[string]*dynamodb.AttributeValue{
									"filename":        {S: aws.String("chef_19.1.27-1_amd64.deb")},
									"install-message": {S: aws.String("")},
									"sha1":            {S: aws.String("SHA1x86_64")},
									"sha256":          {S: aws.String("SHA256x86_64")},
								}},
							}},
						}},
					}},
				},
			},
			dbModelType: reflect.TypeOf(models.PackageDetails{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &DbOperationsService{
				db: &MDB{
					GetItemfunc: func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
						return &tt.dynamodbResp, nil
					},
				},
				dbModelType: tt.dbModelType,
			}
			got, _ := ser.GetMetaData(tt.args.partitionValue, tt.args.sortValue, tt.args.platform, tt.args.platformVersion, tt.args.architecture, tt.args.packageManager)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetMetaDataFailure(t *testing.T) {
	type args struct {
		partitionValue  string
		sortValue       string
		platform        string
		platformVersion string
		architecture    string
		packageManager  string
	}
	tests := []struct {
		name        string
		args        args
		want        *models.MetaData
		wantDBErr   bool
		errorMsg    error
		dbModelType reflect.Type
	}{
		{
			name: "Failure in reading the DataBase",
			args: args{

				partitionValue: "automate",

				sortValue:       "4.0.54",
				platform:        "amazon",
				platformVersion: "2",
				architecture:    "arch64",
			},
			want:        nil,
			wantDBErr:   true,
			errorMsg:    errors.New("ReplicaNotFoundException: Requested resource not found"),
			dbModelType: reflect.TypeOf(models.ProductDetails{}),
		},
		{
			name: "Failure for wrong model type",
			args: args{
				partitionValue:  "automate",
				sortValue:       "4.0.54",
				platform:        "amazon",
				platformVersion: "2",
				architecture:    "arch64",
			},
			want:        nil,
			wantDBErr:   false,
			errorMsg:    fmt.Errorf("unexpected type %T for product details", &models.MetaData{}),
			dbModelType: reflect.TypeOf(models.MetaData{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &DbOperationsService{
				db: &MDB{
					GetItemfunc: func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
						if tt.wantDBErr {
							return nil, &dynamodb.ReplicaNotFoundException{
								Message_: aws.String("Requested resource not found"),
							}
						}
						return &dynamodb.GetItemOutput{
							Item: map[string]*dynamodb.AttributeValue{
								"product": {S: aws.String("chef-ice")},
								"version": {S: aws.String("19.1.2")},
								"metadata": {M: map[string]*dynamodb.AttributeValue{
									"linux": {M: map[string]*dynamodb.AttributeValue{
										"x86_64": {M: map[string]*dynamodb.AttributeValue{
											"deb": {M: map[string]*dynamodb.AttributeValue{
												"filename":        {S: aws.String("chef_19.1.27-1_amd64.deb")},
												"install-message": {S: aws.String("")},
												"sha1":            {S: aws.String("SHA1x86_64")},
												"sha256":          {S: aws.String("SHA256x86_64")},
											}},
										}},
									}},
								}},
							},
						}, nil
					},
				},
				dbModelType: tt.dbModelType,
			}
			got, err := ser.GetMetaData(tt.args.partitionValue, tt.args.sortValue, tt.args.platform, tt.args.platformVersion, tt.args.architecture, tt.args.packageManager)
			assert.Nil(t, got)
			assert.Equal(t, tt.errorMsg.Error(), err.Error())
		})
	}
}

func TestGetVersionLatestSuccess(t *testing.T) {
	type args struct {
		partitionValue string
	}
	tests := []struct {
		name        string
		args        args
		want        string
		wantErr     bool
		dbModelType reflect.Type
	}{
		{
			name: "Success",
			args: args{

				partitionValue: "automate",
			},
			want:        "4.0.91",
			wantErr:     false,
			dbModelType: reflect.TypeOf(models.ProductDetails{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &DbOperationsService{
				db: &MDB{
					GetItemfunc: func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
						return &dynamodb.GetItemOutput{
							Item: map[string]*dynamodb.AttributeValue{
								"product": {S: aws.String("automate")},
								"version": {S: aws.String("4.0.91")},
							},
						}, nil
					},
					Scanfunc: func(si *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
						return &dynamodb.ScanOutput{
							Items: []map[string]*dynamodb.AttributeValue{
								{
									"product": {S: aws.String("automate")},
									"version": {S: aws.String("4.0.91")},
								},
							},
						}, nil
					},
				},
				dbModelType: reflect.TypeOf(models.ProductDetails{}),
			}
			got, _ := ser.GetVersionLatest(tt.args.partitionValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetVersionLatestFailure(t *testing.T) {
	type args struct {
		partitionValue string
	}
	tests := []struct {
		name        string
		args        args
		want        string
		wantErr     error
		dbModelType reflect.Type
	}{
		{
			name: "Failure in reading the DataBase",
			args: args{

				partitionValue: "automate",
			},
			want:        "",
			wantErr:     errors.New("ReplicaNotFoundException: Requested resource not found"),
			dbModelType: reflect.TypeOf(models.ProductDetails{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &DbOperationsService{
				db: &MDB{
					GetItemfunc: func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
						return nil, &dynamodb.ReplicaNotFoundException{
							Message_: aws.String("Requested resource not found"),
						}
					},
					Scanfunc: func(si *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
						return nil, &dynamodb.ReplicaNotFoundException{
							Message_: aws.String("Requested resource not found"),
						}
					},
				},
				dbModelType: tt.dbModelType,
			}
			got, err := ser.GetVersionLatest(tt.args.partitionValue)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr.Error(), err.Error())
		})
	}
}

func TestGetRelatedProductsSuccess(t *testing.T) {
	type args struct {
		partitionValue string
	}
	tests := []struct {
		name    string
		args    args
		want    *models.RelatedProducts
		wantErr bool
	}{
		{
			name: "SuccessFull",
			args: args{
				partitionValue: "Chef Inspec",
			},
			want: &models.RelatedProducts{
				Bom: "Chef InSpec",
				Products: map[string]string{
					"inspec": "Chef InSpec",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &DbOperationsService{
				db: &MDB{
					Scanfunc: func(si *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
						return &dynamodb.ScanOutput{
							Items: []map[string]*dynamodb.AttributeValue{
								{
									"bom": {S: aws.String("Chef InSpec")},
									"products": {M: map[string]*dynamodb.AttributeValue{
										"inspec": {S: aws.String("Chef InSpec")},
									}},
								},
							},
						}, nil
					},
				},
			}
			got, _ := ser.GetRelatedProducts(tt.args.partitionValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetRelatedProductsFailure(t *testing.T) {
	type args struct {
		partitionValue string
	}
	tests := []struct {
		name    string
		args    args
		want    *models.RelatedProducts
		wantErr error
	}{
		{
			name: "Failure in reading the DataBase",
			args: args{

				partitionValue: "habitat",
			},
			want:    nil,
			wantErr: errors.New("ReplicaNotFoundException: Requested resource not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &DbOperationsService{
				db: &MDB{
					Scanfunc: func(si *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
						return nil, &dynamodb.ReplicaNotFoundException{
							Message_: aws.String("Requested resource not found"),
						}
					},
				},
			}
			got, _ := ser.GetRelatedProducts(tt.args.partitionValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetPackageManagers(t *testing.T) {
	tests := []struct {
		name                  string
		mockScanOutput        *dynamodb.ScanOutput
		mockScanError         error
		expected              []string
		expectError           bool
		expectedErrorContains string
	}{
		{
			name: "Success",
			mockScanOutput: &dynamodb.ScanOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					{
						"packages": {S: aws.String("pkg1")},
					},
					{
						"packages": {S: aws.String("pkg2")},
					},
				},
			},
			expected:    []string{"pkg1", "pkg2"},
			expectError: false,
		},
		{
			name:           "Scan failure",
			mockScanOutput: nil,
			mockScanError: &dynamodb.ReplicaNotFoundException{
				Message_: aws.String("Requested resource not found"),
			},
			expectError:           true,
			expectedErrorContains: "Requested resource not found",
		},
		{
			name: "Unmarshal failure",
			mockScanOutput: &dynamodb.ScanOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					{
						"packages": {M: map[string]*dynamodb.AttributeValue{
							"invalid": {S: aws.String("oops")},
						}},
					},
				},
			},
			expectError:           true,
			expectedErrorContains: "unmarshal",
		},
		{
			name:                  "Nil response from Scan",
			mockScanOutput:        nil,
			mockScanError:         nil,
			expectError:           true,
			expectedErrorContains: "scan returned no items",
		},
		{
			name: "Nil Items in Scan result",
			mockScanOutput: &dynamodb.ScanOutput{
				Items: nil,
			},
			mockScanError:         nil,
			expectError:           true,
			expectedErrorContains: "scan returned no items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &DbOperationsService{
				db: &MDB{
					Scanfunc: func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
						return tt.mockScanOutput, tt.mockScanError
					},
				},
				packageManagersTable: "package-manager-dev",
			}

			got, err := ser.GetPackageManagers()

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrorContains != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
