package services

import (
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

func TestGetPackages(t *testing.T) {
	type args struct {
		partitionKey   string
		partitionValue string
		sortKey        string
		sortValue      string
		tableName      string
	}
	tests := []struct {
		name    string
		args    args
		want    models.ProductDetails
		wantErr bool
	}{
		{
			name: "Successful",
			args: args{
				partitionKey:   "product",
				partitionValue: "automate",
				sortKey:        "version",
				sortValue:      "4.3.9",
				tableName:      "test-table",
			},
			want: models.ProductDetails{
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
			got, _ := ser.GetPackages(tt.args.partitionKey, tt.args.partitionValue, tt.args.sortKey, tt.args.sortValue, tt.args.tableName)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestGetVersionAll(t *testing.T) {
	type args struct {
		partitionKey   string
		partitionValue string
		tableName      string
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
				partitionKey:   "product",
				partitionValue: "autoamte",
				tableName:      "test-table",
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
			got, _ := ser.GetVersionAll(tt.args.partitionKey, tt.args.partitionValue, tt.args.tableName)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestGetMetaData(t *testing.T) {
	type args struct {
		partitionKey    string
		partitionValue  string
		sortKey         string
		sortValue       string
		tableName       string
		platform        string
		platformVersion string
		architecture    string
	}
	tests := []struct {
		name    string
		args    args
		want    models.ProductDetails
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				partitionKey:    "product",
				partitionValue:  "automate",
				sortKey:         "version",
				sortValue:       "4.0.54",
				tableName:       "test-table",
				platform:        "amazon",
				platformVersion: "2",
				architecture:    "arch64",
			},
			want: models.ProductDetails{
				Product: "automate",
				Version: "4.3.9",
				MetaData: []models.MetaData{
					{
						Architecture:     "arch64",
						Platform:         "amazon",
						Platform_Version: "2",
						SHA1:             "SHA1arch64",
						SHA256:           "SHA256arch64",
					},
				},
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
			got, _ := ser.GetMetaData(tt.args.partitionKey, tt.args.partitionValue, tt.args.sortKey, tt.args.sortValue, tt.args.tableName, tt.args.platform, tt.args.platformVersion, tt.args.architecture)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetVersionLatest(t *testing.T) {
	type args struct {
		partitionKey   string
		partitionValue string
		tableName      string
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
				partitionKey:   "product",
				partitionValue: "automate",
				tableName:      "test-table",
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
			got, _ := ser.GetVersionLatest(tt.args.partitionKey, tt.args.partitionValue, tt.args.tableName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetRelatedProducts(t *testing.T) {
	type args struct {
		partitionKey   string
		partitionValue string
		tableName      string
	}
	tests := []struct {
		name    string
		args    args
		want    models.Sku
		wantErr bool
	}{
		{
			name: "SuccessFull",
			args: args{
				partitionKey:   "skus",
				partitionValue: "habitat",
				tableName:      "test-table",
			},
			want: models.Sku{
				Skus: "habitat",
				Products: []string{
					"Habitat Premium",
					"Habitat CLI",
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
									"skus": {S: aws.String("habitat")},
									"products": {L: []*dynamodb.AttributeValue{
										{S: aws.String("Habitat Premium")},
										{S: aws.String("Habitat CLI")},
									}},
								},
							},
						}, nil
					},
				},
			}
			got, _ := ser.GetRelatedProducts(tt.args.partitionKey, tt.args.partitionValue, tt.args.tableName)
			assert.Equal(t, tt.want, got)
		})
	}
}
