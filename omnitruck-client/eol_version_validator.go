package omnitruck_client

import "fmt"

type EolVersionValidator struct {
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

func (fv *EolVersionValidator) Validate(pval string, p RequestParams) (string, bool) {
	// Allow any version if user specified eol == true in the query
	if p.Get("eol") == "true" {
		return "", true
	}

	product := p.Get("product")
	minVer := SupportedVersion(product)

	if EolProductVersion(product, ProductVersion(pval)) {
		return fmt.Sprintf("%s %s %v is EOL, must be %s", product, fv.GetField(), pval, minVer), false
	} else {
		return "", true
	}
}
