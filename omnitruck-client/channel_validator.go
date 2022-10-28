package omnitruck_client

import "fmt"

type ChannelValidator struct {
	Values []string
	Code   int
}

func (fv *ChannelValidator) GetValues() interface{} {
	return fv.Values
}

func (fv *ChannelValidator) Msg() string {
	return "%s: %v is not one of %v"
}

func (fv *ChannelValidator) GetCode() int {
	return fv.Code
}

func (fv *ChannelValidator) Validate(p *RequestParams) *ValidationError {
	for _, val := range fv.Values {
		if p.Channel == val {
			return nil
		}
	}

	return &ValidationError{
		FailedField: "channel",
		Value:       p.Channel,
		Msg:         fmt.Sprintf("channel: %v is not one of %v", p.Channel, fv.Values),
		Code:        fv.Code,
	}
}
