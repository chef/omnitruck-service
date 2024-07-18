package logger

import (
	"errors"
	"testing"

	"gotest.tools/assert"
)

func Test_SetupLogger(t *testing.T) {

	tests := []struct {
		name      string
		format    string
		wantError error
	}{
		{
			name:      "json",
			format:    "json",
			wantError: nil,
		},
		{
			name:      "text",
			format:    "text",
			wantError: errors.New("no encoder registered for name \"text\""),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			_, err := NewLogger(
				"debug", //config.Logging.Level,
				"",
				tt.format,
				true)
			if err != nil && tt.wantError != nil {
				assert.Equal(t, tt.wantError.Error(), err.Error())
			} else {
				//check for no error
				assert.Equal(t, tt.wantError, err)
			}

		})
	}
}
