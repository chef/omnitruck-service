package omnitruck

import (
	"strings"
)

type ValidatorInterface interface {
	GetCode() int
	Validate(*RequestParams, Context) *ValidationError
}

type ValidationError struct {
	FailedField string
	Value       string
	Tag         string
	Msg         string
	Code        int
}

type Context struct {
	Path      string
	License   bool
	LicenseId string
}

func (e *ValidationError) Error() string {
	return e.Msg
}

type ValidatorFunc func(string, ValidatorInterface) bool

type RequestValidator struct {
	validators []ValidatorInterface
}

type IRequestValidator interface {
	Params(params *RequestParams, c Context) []*ValidationError
	ErrorMessages(errors []*ValidationError) (string, int)
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

func (o *RequestValidator) Params(params *RequestParams, c Context) []*ValidationError {
	var errors []*ValidationError
	for _, vi := range o.validators {
		if err := vi.Validate(params, c); err != nil {
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
