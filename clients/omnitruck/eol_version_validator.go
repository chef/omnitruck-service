package omnitruck

import (
	"fmt"
)

type EolVersionValidator struct {
	Code int
}

func (fv *EolVersionValidator) GetField() string {
	return "version"
}

func (fv *EolVersionValidator) GetValues() interface{} {
	return nil
}

func (fv *EolVersionValidator) GetCode() int {
	return 400
}

func (fv *EolVersionValidator) Validate(p *RequestParams, c Context) *ValidationError {
	// Allow any version if user specified eol == true in the query
	if p.Eol == "true" {
		return nil
	}

	minVer := SupportedVersion(p.Product)

	if !EolProductVersion(p.Product, ProductVersion(p.Version)) {
		return nil
	}

	return &ValidationError{
		FailedField: "v",
		Value:       p.Version,
		Msg:         fmt.Sprintf("%s %s %v is EOL, must be %s", p.Product, fv.GetField(), p.Version, minVer),
		Code:        fv.Code,
	}
}
