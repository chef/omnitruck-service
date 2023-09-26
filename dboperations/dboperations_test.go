package dboperations

import (
	"errors"
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
		name    string
		args    args
		want    *models.ProductDetails
		wantErr bool
	}{
		{
			name: "Successful",
			args: args{
				partitionValue: "automate",
				sortValue:      "4.3.9",
			},
			want: &models.ProductDetails{
				Product: "automate",
				Version: "4.3.9",
			},
			wantErr: false,
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
								"version": {S: aws.String("4.3.9")},
							},
						}, nil
					},
				},
			}
			got, _ := ser.GetPackages(tt.args.partitionValue, tt.args.sortValue)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestGetPackagesFailure(t *testing.T) {
	type args struct {
		partitionValue string
		sortValue      string
	}
	tests := []struct {
		name    string
		args    args
		want    *models.ProductDetails
		wantErr error
	}{
		{
			name: "Failure in reading the DataBase",
			args: args{
				partitionValue: "automate",
				sortValue:      "4.3.9",
			},
			want:    nil,
			wantErr: errors.New("ReplicaNotFoundException: Requested resource not found"),
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
				},
			}
			_, err := ser.GetPackages(tt.args.partitionValue, tt.args.sortValue)
			assert.Equal(t, err.Error(), tt.wantErr.Error())
		})
	}
}

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
				"4.0.54",
				"4.0.91",
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
									"version": {S: aws.String("4.0.54")},
								},
								{
									"product": {S: aws.String("automate")},
									"version": {S: aws.String("4.0.91")},
								},
							},
						}, nil
					},
				},
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
	}
	tests := []struct {
		name    string
		args    args
		want    *models.MetaData
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				partitionValue:  "automate",
				sortValue:       "4.0.54",
				platform:        "amazon",
				platformVersion: "2",
				architecture:    "arch64",
			},
			want: &models.MetaData{

				Architecture:     "arch64",
				Platform:         "amazon",
				Platform_Version: "2",
				SHA1:             "SHA1arch64",
				SHA256:           "SHA256arch64",
			},
			wantErr: false,
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
								"version": {S: aws.String("4.3.9")},
								"metaData": {L: []*dynamodb.AttributeValue{
									{
										M: map[string]*dynamodb.AttributeValue{
											"architecture":     {S: aws.String(("arch64"))},
											"platform":         {S: aws.String("amazon")},
											"platform_version": {S: aws.String("2")},
											"sha1":             {S: aws.String("SHA1arch64")},
											"sha256":           {S: aws.String("SHA256arch64")},
										},
									},
									{
										M: map[string]*dynamodb.AttributeValue{
											"architecture":     {S: aws.String(("x86_64"))},
											"platform":         {S: aws.String("windows")},
											"platform_version": {S: aws.String("11")},
											"sha1":             {S: aws.String("SHA1arch64")},
											"sha256":           {S: aws.String("SHA256arch64")},
										},
									},
								}},
							},
						}, nil
					},
				},
			}
			got, _ := ser.GetMetaData(tt.args.partitionValue, tt.args.sortValue, tt.args.platform, tt.args.platformVersion, tt.args.architecture)
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
	}
	tests := []struct {
		name    string
		args    args
		want    *models.MetaData
		wantErr error
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
			want:    nil,
			wantErr: errors.New("ReplicaNotFoundException: Requested resource not found"),
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
				},
			}
			got, err := ser.GetMetaData(tt.args.partitionValue, tt.args.sortValue, tt.args.platform, tt.args.platformVersion, tt.args.architecture)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr.Error(), err.Error())
		})
	}
}

func TestGetVersionLatestSuccess(t *testing.T) {
	type args struct {
		partitionValue string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Success",
			args: args{

				partitionValue: "automate",
			},
			want:    "4.0.91",
			wantErr: false,
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
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "Failure in reading the DataBase",
			args: args{

				partitionValue: "automate",
			},
			want:    "",
			wantErr: errors.New("ReplicaNotFoundException: Requested resource not found"),
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
