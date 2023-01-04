package omnitruck

import (
	"reflect"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	type fields struct {
		FailedField string
		Value       string
		Tag         string
		Msg         string
		Code        int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "should return an error message",
			fields: fields{
				Msg: "Testing errors",
			},
			want: "Testing errors",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ValidationError{
				FailedField: tt.fields.FailedField,
				Value:       tt.fields.Value,
				Tag:         tt.fields.Tag,
				Msg:         tt.fields.Msg,
				Code:        tt.fields.Code,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name string
		want RequestValidator
	}{
		{
			name: "Should create a request validator",
			want: RequestValidator{
				validators: []ValidatorInterface{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewValidator(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewValidator() = %v, want %v", got, tt.want)
			}
		})
	}
}

type TestValidator struct {
	Code int
	Ok   bool
}

func (t *TestValidator) GetCode() int {
	return t.Code
}
func (t *TestValidator) Validate(p *RequestParams, c Context) *ValidationError {
	if t.Ok {
		return nil
	}
	return &ValidationError{
		Msg: "failed",
	}
}
func NewTestValidator(code int, ok bool) *TestValidator {
	return &TestValidator{
		Code: code,
		Ok:   ok,
	}
}

func TestRequestValidator_Add(t *testing.T) {
	type fields struct {
		validators []ValidatorInterface
	}
	type args struct {
		f *TestValidator
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name:   "should add validator to the list",
			fields: fields{},
			args: args{
				f: &TestValidator{
					Code: 100,
					Ok:   true,
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv := &RequestValidator{
				validators: tt.fields.validators,
			}
			rv.Add(tt.args.f)
			if got := len(rv.GetValidators()); got != tt.want {
				t.Errorf("RequestValidator.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequestValidator_Params(t *testing.T) {
	type fields struct {
		validators []ValidatorInterface
	}
	type args struct {
		params *RequestParams
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*ValidationError
	}{
		{
			name: "should validate params",
			fields: fields{
				validators: []ValidatorInterface{
					&TestValidator{
						Code: 100,
						Ok:   true,
					},
				},
			},
			args: args{
				params: &RequestParams{},
			},
			want: nil,
		},
		{
			name: "should validate params",
			fields: fields{
				validators: []ValidatorInterface{
					&TestValidator{
						Code: 100,
						Ok:   false,
					},
				},
			},
			args: args{
				params: &RequestParams{},
			},
			want: []*ValidationError{
				{
					Msg: "failed",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &RequestValidator{
				validators: tt.fields.validators,
			}
			if got := o.Params(tt.args.params, Context{}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RequestValidator.Params() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequestValidator_ErrorMessages(t *testing.T) {
	type fields struct {
		validators []ValidatorInterface
	}
	type args struct {
		errors []*ValidationError
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  int
	}{
		{
			name: "should return error messages",
			fields: fields{
				validators: []ValidatorInterface{
					&TestValidator{
						Code: 100,
						Ok:   false,
					},
				},
			},
			args: args{
				errors: []*ValidationError{
					{
						Code: 400,
						Msg:  "failed",
					},
				},
			},
			want:  "failed",
			want1: 400,
		},
		{
			name: "should return error messages",
			fields: fields{
				validators: []ValidatorInterface{
					&TestValidator{
						Code: 100,
						Ok:   false,
					},
				},
			},
			args: args{
				errors: []*ValidationError{
					{
						Code: 400,
						Msg:  "failed",
					},
					{
						Code: 500,
						Msg:  "failed again",
					},
				},
			},
			want:  "failed\nfailed again",
			want1: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &RequestValidator{
				validators: tt.fields.validators,
			}
			got, got1 := o.ErrorMessages(tt.args.errors)
			if got != tt.want {
				t.Errorf("RequestValidator.ErrorMessages() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("RequestValidator.ErrorMessages() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
