package utils

import "testing"

func Test_AddLogFields(t *testing.T) {

	lf := AddLogFields("testApi", "")
	if lf == nil {
		t.Error("Wanted lof field object but got nil")
	}
}
