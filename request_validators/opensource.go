package request_validator

import (
	"fmt"
	"strings"

	validator "github.com/go-playground/validator/v10"
)

type OpensourceValidator struct {
	validate *validator.Validate
}

type RequestParams interface {
}

type ErrorResponse struct {
	FailedField string
	Value       string
	Tag         string
	Msg         string
	Code        int
}

func NewOpensourceValidator() OpensourceValidator {
	return OpensourceValidator{
		validate: validator.New(),
	}
}

func (o *OpensourceValidator) ValidateParams(params RequestParams) []*ErrorResponse {
	var errors []*ErrorResponse
	err := o.validate.Struct(params)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			msg, code := o.FieldError(err)
			element.FailedField = err.Field()
			element.Tag = err.Tag()
			element.Value = err.Param()
			element.Msg = msg
			element.Code = code
			errors = append(errors, &element)
		}
	}

	return errors
}

// FieldError translates validation failures into user friendly messages
func (o *OpensourceValidator) FieldError(err validator.FieldError) (string, int) {
	switch err.Field() {
	case "Channel":
		return fmt.Sprintf("Authorization failed for %s %s, please use %s instead", strings.ToLower(err.Field()), err.Value(), err.Param()), 403
	default:
		return fmt.Sprintf("Validation failed for %s (%s %s %s)", err.Field(), err.Value(), err.Tag(), err.Param()), 400
	}
}

func (o *OpensourceValidator) ErrorMessages(errors []*ErrorResponse) (string, int) {
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
