package logger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"statora-cli/internal/logger"
)

func TestNew_Production(t *testing.T) {
	l, err := logger.New(false)
	require.NoError(t, err)
	assert.NotNil(t, l)
}

func TestNew_Debug(t *testing.T) {
	l, err := logger.New(true)
	require.NoError(t, err)
	assert.NotNil(t, l)
}

func TestMust_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		l := logger.Must(false)
		assert.NotNil(t, l)
	})
}
