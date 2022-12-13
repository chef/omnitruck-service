package omnitruck

import (
	"fmt"
	"reflect"
)

type ContainsValidator struct {
	Field      string
	Values     []string
	Code       int
	AllowEmpty bool
	Skip       func(c Context) bool
}

func (fv *ContainsValidator) GetField() string {
	return fv.Field
}

func (fv *ContainsValidator) GetValues() interface{} {
	return fv.Values
}

func (fv *ContainsValidator) GetCode() int {
	return fv.Code
}

func (fv *ContainsValidator) Validate(p *RequestParams, c Context) *ValidationError {
	if fv.Skip != nil && fv.Skip(c) {
		return nil
	}

	pr := reflect.ValueOf(*p)
	fieldValue := pr.FieldByName(fv.Field).String()

	for _, val := range fv.Values {
		if fv.AllowEmpty && len(fieldValue) == 0 {
			return nil
		}
		if !fv.AllowEmpty && len(fieldValue) == 0 {
			return &ValidationError{
				FailedField: fv.Field,
				Value:       fieldValue,
				Msg:         fmt.Sprintf("%s: cannot be empty", fv.Field),
				Code:        fv.Code,
			}
		}
		if fieldValue != val {
			return &ValidationError{
				FailedField: fv.Field,
				Value:       fieldValue,
				Msg:         fmt.Sprintf("%s: %v must be one of %v", fv.Field, fieldValue, fv.Values),
				Code:        fv.Code,
			}
		}
	}

	return nil
}
