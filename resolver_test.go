package main

import (
	"testing"
)

func TestParseSemver(t *testing.T) {
	tests := []struct {
		name          string
		tag           string
		expectedMajor int
		expectedMinor int
		expectedPatch int
		ok            bool
	}{
		{
			name:          "full semver",
			tag:           "v6.2.3",
			expectedMajor: 6,
			expectedMinor: 2,
			expectedPatch: 3,
			ok:            true,
		},
		{
			name:          "zero version",
			tag:           "v0.0.0",
			expectedMajor: 0,
			expectedMinor: 0,
			expectedPatch: 0,
			ok:            true,
		},
		{
			name:          "large numbers",
			tag:           "v100.200.300",
			expectedMajor: 100,
			expectedMinor: 200,
			expectedPatch: 300,
			ok:            true,
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
			got, ok := parseSemver(tt.tag)
			if ok != tt.ok {
				t.Fatalf("parseSemver(%q): ok = %v, want %v", tt.tag, ok, tt.ok)
			}

			if !ok {
				return
			}

			if got.major != tt.expectedMajor || got.minor != tt.expectedMinor || got.patch != tt.expectedPatch {
				t.Errorf("parseSemver(%q) = %d.%d.%d, want %d.%d.%d",
					tt.tag, got.major, got.minor, got.patch,
					tt.expectedMajor, tt.expectedMinor, tt.expectedPatch)
			}

			if got.raw != tt.tag {
				t.Errorf("parseSemver(%q).raw = %q, want %q", tt.tag, got.raw, tt.tag)
			}
		})
	}
}

func TestSemverLess(t *testing.T) {
	tests := []struct {
		name     string
		left     semver
		right    semver
		expected bool
	}{
		{
			name:     "less by major",
			left:     semver{major: 1, minor: 9, patch: 9},
			right:    semver{major: 2, minor: 0, patch: 0},
			expected: true,
		},
		{
			name:     "less by minor",
			left:     semver{major: 1, minor: 2, patch: 9},
			right:    semver{major: 1, minor: 3, patch: 0},
			expected: true,
		},
		{
			name:     "less by patch",
			left:     semver{major: 1, minor: 2, patch: 3},
			right:    semver{major: 1, minor: 2, patch: 4},
			expected: true,
		},
		{
			name:     "equal",
			left:     semver{major: 1, minor: 2, patch: 3},
			right:    semver{major: 1, minor: 2, patch: 3},
			expected: false,
		},
		{
			name:     "greater",
			left:     semver{major: 2, minor: 0, patch: 0},
			right:    semver{major: 1, minor: 9, patch: 9},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.left.less(tt.right); got != tt.expected {
				t.Errorf("(%d.%d.%d).less(%d.%d.%d) = %v, want %v",
					tt.left.major, tt.left.minor, tt.left.patch,
					tt.right.major, tt.right.minor, tt.right.patch,
					got, tt.expected)
			}
		})
	}
}

func TestParseVersionSpec(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		expectedSpec versionSpec
		ok           bool
	}{
		{
			name:         "major only",
			version:      "v6",
			expectedSpec: versionSpec{major: 6},
			ok:           true,
		},
		{
			name:         "major and minor",
			version:      "v6.2",
			expectedSpec: versionSpec{major: 6, minor: 2, hasMinor: true},
			ok:           true,
		},
		{
			name:         "full semver",
			version:      "v6.2.3",
			expectedSpec: versionSpec{major: 6, minor: 2, patch: 3, hasMinor: true, hasPatch: true},
			ok:           true,
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
			got, ok := parseVersionSpec(tt.version)
			if ok != tt.ok {
				t.Fatalf("parseVersionSpec(%q): ok = %v, want %v", tt.version, ok, tt.ok)
			}

			if !ok {
				return
			}

			if got != tt.expectedSpec {
				t.Errorf("parseVersionSpec(%q) = %+v, want %+v", tt.version, got, tt.expectedSpec)
			}
		})
	}
}

func TestVersionSpecMatches(t *testing.T) {
	tests := []struct {
		name     string
		spec     versionSpec
		version  semver
		expected bool
	}{
		{
			name:     "major matches",
			spec:     versionSpec{major: 6},
			version:  semver{major: 6, minor: 2, patch: 3},
			expected: true,
		},
		{
			name:     "major does not match",
			spec:     versionSpec{major: 6},
			version:  semver{major: 7, minor: 0, patch: 0},
			expected: false,
		},
		{
			name:     "major and minor match",
			spec:     versionSpec{major: 6, minor: 2, hasMinor: true},
			version:  semver{major: 6, minor: 2, patch: 5},
			expected: true,
		},
		{
			name:     "minor does not match",
			spec:     versionSpec{major: 6, minor: 2, hasMinor: true},
			version:  semver{major: 6, minor: 3, patch: 0},
			expected: false,
		},
		{
			name:     "full semver matches",
			spec:     versionSpec{major: 6, minor: 2, patch: 3, hasMinor: true, hasPatch: true},
			version:  semver{major: 6, minor: 2, patch: 3},
			expected: true,
		},
		{
			name:     "patch does not match",
			spec:     versionSpec{major: 6, minor: 2, patch: 3, hasMinor: true, hasPatch: true},
			version:  semver{major: 6, minor: 2, patch: 4},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.spec.matches(tt.version); got != tt.expected {
				t.Errorf("spec.matches(version) = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVersionSpecIsFullSemver(t *testing.T) {
	tests := []struct {
		name     string
		spec     versionSpec
		expected bool
	}{
		{
			name:     "major only",
			spec:     versionSpec{major: 6},
			expected: false,
		},
		{
			name:     "major and minor",
			spec:     versionSpec{major: 6, minor: 2, hasMinor: true},
			expected: false,
		},
		{
			name:     "full semver",
			spec:     versionSpec{major: 6, minor: 2, patch: 3, hasMinor: true, hasPatch: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.spec.isFullSemver(); got != tt.expected {
				t.Errorf("isFullSemver() = %v, want %v", got, tt.expected)
			}
		})
	}
}
