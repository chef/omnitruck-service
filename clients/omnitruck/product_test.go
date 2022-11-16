package omnitruck

import (
	"testing"

	"github.com/hashicorp/go-version"
)

func TestNewConstraint(t *testing.T) {
	type args struct {
		i string
		v string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Should working constraint",
			args: args{
				i: ">= 1.0",
				v: "1.5",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, _ := version.NewVersion(tt.args.v)
			if got := NewConstraint(tt.args.i); !got.Check(v) {
				t.Errorf("NewConstraint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupportedVersion(t *testing.T) {
	type args struct {
		product string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should get the product def",
			args: args{
				product: "chef",
			},
			want: ">= 16.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SupportedVersion(tt.args.product); got != tt.want {
				t.Errorf("SupportedVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEolProductName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should not be eol",
			args: args{
				name: "chef",
			},
			want: false,
		},
		{
			name: "should be eol for unknown product",
			args: args{
				name: "unknown",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EolProductName(tt.args.name); got != tt.want {
				t.Errorf("EolProductName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEolProductVersion(t *testing.T) {
	type args struct {
		product string
		v       ProductVersion
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should return true for old chef version",
			args: args{
				product: "chef",
				v:       ProductVersion("0.10.0"),
			},
			want: true,
		},
		{
			name: "should return false for unknown product",
			args: args{
				product: "unknown",
				v:       ProductVersion("1.0.0"),
			},
			want: false,
		},
		{
			name: "should return false for new chef version",
			args: args{
				product: "chef",
				v:       ProductVersion("100.0.0"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EolProductVersion(tt.args.product, tt.args.v); got != tt.want {
				t.Errorf("EolProductVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOsProductName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should be opensource for chef",
			args: args{
				name: "chef",
			},
			want: true,
		},
		{
			name: "should not be opensource for manage",
			args: args{
				name: "manage",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OsProductName(tt.args.name); got != tt.want {
				t.Errorf("OsProductName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOsProductVersion(t *testing.T) {
	type args struct {
		name    string
		version ProductVersion
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should be opensource for old chef",
			args: args{
				name:    "chef",
				version: ProductVersion("12.0.0"),
			},
			want: true,
		},
		{
			name: "should not be opensource new chef",
			args: args{
				name:    "chef",
				version: ProductVersion("18.0.0"),
			},
			want: false,
		},
		{
			name: "should not be opensource for unknown product",
			args: args{
				name:    "unknown",
				version: ProductVersion("18.0.0"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OsProductVersion(tt.args.name, tt.args.version); got != tt.want {
				t.Errorf("OsProductVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
