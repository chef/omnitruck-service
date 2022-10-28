package omnitruck_client

import "fmt"

type OsVersionValidator struct {
	Code int
}

func (fv *OsVersionValidator) GetField() string {
	return "version"
}

func (fv *OsVersionValidator) GetValues() interface{} {
	return nil
}

func (fv *OsVersionValidator) GetCode() int {
	return 400
}

func (fv *OsVersionValidator) Validate(p *RequestParams) *ValidationError {
	// Allow any version if user specified eol == true in the query
	if p.Eol == "true" {
		return nil
	}

	minVer := SupportedVersion(p.Product)

	if !OsProductVersion(p.Product, ProductVersion(p.Version)) {
		return nil
	}

	return &ValidationError{
		FailedField: "v",
		Value:       p.Version,
		Msg:         fmt.Sprintf("%s %s %v is not opensource, must be %s", p.Product, fv.GetField(), p.Version, minVer),
		Code:        fv.Code,
	}
}
