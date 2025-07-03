package template

import "github.com/chef/omnitruck-service/clients/omnitruck"

type MockTemplateRenderer struct {
	GetScriptfunc func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error)
}

func (mfu *MockTemplateRenderer) GetScript(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
	return mfu.GetScriptfunc(baseUrl, params, filePath)
}
