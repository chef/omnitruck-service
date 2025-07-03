package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogrusStandardLogger(t *testing.T) {
	log := NewLogrusStandardLogger()
	assert.NotNil(t, log)
}

func TestSetupLogger_JSON_NoDebug(t *testing.T) {
	l, err := SetupLogger("info", "", LogFormatJSON, false)
	assert.NoError(t, err)
	assert.NotNil(t, l)
}

func TestSetupLogger_JSON_Debug(t *testing.T) {
	l, err := SetupLogger("debug", "", LogFormatJSON, true)
	assert.NoError(t, err)
	assert.NotNil(t, l)
}

func TestSetupLogger_Text_NoDebug(t *testing.T) {
	l, err := SetupLogger("warn", "", LogFormatText, false)
	assert.NoError(t, err)
	assert.NotNil(t, l)
}

func TestSetupLogger_Text_Debug(t *testing.T) {
	l, err := SetupLogger("error", "", LogFormatText, true)
	assert.NoError(t, err)
	assert.NotNil(t, l)
}

func TestSetupLogger_InvalidLevel(t *testing.T) {
	l, err := SetupLogger("invalid-level", "", LogFormatText, false)
	assert.Error(t, err)
	assert.Nil(t, l)
}
