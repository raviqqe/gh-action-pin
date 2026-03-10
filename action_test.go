package main_test

import (
	"testing"

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
			line: "      - uses: actions/checkout@abc123def456abc123def456abc123def456abc12345",
			expected: pin.ActionReference{
				Prefix:  "      - uses: ",
				Owner:   "actions",
				Repo:    "checkout",
				Version: "abc123def456abc123def456abc123def456abc12345",
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
			if ok != tt.ok {
				t.Fatalf("ParseUses(%q): ok = %v, want %v", tt.line, ok, tt.ok)
			}

			if !ok {
				return
			}

			if got != tt.expected {
				t.Errorf("ParseUses(%q) = %+v, want %+v", tt.line, got, tt.expected)
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
			version:  "abc123def456abc123def456abc123def456abc12345",
			expected: false,
		},
		{
			name:     "branch reference",
			version:  "main",
			expected: false,
		},
		{
			name:     "pre-release version",
			version:  "v6.2.3-beta.1",
			expected: false,
		},
		{
			name:     "version without v prefix",
			version:  "6.2.3",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := pin.ActionReference{Version: tt.version}

			if got := ref.NeedsPin(); got != tt.expected {
				t.Errorf("NeedsPin() with version %q = %v, want %v", tt.version, got, tt.expected)
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
			if got := tt.ref.ActionPath(); got != tt.expected {
				t.Errorf("ActionPath() = %q, want %q", got, tt.expected)
			}
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
			if got := pin.IsHexString(tt.input); got != tt.expected {
				t.Errorf("IsHexString(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
