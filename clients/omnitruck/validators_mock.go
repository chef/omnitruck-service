package omnitruck

type MockRequestValidator struct {
	ParamsFunc        func(params *RequestParams, c Context) []*ValidationError
	ErrorMessagesFunc func(errors []*ValidationError) (string, int)
}

func (m *MockRequestValidator) Params(params *RequestParams, c Context) []*ValidationError {
	if m.ParamsFunc != nil {
		return m.ParamsFunc(params, c)
	}
	return nil
}

func (m *MockRequestValidator) ErrorMessages(errors []*ValidationError) (string, int) {
	if m.ErrorMessagesFunc != nil {
		return m.ErrorMessagesFunc(errors)
	}
	return "", 0
}
