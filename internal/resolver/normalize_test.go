package resolver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"statora-cli/internal/resolver"
)

func TestNormalizeInstalled_ExactMatch(t *testing.T) {
	installed := []string{"8.1.20", "8.1.25", "8.2.15"}
	got := resolver.NormalizeInstalled("8.1.25", installed)
	assert.Equal(t, "8.1.25", got)
}

func TestNormalizeInstalled_PartialMinor(t *testing.T) {
	installed := []string{"8.1.20", "8.1.25", "8.2.15"}
	got := resolver.NormalizeInstalled("8.1", installed)
	assert.Equal(t, "8.1.25", got)
}

func TestNormalizeInstalled_PartialMajor(t *testing.T) {
	installed := []string{"8.1.20", "8.2.15"}
	got := resolver.NormalizeInstalled("8", installed)
	assert.Equal(t, "8.2.15", got)
}

func TestNormalizeInstalled_NoMatch(t *testing.T) {
	installed := []string{"8.2.15"}
	got := resolver.NormalizeInstalled("8.1", installed)
	assert.Equal(t, "", got)
}

func TestNormalizeInstalled_Empty(t *testing.T) {
	got := resolver.NormalizeInstalled("8.1", nil)
	assert.Equal(t, "", got)
}
