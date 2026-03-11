package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pin "github.com/raviqqe/gh-action-pin"
)

func TestParseSemver(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		expected string
		ok       bool
	}{
		{
			name:     "full semver",
			tag:      "v6.2.3",
			expected: "v6.2.3",
			ok:       true,
		},
		{
			name:     "zero version",
			tag:      "v0.0.0",
			expected: "v0.0.0",
			ok:       true,
		},
		{
			name:     "large numbers",
			tag:      "v100.200.300",
			expected: "v100.200.300",
			ok:       true,
		},
		{
			name: "major only",
			tag:  "v6",
			ok:   false,
		},
		{
			name: "major and minor",
			tag:  "v6.2",
			ok:   false,
		},
		{
			name: "without v prefix",
			tag:  "6.2.3",
			ok:   false,
		},
		{
			name: "pre-release",
			tag:  "v6.2.3-beta.1",
			ok:   false,
		},
		{
			name: "build metadata",
			tag:  "v6.2.3+build",
			ok:   false,
		},
		{
			name: "non-numeric",
			tag:  "v6.2.abc",
			ok:   false,
		},
		{
			name: "empty",
			tag:  "",
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := pin.IsSemver(tt.tag)

			assert.Equal(t, tt.ok, ok)

			if ok {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
