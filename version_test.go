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
		expected pin.Semver
		ok       bool
	}{
		{
			name:     "full semver",
			tag:      "v6.2.3",
			expected: pin.Semver{Major: 6, Minor: 2, Patch: 3, Raw: "v6.2.3"},
			ok:       true,
		},
		{
			name:     "zero version",
			tag:      "v0.0.0",
			expected: pin.Semver{Raw: "v0.0.0"},
			ok:       true,
		},
		{
			name:     "large numbers",
			tag:      "v100.200.300",
			expected: pin.Semver{Major: 100, Minor: 200, Patch: 300, Raw: "v100.200.300"},
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

func TestSemverLess(t *testing.T) {
	tests := []struct {
		name     string
		left     pin.Semver
		right    pin.Semver
		expected bool
	}{
		{
			name:     "less by major",
			left:     pin.Semver{Major: 1, Minor: 9, Patch: 9},
			right:    pin.Semver{Major: 2, Minor: 0, Patch: 0},
			expected: true,
		},
		{
			name:     "less by minor",
			left:     pin.Semver{Major: 1, Minor: 2, Patch: 9},
			right:    pin.Semver{Major: 1, Minor: 3, Patch: 0},
			expected: true,
		},
		{
			name:     "less by patch",
			left:     pin.Semver{Major: 1, Minor: 2, Patch: 3},
			right:    pin.Semver{Major: 1, Minor: 2, Patch: 4},
			expected: true,
		},
		{
			name:     "equal",
			left:     pin.Semver{Major: 1, Minor: 2, Patch: 3},
			right:    pin.Semver{Major: 1, Minor: 2, Patch: 3},
			expected: false,
		},
		{
			name:     "greater",
			left:     pin.Semver{Major: 2, Minor: 0, Patch: 0},
			right:    pin.Semver{Major: 1, Minor: 9, Patch: 9},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.left.Less(tt.right))
		})
	}
}

func TestParseVersionSpec(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected pin.VersionSpec
		ok       bool
	}{
		{
			name:     "major only",
			version:  "v6",
			expected: pin.VersionSpec{Major: 6},
			ok:       true,
		},
		{
			name:     "major and minor",
			version:  "v6.2",
			expected: pin.VersionSpec{Major: 6, Minor: 2, HasMinor: true},
			ok:       true,
		},
		{
			name:     "full semver",
			version:  "v6.2.3",
			expected: pin.VersionSpec{Major: 6, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true},
			ok:       true,
		},
		{
			name:    "without v prefix",
			version: "6.2.3",
			ok:      false,
		},
		{
			name:    "non-numeric component",
			version: "v6.abc",
			ok:      false,
		},
		{
			name:    "too many components",
			version: "v1.2.3.4",
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
			got, ok := pin.ParseVersionSpec(tt.version)

			assert.Equal(t, tt.ok, ok)

			if ok {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestVersionSpecMatches(t *testing.T) {
	tests := []struct {
		name     string
		spec     pin.VersionSpec
		version  pin.Semver
		expected bool
	}{
		{
			name:     "major matches",
			spec:     pin.VersionSpec{Major: 6},
			version:  pin.Semver{Major: 6, Minor: 2, Patch: 3},
			expected: true,
		},
		{
			name:     "major does not match",
			spec:     pin.VersionSpec{Major: 6},
			version:  pin.Semver{Major: 7, Minor: 0, Patch: 0},
			expected: false,
		},
		{
			name:     "major and minor match",
			spec:     pin.VersionSpec{Major: 6, Minor: 2, HasMinor: true},
			version:  pin.Semver{Major: 6, Minor: 2, Patch: 5},
			expected: true,
		},
		{
			name:     "minor does not match",
			spec:     pin.VersionSpec{Major: 6, Minor: 2, HasMinor: true},
			version:  pin.Semver{Major: 6, Minor: 3, Patch: 0},
			expected: false,
		},
		{
			name:     "full semver matches",
			spec:     pin.VersionSpec{Major: 6, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true},
			version:  pin.Semver{Major: 6, Minor: 2, Patch: 3},
			expected: true,
		},
		{
			name:     "patch does not match",
			spec:     pin.VersionSpec{Major: 6, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true},
			version:  pin.Semver{Major: 6, Minor: 2, Patch: 4},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.spec.Matches(tt.version))
		})
	}
}

func TestVersionSpecIsFullSemver(t *testing.T) {
	tests := []struct {
		name     string
		spec     pin.VersionSpec
		expected bool
	}{
		{
			name:     "major only",
			spec:     pin.VersionSpec{Major: 6},
			expected: false,
		},
		{
			name:     "major and minor",
			spec:     pin.VersionSpec{Major: 6, Minor: 2, HasMinor: true},
			expected: false,
		},
		{
			name:     "full semver",
			spec:     pin.VersionSpec{Major: 6, Minor: 2, Patch: 3, HasMinor: true, HasPatch: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.spec.IsFullSemver())
		})
	}
}
