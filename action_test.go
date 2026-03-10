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
				Prefix: "      - uses: ",
				Owner:  "actions",
				Repo:   "checkout",
			},
			ok: true,
		},
		{
			name: "step action with full semver",
			line: "      - uses: actions/checkout@v6.2.3",
			expected: pin.ActionReference{
				Prefix: "      - uses: ",
				Owner:  "actions",
				Repo:   "checkout",
			},
			ok: true,
		},
		{
			name: "action with existing comment",
			line: "      - uses: actions/checkout@v6 # some comment",
			expected: pin.ActionReference{
				Prefix: "      - uses: ",
				Owner:  "actions",
				Repo:   "checkout",
			},
			ok: true,
		},
		{
			name: "action without list marker",
			line: "        uses: actions/checkout@v6",
			expected: pin.ActionReference{
				Prefix: "        uses: ",
				Owner:  "actions",
				Repo:   "checkout",
			},
			ok: true,
		},
		{
			name: "action with sub-path",
			line: "      - uses: owner/repo/path/to/action@v1",
			expected: pin.ActionReference{
				Prefix: "      - uses: ",
				Owner:  "owner",
				Repo:   "repo",
				Path:   "/path/to/action",
			},
			ok: true,
		},
		{
			name: "action pinned with hash",
			line: "      - uses: actions/checkout@abc123def456abc123def456abc123def456abcd",
			expected: pin.ActionReference{
				Prefix: "      - uses: ",
				Owner:  "actions",
				Repo:   "checkout",
			},
			ok: true,
		},
		{
			name: "action with branch reference",
			line: "      - uses: owner/repo@main",
			expected: pin.ActionReference{
				Prefix: "      - uses: ",
				Owner:  "owner",
				Repo:   "repo",
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
				Prefix: "    uses: ",
				Owner:  "owner",
				Repo:   "repo",
				Path:   "/.github/workflows/ci.yml",
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
