package omnitruck_client

import (
	"strings"
)

type ValidatorInterface interface {
	GetValues() interface{}
	GetField() string
	GetCode() int
	Validate(string, RequestParams) (string, bool)
}

type ValidatorFunc func(string, ValidatorInterface) bool

type RequestValidator struct {
	validators []ValidatorInterface
}

func (rv *RequestValidator) Add(f ValidatorInterface) {
	rv.validators = append(rv.validators, f)
}

type RequestParams interface {
	Get(string) string
}

type ErrorResponse struct {
	FailedField string
	Value       string
	Tag         string
	Msg         string
	Code        int
}

func NewValidator() RequestValidator {
	rv := RequestValidator{
		validators: []ValidatorInterface{},
	}

	return rv
}

func (o *RequestValidator) Params(params RequestParams) []*ErrorResponse {
	var errors []*ErrorResponse
	for _, vi := range o.validators {
		pVal := params.Get(vi.GetField())
		if len(pVal) > 0 {
			if msg, ok := vi.Validate(pVal, params); !ok {
				element := ErrorResponse{
					FailedField: vi.GetField(),
					Value:       pVal,
					Msg:         msg,
					Code:        vi.GetCode(),
				}
				errors = append(errors, &element)
			}
		}
	}

	return errors
}

func (o *RequestValidator) ErrorMessages(errors []*ErrorResponse) (string, int) {
	var msgs []string
	var code int

	for _, err := range errors {
		if err.Code > code {
			code = err.Code
		}
		msgs = append(msgs, err.Msg)
	}

	return strings.Join(msgs, "\n"), code
}
