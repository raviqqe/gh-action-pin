package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pin "github.com/raviqqe/gh-action-pin"
)

func TestParseUses(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected pin.ActionReference
		ok       bool
	}{
		{
			name: "step action with major version",
			line: "      - uses: actions/checkout@v6",
			expected: pin.ActionReference{
				Prefix:  "      - uses: ",
				Owner:   "actions",
				Repo:    "checkout",
				Version: "v6",
			},
			ok: true,
		},
		{
			name: "step action with full semver",
			line: "      - uses: actions/checkout@v6.2.3",
			expected: pin.ActionReference{
				Prefix:  "      - uses: ",
				Owner:   "actions",
				Repo:    "checkout",
				Version: "v6.2.3",
			},
			ok: true,
		},
		{
			name: "action with existing comment",
			line: "      - uses: actions/checkout@v6 # some comment",
			expected: pin.ActionReference{
				Prefix:  "      - uses: ",
				Owner:   "actions",
				Repo:    "checkout",
				Version: "v6",
				Comment: " # some comment",
			},
			ok: true,
		},
		{
			name: "action without list marker",
			line: "        uses: actions/checkout@v6",
			expected: pin.ActionReference{
				Prefix:  "        uses: ",
				Owner:   "actions",
				Repo:    "checkout",
				Version: "v6",
			},
			ok: true,
		},
		{
			name: "action with sub-path",
			line: "      - uses: owner/repo/path/to/action@v1",
			expected: pin.ActionReference{
				Prefix:  "      - uses: ",
				Owner:   "owner",
				Repo:    "repo",
				Path:    "/path/to/action",
				Version: "v1",
			},
			ok: true,
		},
		{
			name: "action pinned with hash",
			line: "      - uses: actions/checkout@abc123def456abc123def456abc123def456abcd",
			expected: pin.ActionReference{
				Prefix:  "      - uses: ",
				Owner:   "actions",
				Repo:    "checkout",
				Version: "abc123def456abc123def456abc123def456abcd",
			},
			ok: true,
		},
		{
			name: "action with branch reference",
			line: "      - uses: owner/repo@main",
			expected: pin.ActionReference{
				Prefix:  "      - uses: ",
				Owner:   "owner",
				Repo:    "repo",
				Version: "main",
			},
			ok: true,
		},
		{
			name: "local action",
			line: "      - uses: ./local-action",
			ok:   false,
		},
		{
			name: "docker action",
			line: "      - uses: docker://alpine:3.8",
			ok:   false,
		},
		{
			name: "action without version",
			line: "      - uses: actions/checkout",
			ok:   false,
		},
		{
			name: "non-uses line",
			line: "      - run: echo hello",
			ok:   false,
		},
		{
			name: "comment line",
			line: "      # uses: actions/checkout@v6",
			ok:   false,
		},
		{
			name: "empty line",
			line: "",
			ok:   false,
		},
		{
			name: "reusable workflow reference",
			line: "    uses: owner/repo/.github/workflows/ci.yml@v2",
			expected: pin.ActionReference{
				Prefix:  "    uses: ",
				Owner:   "owner",
				Repo:    "repo",
				Path:    "/.github/workflows/ci.yml",
				Version: "v2",
			},
			ok: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := pin.ParseUses(tt.line)

			assert.Equal(t, tt.ok, ok)

			if ok {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestNeedsPin(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "major version",
			version:  "v6",
			expected: true,
		},
		{
			name:     "major and minor version",
			version:  "v6.2",
			expected: true,
		},
		{
			name:     "full semver",
			version:  "v6.2.3",
			expected: true,
		},
		{
			name:     "already pinned with hash",
			version:  "abc123def456abc123def456abc123def456abcd",
			expected: false,
		},
		{
			name:     "branch reference",
			version:  "main",
			expected: true,
		},
		{
			name:     "pre-release version",
			version:  "v6.2.3-beta.1",
			expected: true,
		},
		{
			name:     "version without v prefix",
			version:  "6.2.3",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := pin.ActionReference{Version: tt.version}

			assert.Equal(t, tt.expected, ref.NeedsPin())
		})
	}
}

func TestActionPath(t *testing.T) {
	tests := []struct {
		name     string
		ref      pin.ActionReference
		expected string
	}{
		{
			name:     "simple action",
			ref:      pin.ActionReference{Owner: "actions", Repo: "checkout"},
			expected: "actions/checkout",
		},
		{
			name:     "action with sub-path",
			ref:      pin.ActionReference{Owner: "owner", Repo: "repo", Path: "/path/to/action"},
			expected: "owner/repo/path/to/action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ref.ActionPath())
		})
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "lowercase hex",
			input:    "abc123def456",
			expected: true,
		},
		{
			name:     "uppercase hex",
			input:    "ABC123DEF456",
			expected: true,
		},
		{
			name:     "mixed case hex",
			input:    "aBc123DeF456",
			expected: true,
		},
		{
			name:     "non-hex characters",
			input:    "xyz123",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pin.IsHexString(tt.input))
		})
	}
}

func TestCommentVersion(t *testing.T) {
	tests := []struct {
		name     string
		comment  string
		expected string
	}{
		{
			name:     "full semver comment",
			comment:  " # v6.2.3",
			expected: "v6.2.3",
		},
		{
			name:     "major version comment",
			comment:  " # v6",
			expected: "v6",
		},
		{
			name:     "major and minor version comment",
			comment:  " # v6.2",
			expected: "v6.2",
		},
		{
			name:     "non-version comment",
			comment:  " # some comment",
			expected: "",
		},
		{
			name:     "empty comment",
			comment:  "",
			expected: "",
		},
		{
			name:     "hash only",
			comment:  " #",
			expected: "",
		},
		{
			name:     "version with trailing text",
			comment:  " # v6.2.3 pinned",
			expected: "v6.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := pin.ActionReference{Comment: tt.comment}

			assert.Equal(t, tt.expected, ref.CommentVersion())
		})
	}
}
