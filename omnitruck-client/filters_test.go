package omnitruck_client

import (
	"reflect"
	"testing"
)

func TestFilterList(t *testing.T) {
	type args struct {
		s      []string
		filter func(string) bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "should filter strings",
			args: args{
				s: []string{"a", "b", "c"},
				filter: func(a string) bool {
					if a == "a" {
						return true
					}
					return false
				},
			},
			want: []string{"b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterList(tt.args.s, tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterProductList(t *testing.T) {
	type Product struct {
		Name    string
		Version string
	}
	type args struct {
		s       []Product
		product string
		filter  func(string, Product) bool
	}
	tests := []struct {
		name string
		args args
		want []Product
	}{
		{
			name: "should filter by product",
			args: args{
				s: []Product{
					{
						Name:    "a",
						Version: "1.0",
					},
					{
						Name:    "b",
						Version: "2.0",
					},
				},
				product: "a",
				filter: func(name string, product Product) bool {
					if product.Name == name {
						return true
					}
					return false
				},
			},
			want: []Product{
				{
					Name:    "b",
					Version: "2.0",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterProductList(tt.args.s, tt.args.product, tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterProductList() = %v, want %v", got, tt.want)
			}
		})
	}
}
