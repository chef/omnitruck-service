package template

import "github.com/chef/omnitruck-service/clients/omnitruck"

type MockTemplateRennder struct {
	GetScriptfunc func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error)
}

func (mfu *MockTemplateRennder) GetScript(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
	return mfu.GetScriptfunc(baseUrl, params, filePath)
}
