package fileutils

import "github.com/chef/omnitruck-service/clients/omnitruck"

type MockFileUtils struct {
	GetScriptfunc func(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error)
}

func (mfu *MockFileUtils) GetScript(baseUrl string, params *omnitruck.RequestParams, filePath string) (string, error) {
	return mfu.GetScriptfunc(baseUrl, params, filePath)
}
