package omnitruck

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validator ContainsValidator
		params    RequestParams
		context   Context
		wantErr   *ValidationError
	}{		
		{
			name: "Invalid field value",
			validator: ContainsValidator{
				Field:  "Product",
				Values: []string{"chef", "habitat"},
				Code:   400,
			},
			params: RequestParams{
				Product: "invalid",
			},
			wantErr: &ValidationError{
				FailedField: "Product",
				Value:       "invalid",
				Msg:         "Product: invalid must be one of [chef habitat]",
				Code:        400,
			},
		},
		{
			name: "Empty field value allowed",
			validator: ContainsValidator{
				Field:      "Product",
				Values:     []string{"chef", "habitat"},
				Code:       400,
				AllowEmpty: true,
			},
			params: RequestParams{
				Product: "",
			},
			wantErr: nil,
		},
		{
			name: "Empty field value not allowed",
			validator: ContainsValidator{
				Field:      "Product",
				Values:     []string{"chef", "habitat"},
				Code:       400,
				AllowEmpty: false,
			},
			params: RequestParams{
				Product: "",
			},
			wantErr: &ValidationError{
				FailedField: "Product",
				Value:       "",
				Msg:         "Product: cannot be empty",
				Code:        400,
			},
		},
		{
			name: "Skip validation",
			validator: ContainsValidator{
				Field:  "Product",
				Values: []string{"chef", "habitat"},
				Code:   400,
				Skip: func(c Context) bool {
					return true
				},
			},
			params:  RequestParams{Product: "invalid"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.Validate(&tt.params, tt.context)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
