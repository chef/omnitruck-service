package omnitruck_client

import "fmt"

type ContainsValidator[T comparable] struct {
	Field  string
	Values []T
	Code   int
}

func (fv *ContainsValidator[T]) GetField() string {
	return fv.Field
}

func (fv *ContainsValidator[T]) GetValues() interface{} {
	return fv.Values
}

func (fv *ContainsValidator[T]) Msg() string {
	return "%s: %v is not one of %v"
}

func (fv *ContainsValidator[T]) GetCode() int {
	return fv.Code
}

func (fv *ContainsValidator[T]) Validate(pval T, p RequestParams) (string, bool) {
	for _, val := range fv.Values {
		if pval == val {
			return "", true
		}
	}

	return fmt.Sprintf("%s: %v is not one of %v", fv.GetField(), pval, fv.Values), false
}
