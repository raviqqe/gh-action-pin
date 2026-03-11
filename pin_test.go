package main_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pin "github.com/raviqqe/gh-action-pin"
)

type mockResolver struct {
	versions map[string]struct {
		hash        string
		fullVersion string
	}
	err error
}

func (resolver *mockResolver) Resolve(owner, repo string) (string, string, error) {
	if resolver.err != nil {
		return "", "", resolver.err
	}

	key := owner + "/" + repo

	resolved, ok := resolver.versions[key]
	if !ok {
		return "", "", fmt.Errorf("%w for %s", pin.ErrVersionNotFound, key)
	}

	return resolved.hash, resolved.fullVersion, nil
}

func TestFindWorkflowFiles(t *testing.T) {
	t.Run("find yaml and yml files", func(t *testing.T) {
		root := t.TempDir()
		workflowDir := filepath.Join(root, ".github", "workflows")

		require.NoError(t, os.MkdirAll(workflowDir, 0755))

		for _, name := range []string{"test.yaml", "build.yml", "notes.txt"} {
			require.NoError(t, os.WriteFile(filepath.Join(workflowDir, name), []byte(""), 0644))
		}

		files, err := pin.FindWorkflowFiles(root)

		require.NoError(t, err)
		assert.Len(t, files, 2)
	})

	t.Run("return nil for missing workflow directory", func(t *testing.T) {
		files, err := pin.FindWorkflowFiles(t.TempDir())

		require.NoError(t, err)
		assert.Nil(t, files)
	})

	t.Run("skip subdirectories", func(t *testing.T) {
		root := t.TempDir()
		workflowDir := filepath.Join(root, ".github", "workflows")

		require.NoError(t, os.MkdirAll(filepath.Join(workflowDir, "subdir"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(workflowDir, "test.yaml"), []byte(""), 0644))

		files, err := pin.FindWorkflowFiles(root)

		require.NoError(t, err)
		assert.Len(t, files, 1)
	})
}

func TestPinWorkflowFile(t *testing.T) {
	resolver := &mockResolver{
		versions: map[string]struct {
			hash        string
			fullVersion string
		}{
			"actions/checkout":              {hash: "aabbccdd00112233445566778899aabbccddeeff", fullVersion: "v6.2.3"},
			"golangci/golangci-lint-action": {hash: "1122334455667788990011223344556677889900", fullVersion: "v9.1.0"},
			"owner/repo":                    {hash: "0011223344556677889900112233445566778899", fullVersion: "v2.0.1"},
		},
	}

	t.Run("pin actions with version tags", func(t *testing.T) {
		content := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - run: go build
      - uses: golangci/golangci-lint-action@v9
`

		expected := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@aabbccdd00112233445566778899aabbccddeeff # v6.2.3
      - run: go build
      - uses: golangci/golangci-lint-action@1122334455667788990011223344556677889900 # v9.1.0
`

		path := filepath.Join(t.TempDir(), "test.yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		require.NoError(t, pin.PinWorkflowFile(path, resolver, io.Discard))

		got, err := os.ReadFile(path)

		require.NoError(t, err)
		assert.Equal(t, expected, string(got))
	})

	t.Run("pin branch references", func(t *testing.T) {
		content := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: owner/repo@main
`

		expected := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: owner/repo@0011223344556677889900112233445566778899 # v2.0.1
`

		path := filepath.Join(t.TempDir(), "test.yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		require.NoError(t, pin.PinWorkflowFile(path, resolver, io.Discard))

		got, err := os.ReadFile(path)

		require.NoError(t, err)
		assert.Equal(t, expected, string(got))
	})

	t.Run("replace existing comments", func(t *testing.T) {
		content := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6 # old comment
`

		expected := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@aabbccdd00112233445566778899aabbccddeeff # v6.2.3
`

		path := filepath.Join(t.TempDir(), "test.yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		require.NoError(t, pin.PinWorkflowFile(path, resolver, io.Discard))

		got, err := os.ReadFile(path)

		require.NoError(t, err)
		assert.Equal(t, expected, string(got))
	})

	t.Run("pin reusable workflow references", func(t *testing.T) {
		content := `name: test
on: push
jobs:
  ci:
    uses: owner/repo/.github/workflows/ci.yml@v2
`

		expected := `name: test
on: push
jobs:
  ci:
    uses: owner/repo/.github/workflows/ci.yml@0011223344556677889900112233445566778899 # v2.0.1
`

		path := filepath.Join(t.TempDir(), "test.yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		require.NoError(t, pin.PinWorkflowFile(path, resolver, io.Discard))

		got, err := os.ReadFile(path)

		require.NoError(t, err)
		assert.Equal(t, expected, string(got))
	})

	t.Run("pin action with sub-path", func(t *testing.T) {
		content := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: owner/repo@v1
`

		expected := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: owner/repo@0011223344556677889900112233445566778899 # v2.0.1
`

		path := filepath.Join(t.TempDir(), "test.yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		require.NoError(t, pin.PinWorkflowFile(path, resolver, io.Discard))

		got, err := os.ReadFile(path)

		require.NoError(t, err)
		assert.Equal(t, expected, string(got))
	})

	t.Run("update hash-pinned actions", func(t *testing.T) {
		updateResolver := &mockResolver{
			versions: map[string]struct {
				hash        string
				fullVersion string
			}{
				"actions/checkout": {hash: "11111111111111111111aaaaaaaaaaaaaaaaaaaa", fullVersion: "v6.3.0"},
			},
		}

		content := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@aabbccdd00112233445566778899aabbccddeeff # v6.2.3
`

		expected := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@11111111111111111111aaaaaaaaaaaaaaaaaaaa # v6.3.0
`

		path := filepath.Join(t.TempDir(), "test.yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		require.NoError(t, pin.PinWorkflowFile(path, updateResolver, io.Discard))

		got, err := os.ReadFile(path)

		require.NoError(t, err)
		assert.Equal(t, expected, string(got))
	})

	t.Run("update hash-pinned actions without comment", func(t *testing.T) {
		content := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@aabbccdd00112233445566778899aabbccddeeff
`

		expected := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@aabbccdd00112233445566778899aabbccddeeff # v6.2.3
`

		path := filepath.Join(t.TempDir(), "test.yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		require.NoError(t, pin.PinWorkflowFile(path, resolver, io.Discard))

		got, err := os.ReadFile(path)

		require.NoError(t, err)
		assert.Equal(t, expected, string(got))
	})

	t.Run("preserve file when already up to date", func(t *testing.T) {
		content := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@aabbccdd00112233445566778899aabbccddeeff # v6.2.3
`

		path := filepath.Join(t.TempDir(), "test.yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		require.NoError(t, pin.PinWorkflowFile(path, resolver, io.Discard))

		got, err := os.ReadFile(path)

		require.NoError(t, err)
		assert.Equal(t, content, string(got))
	})

	t.Run("skip actions without semantic version tags", func(t *testing.T) {
		content := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: unknown/action@v1
`

		expected := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@aabbccdd00112233445566778899aabbccddeeff # v6.2.3
      - uses: unknown/action@v1
`

		warning := &bytes.Buffer{}
		path := filepath.Join(t.TempDir(), "test.yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		require.NoError(t, pin.PinWorkflowFile(path, resolver, warning))

		got, err := os.ReadFile(path)

		require.NoError(t, err)
		assert.Equal(t, expected, string(got))
		assert.Equal(t, "warning: no semantic version tag found for unknown/action\n", warning.String())
	})

	t.Run("fail when resolver returns an error", func(t *testing.T) {
		errorResolver := &mockResolver{
			versions: map[string]struct {
				hash        string
				fullVersion string
			}{},
		}
		errorResolver.err = fmt.Errorf("API rate limit exceeded")

		content := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: unknown/action@v1
`

		path := filepath.Join(t.TempDir(), "test.yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		assert.Error(t, pin.PinWorkflowFile(path, errorResolver, io.Discard))
	})

	t.Run("preserve file when nothing to pin", func(t *testing.T) {
		content := `name: test
on: push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: echo hello
`

		path := filepath.Join(t.TempDir(), "test.yaml")
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
		require.NoError(t, pin.PinWorkflowFile(path, resolver, io.Discard))

		got, err := os.ReadFile(path)

		require.NoError(t, err)
		assert.Equal(t, content, string(got))
	})
}
