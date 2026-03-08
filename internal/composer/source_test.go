package composer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"statora-cli/internal/composer"
)

func TestIsPartialVersion(t *testing.T) {
	assert.True(t, composer.IsPartialVersion("2"))
	assert.True(t, composer.IsPartialVersion("2.7"))
	assert.False(t, composer.IsPartialVersion("2.7.4"))  // already concrete
	assert.False(t, composer.IsPartialVersion(">= 2.2")) // constraint
	assert.False(t, composer.IsPartialVersion(""))
	assert.False(t, composer.IsPartialVersion("abc"))
}

func TestMatchesPartialPrefix(t *testing.T) {
	cases := []struct {
		partial  string
		version  string
		expected bool
	}{
		{"2", "2.7.4", true},
		{"2", "2.0.0", true},
		{"2", "1.10.0", false},
		{"2.7", "2.7.4", true},
		{"2.7", "2.7.0", true},
		{"2.7", "2.8.0", false},
		{"2.7", "2.70.0", false},
	}
	for _, tc := range cases {
		got := composer.MatchesPartialPrefix(tc.partial, tc.version)
		assert.Equal(t, tc.expected, got, "partial=%s version=%s", tc.partial, tc.version)
	}
}
