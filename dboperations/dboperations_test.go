package dboperations

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chef/omnitruck-service/models"
	"github.com/stretchr/testify/assert"
)

type MDB struct {
	GetItemfunc func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	Scanfunc    func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error)
}

// Implement IDynamoDBOps interface from dboperations.go (v1 signatures)
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
		mockItem map[string]types.AttributeValue
		want     interface{}
	}{
		{
			name:  "Success with ProductDetails",
			args:  args{"automate", "4.3.9"},
			model: models.ProductDetails{},
			mockItem: map[string]types.AttributeValue{
				"product": &types.AttributeValueMemberS{Value: "automate"},
				"version": &types.AttributeValueMemberS{Value: "4.3.9"},
			},
			want: &models.ProductDetails{Product: "automate", Version: "4.3.9"},
		},
		{
			name:  "Success with PackageDetails",
			args:  args{"chef-ice", "19.1.27"},
			model: models.PackageDetails{},
			mockItem: map[string]types.AttributeValue{
				"product":  &types.AttributeValueMemberS{Value: "chef-ice"},
				"version":  &types.AttributeValueMemberS{Value: "19.1.27"},
				"metadata": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}},
			},
			want: &models.PackageDetails{
				Product:  "chef-ice",
				Version:  "19.1.27",
				Metadata: map[string]models.Platform{}, // fixed
			},
		},
		{
			name:  "Success with PackageDetails",
			args:  args{"migrate-ice", "19.0.1"},
			model: models.PackageDetails{},
			mockItem: map[string]types.AttributeValue{
				"product":  &types.AttributeValueMemberS{Value: "migrate-ice"},
				"version":  &types.AttributeValueMemberS{Value: "19.0.1"},
				"metadata": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}},
			},
			want: &models.PackageDetails{
				Product:  "migrate-ice",
				Version:  "19.0.1",
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
		mockItem    map[string]types.AttributeValue
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
			mockItem: map[string]types.AttributeValue{
				"foo": &types.AttributeValueMemberS{Value: "bar"},
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
					return nil, &types.ResourceNotFoundException{Message: aws.String("Requested resource not found")}
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
			name: "Successful",
			args: args{
				partitionValue: "autoamte",
			},
			want: []string{
				version4054,
				version4091,
			},
			wantErr: false,
		},
		{
			name: "Successful",
			args: args{
				partitionValue: "chef-ice",
			},
			want: []string{
				version4054,
				version4091,
			},
			wantErr: false,
		},
		{
			name: "Successful",
			args: args{
				partitionValue: "migrate-ice",
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
			var ser *DbOperationsService
			if tt.args.partitionValue == "automate" {
				ser = &DbOperationsService{
					db: &MDB{
						Scanfunc: func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
							return &dynamodb.ScanOutput{
								Items: []map[string]types.AttributeValue{
									{
										"product": &types.AttributeValueMemberS{Value: "automate"},
										"version": &types.AttributeValueMemberS{Value: version4054},
									},
									{
										"product": &types.AttributeValueMemberS{Value: "automate"},
										"version": &types.AttributeValueMemberS{Value: version4091},
									},
								},
							}, nil
						},
					},
					dbModelType: reflect.TypeOf(models.ProductDetails{}),
				}
			} else if tt.args.partitionValue == "chef-ice" {
				ser = &DbOperationsService{
					db: &MDB{
						Scanfunc: func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
							return &dynamodb.ScanOutput{
								Items: []map[string]types.AttributeValue{
									{
										"product": &types.AttributeValueMemberS{Value: "chef-ice"},
										"version": &types.AttributeValueMemberS{Value: version4054},
									},
									{
										"product": &types.AttributeValueMemberS{Value: "chef-ice"},
										"version": &types.AttributeValueMemberS{Value: version4091},
									},
								},
							}, nil
						},
					},
					dbModelType: reflect.TypeOf(models.PackageDetails{}),
				}
			} else {
				ser = &DbOperationsService{
					db: &MDB{
						Scanfunc: func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
							return &dynamodb.ScanOutput{
								Items: []map[string]types.AttributeValue{
									{
										"product": &types.AttributeValueMemberS{Value: "migrate-ice"},
										"version": &types.AttributeValueMemberS{Value: version4054},
									},
									{
										"product": &types.AttributeValueMemberS{Value: "migrate-ice"},
										"version": &types.AttributeValueMemberS{Value: version4091},
									},
								},
							}, nil
						},
					},
					dbModelType: reflect.TypeOf(models.PackageDetails{}),
				}
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
					Scanfunc: func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
						return nil, &types.ReplicaNotFoundException{Message: aws.String("Requested resource not found")}
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
				Item: map[string]types.AttributeValue{
					"product": &types.AttributeValueMemberS{Value: "automate"},
					"version": &types.AttributeValueMemberS{Value: "4.3.9"},
					"metaData": &types.AttributeValueMemberL{Value: []types.AttributeValue{
						&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"architecture": &types.AttributeValueMemberS{Value: "arch64"},
							"platform":     &types.AttributeValueMemberS{Value: "amazon"},
							"sha1":         &types.AttributeValueMemberS{Value: "SHA1arch64"},
							"sha256":       &types.AttributeValueMemberS{Value: "SHA256arch64"},
						}},
						&types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"architecture": &types.AttributeValueMemberS{Value: "x86_64"},
							"platform":     &types.AttributeValueMemberS{Value: "windows"},
							"sha1":         &types.AttributeValueMemberS{Value: "SHA1x86_64"},
							"sha256":       &types.AttributeValueMemberS{Value: "SHA256x86_64"},
						}},
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
				Item: map[string]types.AttributeValue{
					"product": &types.AttributeValueMemberS{Value: "chef-ice"},
					"version": &types.AttributeValueMemberS{Value: "19.1.2"},
					"metadata": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
						"linux": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"x86_64": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
								"deb": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
									"filename":        &types.AttributeValueMemberS{Value: "chef_19.1.27-1_amd64.deb"},
									"install-message": &types.AttributeValueMemberS{Value: ""},
									"sha1":            &types.AttributeValueMemberS{Value: "SHA1x86_64"},
									"sha256":          &types.AttributeValueMemberS{Value: "SHA256x86_64"},
								}},
							}},
						}},
					}},
				},
			},
			dbModelType: reflect.TypeOf(models.PackageDetails{}),
		},
		{
			name: "Success for migrate-ice",
			args: args{
				partitionValue:  "migrate-ice",
				sortValue:       "19.0.1",
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
				FileName:        "migrate-ice-19.0.1-1_amd64.deb",
				PackageManager:  "deb",
			},
			wantErr: false,
			dynamodbResp: dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"product": &types.AttributeValueMemberS{Value: "migrate-ice"},
					"version": &types.AttributeValueMemberS{Value: "19.0.1"},
					"metadata": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
						"linux": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"x86_64": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
								"deb": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
									"filename":        &types.AttributeValueMemberS{Value: "migrate-ice-19.0.1-1_amd64.deb"},
									"install-message": &types.AttributeValueMemberS{Value: ""},
									"sha1":            &types.AttributeValueMemberS{Value: "SHA1x86_64"},
									"sha256":          &types.AttributeValueMemberS{Value: "SHA256x86_64"},
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
							return nil, &types.ReplicaNotFoundException{Message: aws.String("Requested resource not found")}
						}
						return &dynamodb.GetItemOutput{
							Item: map[string]types.AttributeValue{
								"product": &types.AttributeValueMemberS{Value: "chef-ice"},
								"version": &types.AttributeValueMemberS{Value: "19.1.2"},
								"metadata": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
									"linux": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
										"x86_64": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
											"deb": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
												"filename":        &types.AttributeValueMemberS{Value: "chef_19.1.27-1_amd64.deb"},
												"install-message": &types.AttributeValueMemberS{Value: ""},
												"sha1":            &types.AttributeValueMemberS{Value: "SHA1x86_64"},
												"sha256":          &types.AttributeValueMemberS{Value: "SHA256x86_64"},
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ser := &DbOperationsService{
				db: &MDB{
					GetItemfunc: func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
						if tt.wantDBErr {
							return nil, &types.ReplicaNotFoundException{Message: aws.String("Requested resource not found")}
						}
						return &dynamodb.GetItemOutput{
							Item: map[string]types.AttributeValue{
								"product": &types.AttributeValueMemberS{Value: "migrate-ice"},
								"version": &types.AttributeValueMemberS{Value: "19.0.1"},
								"metadata": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
									"linux": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
										"x86_64": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
											"deb": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
												"filename":        &types.AttributeValueMemberS{Value: "migration-tool_19.0.1-1_amd64.deb"},
												"install-message": &types.AttributeValueMemberS{Value: ""},
												"sha1":            &types.AttributeValueMemberS{Value: "SHA1x86_64"},
												"sha256":          &types.AttributeValueMemberS{Value: "SHA256x86_64"},
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
		{
			name: "Success",
			args: args{

				partitionValue: "chef-ice",
			},
			want:    "19.0.0",
			wantErr: false,
		},
		{
			name: "Success",
			args: args{

				partitionValue: "migrate-ice",
			},
			want:    "19.0.1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ser *DbOperationsService
			if tt.args.partitionValue == "automate" {
				ser = &DbOperationsService{
					db: &MDB{
						GetItemfunc: func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
							return &dynamodb.GetItemOutput{
								Item: map[string]types.AttributeValue{
									"product": &types.AttributeValueMemberS{Value: "automate"},
									"version": &types.AttributeValueMemberS{Value: "4.0.91"},
								},
							}, nil
						},
						Scanfunc: func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
							return &dynamodb.ScanOutput{
								Items: []map[string]types.AttributeValue{
									{
										"product": &types.AttributeValueMemberS{Value: "automate"},
										"version": &types.AttributeValueMemberS{Value: "4.0.91"},
									},
								},
							}, nil
						},
					},
					dbModelType: reflect.TypeOf(models.ProductDetails{}),
				}
			} else if tt.args.partitionValue == "chef-ice" {
				ser = &DbOperationsService{
					db: &MDB{
						GetItemfunc: func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
							return &dynamodb.GetItemOutput{
								Item: map[string]types.AttributeValue{
									"product": &types.AttributeValueMemberS{Value: "chef-ice"},
									"version": &types.AttributeValueMemberS{Value: "19.0.0"},
								},
							}, nil
						},
						Scanfunc: func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
							return &dynamodb.ScanOutput{
								Items: []map[string]types.AttributeValue{
									{
										"product": &types.AttributeValueMemberS{Value: "chef-ice"},
										"version": &types.AttributeValueMemberS{Value: "19.0.0"},
									},
								},
							}, nil
						},
					},
					dbModelType: reflect.TypeOf(models.PackageDetails{}),
				}
			} else {
				ser = &DbOperationsService{
					db: &MDB{
						GetItemfunc: func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
							return &dynamodb.GetItemOutput{
								Item: map[string]types.AttributeValue{
									"product": &types.AttributeValueMemberS{Value: "migrate-ice"},
									"version": &types.AttributeValueMemberS{Value: "19.0.1"},
								},
							}, nil
						},
						Scanfunc: func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
							return &dynamodb.ScanOutput{
								Items: []map[string]types.AttributeValue{
									{
										"product": &types.AttributeValueMemberS{Value: "migrate-ice"},
										"version": &types.AttributeValueMemberS{Value: "19.0.1"},
									},
								},
							}, nil
						},
					},
					dbModelType: reflect.TypeOf(models.PackageDetails{}),
				}
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
						return nil, &types.ReplicaNotFoundException{Message: aws.String("Requested resource not found")}
					},
					Scanfunc: func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
						return nil, &types.ReplicaNotFoundException{Message: aws.String("Requested resource not found")}
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
					Scanfunc: func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
						return &dynamodb.ScanOutput{
							Items: []map[string]types.AttributeValue{
								{
									"bom": &types.AttributeValueMemberS{Value: "Chef InSpec"},
									"products": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
										"inspec": &types.AttributeValueMemberS{Value: "Chef InSpec"},
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
					Scanfunc: func(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
						return nil, &types.ReplicaNotFoundException{Message: aws.String("Requested resource not found")}
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
				Items: []map[string]types.AttributeValue{
					{
						"packages": &types.AttributeValueMemberS{Value: "pkg1"},
					},
					{
						"packages": &types.AttributeValueMemberS{Value: "pkg2"},
					},
				},
			},
			expected:    []string{"pkg1", "pkg2"},
			expectError: false,
		},
		{
			name:           "Scan failure",
			mockScanOutput: nil,
			mockScanError: &types.ReplicaNotFoundException{
				Message: aws.String("Requested resource not found"),
			},
			expectError:           true,
			expectedErrorContains: "Requested resource not found",
		},
		{
			name: "Unmarshal failure",
			mockScanOutput: &dynamodb.ScanOutput{
				Items: []map[string]types.AttributeValue{
					{
						"packages": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
							"invalid": &types.AttributeValueMemberS{Value: "oops"},
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
						// Convert v1 mockScanOutput to v2 if needed
						if tt.mockScanOutput != nil {
							// Convert Items if they are v1
							v2Items := []map[string]types.AttributeValue{}
							for _, item := range tt.mockScanOutput.Items {
								v2Item := map[string]types.AttributeValue{}
								for k, v := range item {
									if s, ok := v.(*types.AttributeValueMemberS); ok {
										v2Item[k] = &types.AttributeValueMemberS{Value: s.Value}
									} else if m, ok := v.(*types.AttributeValueMemberM); ok {
										mapped := map[string]types.AttributeValue{}
										for mk, mv := range m.Value {
											if ms, ok := mv.(*types.AttributeValueMemberS); ok {
												mapped[mk] = &types.AttributeValueMemberS{Value: ms.Value}
											}
										}
										v2Item[k] = &types.AttributeValueMemberM{Value: mapped}
									}
								}
								v2Items = append(v2Items, v2Item)
							}
							return &dynamodb.ScanOutput{Items: v2Items}, tt.mockScanError
						}
						return nil, tt.mockScanError
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
