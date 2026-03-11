package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pin "github.com/raviqqe/gh-action-pin"
)

func TestIsSemver(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected bool
	}{
		{
			name:     "full semver",
			tag:      "v6.2.3",
			expected: true,
		},
		{
			name:     "zero version",
			tag:      "v0.0.0",
			expected: true,
		},
		{
			name:     "large numbers",
			tag:      "v100.200.300",
			expected: true,
		},
		{
			name:     "major only",
			tag:      "v6",
			expected: false,
		},
		{
			name:     "major and minor",
			tag:      "v6.2",
			expected: false,
		},
		{
			name:     "without v prefix",
			tag:      "6.2.3",
			expected: false,
		},
		{
			name:     "pre-release",
			tag:      "v6.2.3-beta.1",
			expected: false,
		},
		{
			name:     "build metadata",
			tag:      "v6.2.3+build",
			expected: false,
		},
		{
			name:     "non-numeric",
			tag:      "v6.2.abc",
			expected: false,
		},
		{
			name:     "empty",
			tag:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pin.IsSemver(tt.tag))
		})
	}
}
