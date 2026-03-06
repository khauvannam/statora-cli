package switcher_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"statora-cli/internal/resolver"
	"statora-cli/internal/switcher"
)

func TestPlan_HasChanges(t *testing.T) {
	p := &switcher.Plan{}
	assert.False(t, p.HasChanges())

	p.Actions = append(p.Actions, switcher.Action{Kind: "php", Next: "8.2.15"})
	assert.True(t, p.HasChanges())
}

func TestPlan_Resolution(t *testing.T) {
	res := resolver.Resolution{PHP: "8.2.15", Composer: "2.7.1", Source: "project"}
	p := &switcher.Plan{Resolution: res}
	assert.Equal(t, "8.2.15", p.Resolution.PHP)
	assert.Equal(t, "project", p.Resolution.Source)
}
