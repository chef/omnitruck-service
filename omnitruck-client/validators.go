package omnitruck_client

import (
	"strings"
)

type ValidatorInterface interface {
	GetCode() int
	Validate(*RequestParams) *ValidationError
}

type ValidationError struct {
	FailedField string
	Value       string
	Tag         string
	Msg         string
	Code        int
}

func (e *ValidationError) Error() string {
	return e.Msg
}

type ValidatorFunc func(string, ValidatorInterface) bool

type RequestValidator struct {
	validators []ValidatorInterface
}

func NewValidator() RequestValidator {
	rv := RequestValidator{
		validators: []ValidatorInterface{},
	}

	return rv
}

func (rv *RequestValidator) GetValidators() []ValidatorInterface {
	return rv.validators
}

func (rv *RequestValidator) Add(f ValidatorInterface) {
	rv.validators = append(rv.validators, f)
}

func (o *RequestValidator) Params(params *RequestParams) []*ValidationError {
	var errors []*ValidationError
	for _, vi := range o.validators {
		if err := vi.Validate(params); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (o *RequestValidator) ErrorMessages(errors []*ValidationError) (string, int) {
	var msgs []string
	var code int

	for _, err := range errors {
		if err.Code > code {
			code = err.Code
		}
		msgs = append(msgs, err.Error())
	}

	return strings.Join(msgs, "\n"), code
}
