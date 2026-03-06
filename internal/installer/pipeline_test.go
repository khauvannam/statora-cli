package installer_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"statora-cli/internal/installer"
)

type mockStage struct {
	name    string
	runFunc func(ctx *installer.Context) error
}

func (m *mockStage) Name() string                     { return m.name }
func (m *mockStage) Run(ctx *installer.Context) error { return m.runFunc(ctx) }

func okStage(name string) *mockStage {
	return &mockStage{name: name, runFunc: func(_ *installer.Context) error { return nil }}
}

func failStage(name string) *mockStage {
	return &mockStage{name: name, runFunc: func(_ *installer.Context) error {
		return errors.New("stage failed")
	}}
}

func TestPipeline_AllPass(t *testing.T) {
	order := []string{}
	stages := []*mockStage{
		{name: "a", runFunc: func(_ *installer.Context) error { order = append(order, "a"); return nil }},
		{name: "b", runFunc: func(_ *installer.Context) error { order = append(order, "b"); return nil }},
	}

	p := installer.New(stages[0], stages[1])
	err := p.Run(&installer.Context{Version: "8.2.15", Data: map[string]any{}})
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, order)
}

func TestPipeline_StopsOnFirstError(t *testing.T) {
	ran := []string{}
	p := installer.New(
		&mockStage{name: "ok", runFunc: func(_ *installer.Context) error { ran = append(ran, "ok"); return nil }},
		failStage("fail"),
		&mockStage{name: "never", runFunc: func(_ *installer.Context) error { ran = append(ran, "never"); return nil }},
	)

	err := p.Run(&installer.Context{Version: "8.2.15", Data: map[string]any{}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fail")
	assert.NotContains(t, ran, "never")
}

func TestPipeline_DataPassthrough(t *testing.T) {
	p := installer.New(
		&mockStage{name: "write", runFunc: func(ctx *installer.Context) error {
			ctx.Data["key"] = "value"
			return nil
		}},
		&mockStage{name: "read", runFunc: func(ctx *installer.Context) error {
			assert.Equal(t, "value", ctx.Data["key"])
			return nil
		}},
	)
	_ = p.Run(&installer.Context{Version: "1.0.0", Data: map[string]any{}})
}

func TestPipeline_Empty(t *testing.T) {
	p := installer.New()
	err := p.Run(&installer.Context{Data: map[string]any{}})
	require.NoError(t, err)
}

// Silence unused import for okStage / failStage helpers used above.
var _ = okStage
var _ = failStage
