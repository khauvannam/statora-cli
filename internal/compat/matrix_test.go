package compat_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"statora-cli/internal/compat"
)

func TestResolveComposer(t *testing.T) {
	tests := []struct {
		php      string
		wantErr  bool
		wantEmpty bool
	}{
		{php: "8.2.15", wantEmpty: false},
		{php: "7.4.33", wantEmpty: false},
		{php: "7.1.33", wantEmpty: false},
		{php: "5.6.40", wantEmpty: false},
		{php: "6.0.0", wantEmpty: true}, // no PHP 6 in matrix
		{php: "not-a-version", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.php, func(t *testing.T) {
			got, err := compat.ResolveComposer(tt.php)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.NotEmpty(t, got)
			}
		})
	}
}

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		php      string
		composer string
		want     bool
		wantErr  bool
	}{
		{php: "8.2.15", composer: "2.7.1", want: true},
		{php: "8.2.15", composer: "1.10.27", want: false},
		{php: "7.4.33", composer: "2.5.0", want: true},
		{php: "7.4.33", composer: "1.10.27", want: false},
		{php: "5.6.40", composer: "1.10.27", want: true},
		{php: "5.6.40", composer: "2.0.0", want: false},
		{php: "bad", composer: "2.0.0", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.php+"_"+tt.composer, func(t *testing.T) {
			got, err := compat.IsCompatible(tt.php, tt.composer)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
