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
			got, ok := pin.ParseSemver(tt.tag)

			assert.Equal(t, tt.ok, ok)

			if ok {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestParseVersionSpec(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		ok           bool
		isFullSemver bool
	}{
		{
			name:    "major only",
			version: "v6",
			ok:      true,
		},
		{
			name:    "major and minor",
			version: "v6.2",
			ok:      true,
		},
		{
			name:         "full semver",
			version:      "v6.2.3",
			ok:           true,
			isFullSemver: true,
		},
		{
			name:    "without v prefix",
			version: "6.2.3",
			ok:      false,
		},
		{
			name:    "pre-release",
			version: "v6.2.3-beta.1",
			ok:      false,
		},
		{
			name:    "non-numeric component",
			version: "v6.abc",
			ok:      false,
		},
		{
			name:    "empty",
			version: "",
			ok:      false,
		},
		{
			name:    "just v",
			version: "v",
			ok:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, ok := pin.ParseVersionSpec(tt.version)

			assert.Equal(t, tt.ok, ok)

			if ok {
				assert.Equal(t, tt.isFullSemver, spec.IsFullSemver())
			}
		})
	}
}

func TestVersionSpecMatches(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		version  string
		expected bool
	}{
		{
			name:     "major matches",
			spec:     "v6",
			version:  "v6.2.3",
			expected: true,
		},
		{
			name:     "major does not match",
			spec:     "v6",
			version:  "v7.0.0",
			expected: false,
		},
		{
			name:     "major and minor match",
			spec:     "v6.2",
			version:  "v6.2.5",
			expected: true,
		},
		{
			name:     "minor does not match",
			spec:     "v6.2",
			version:  "v6.3.0",
			expected: false,
		},
		{
			name:     "full semver matches",
			spec:     "v6.2.3",
			version:  "v6.2.3",
			expected: true,
		},
		{
			name:     "patch does not match",
			spec:     "v6.2.3",
			version:  "v6.2.4",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, ok := pin.ParseVersionSpec(tt.spec)
			assert.True(t, ok)
			assert.Equal(t, tt.expected, spec.Matches(tt.version))
		})
	}
}
