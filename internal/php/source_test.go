package php_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"statora-cli/internal/php"
)

func TestExtractPHPVersion(t *testing.T) {
	cases := []struct {
		url      string
		expected string
	}{
		{"https://www.php.net/distributions/php-8.1.25.tar.gz", "8.1.25"},
		{"https://www.php.net/distributions/php-8.2.15.tar.gz", "8.2.15"},
		{"https://www.php.net/distributions/php-8.3.0.tar.gz", "8.3.0"},
	}
	for _, tc := range cases {
		got := php.ExtractPHPVersion(tc.url)
		assert.Equal(t, tc.expected, got, "url: %s", tc.url)
	}
}

func TestExtractPHPVersion_Empty(t *testing.T) {
	assert.Equal(t, "", php.ExtractPHPVersion("https://example.com/unknown.tar.gz"))
	assert.Equal(t, "", php.ExtractPHPVersion(""))
}

func TestResolveSourceStage_NormalizesVersion(t *testing.T) {
	// ExtractPHPVersion should give back the patch version embedded in the URL.
	url := "https://www.php.net/distributions/php-8.1.25.tar.gz"
	assert.Equal(t, "8.1.25", php.ExtractPHPVersion(url))
}
