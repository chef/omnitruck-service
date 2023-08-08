package services

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/chef/omnitruck-service/models"
)

type MDB struct {
	GetItemfunc func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	Scanfunc func (*dynamodb.ScanInput) (*dynamodb.ScanOutput, error)
}

func (mdb *MDB)GetItem(input *dynamodb.GetItemInput)(*dynamodb.GetItemOutput, error) {
	return mdb.GetItemfunc(input)
}

func (mdb *MDB) Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	return mdb.Scanfunc(input)
}

func TestDbOperationsService_GetPackages(t *testing.T) {
	type fields struct {
		db *dynamodb.DynamoDB
	}
	type args struct {
		partitionKey   string
		partitionValue string
		sortKey        string
		sortValue      string
		tableName      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.ProductDetails
		wantErr bool
	}{
		{
			name: "Successful",
			args: args{
				partitionKey: "product",
				partitionValue: "automate",
				sortKey: "version",
				sortValue: "4.3.9",
				tableName: "test-table",
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
			got, err := ser.GetPackages(tt.args.partitionKey, tt.args.partitionValue, tt.args.sortKey, tt.args.sortValue, tt.args.tableName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DbOperationsService.GetPackages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DbOperationsService.GetPackages() = %v, want %v", got, tt.want)
			}
		})
	}
}
