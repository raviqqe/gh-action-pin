package main

import (
	"testing"
)

func TestParseUses(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected actionReference
		ok       bool
	}{
		{
			name: "step action with major version",
			line: "      - uses: actions/checkout@v6",
			expected: actionReference{
				prefix:  "      - uses: ",
				owner:   "actions",
				repo:    "checkout",
				version: "v6",
			},
			ok: true,
		},
		{
			name: "step action with full semver",
			line: "      - uses: actions/checkout@v6.2.3",
			expected: actionReference{
				prefix:  "      - uses: ",
				owner:   "actions",
				repo:    "checkout",
				version: "v6.2.3",
			},
			ok: true,
		},
		{
			name: "action with existing comment",
			line: "      - uses: actions/checkout@v6 # some comment",
			expected: actionReference{
				prefix:  "      - uses: ",
				owner:   "actions",
				repo:    "checkout",
				version: "v6",
			},
			ok: true,
		},
		{
			name: "action without list marker",
			line: "        uses: actions/checkout@v6",
			expected: actionReference{
				prefix:  "        uses: ",
				owner:   "actions",
				repo:    "checkout",
				version: "v6",
			},
			ok: true,
		},
		{
			name: "action with sub-path",
			line: "      - uses: owner/repo/path/to/action@v1",
			expected: actionReference{
				prefix:  "      - uses: ",
				owner:   "owner",
				repo:    "repo",
				path:    "/path/to/action",
				version: "v1",
			},
			ok: true,
		},
		{
			name: "action pinned with hash",
			line: "      - uses: actions/checkout@abc123def456abc123def456abc123def456abc12345",
			expected: actionReference{
				prefix:  "      - uses: ",
				owner:   "actions",
				repo:    "checkout",
				version: "abc123def456abc123def456abc123def456abc12345",
			},
			ok: true,
		},
		{
			name: "action with branch reference",
			line: "      - uses: owner/repo@main",
			expected: actionReference{
				prefix:  "      - uses: ",
				owner:   "owner",
				repo:    "repo",
				version: "main",
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
			expected: actionReference{
				prefix:  "    uses: ",
				owner:   "owner",
				repo:    "repo",
				path:    "/.github/workflows/ci.yml",
				version: "v2",
			},
			ok: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseUses(tt.line)
			if ok != tt.ok {
				t.Fatalf("parseUses(%q): ok = %v, want %v", tt.line, ok, tt.ok)
			}

			if !ok {
				return
			}

			if got != tt.expected {
				t.Errorf("parseUses(%q) = %+v, want %+v", tt.line, got, tt.expected)
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
			ref := actionReference{version: tt.version}

			if got := ref.needsPin(); got != tt.expected {
				t.Errorf("needsPin() with version %q = %v, want %v", tt.version, got, tt.expected)
			}
		})
	}
}

func TestActionPath(t *testing.T) {
	tests := []struct {
		name     string
		ref      actionReference
		expected string
	}{
		{
			name:     "simple action",
			ref:      actionReference{owner: "actions", repo: "checkout"},
			expected: "actions/checkout",
		},
		{
			name:     "action with sub-path",
			ref:      actionReference{owner: "owner", repo: "repo", path: "/path/to/action"},
			expected: "owner/repo/path/to/action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ref.actionPath(); got != tt.expected {
				t.Errorf("actionPath() = %q, want %q", got, tt.expected)
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
			if got := isHexString(tt.input); got != tt.expected {
				t.Errorf("isHexString(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
