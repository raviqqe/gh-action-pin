package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func FindWorkflowFiles(root string) ([]string, error) {
	var files []string

	for _, pattern := range []string{
		filepath.Join(root, ".github", "workflows", "*"),
		filepath.Join(root, ".github", "actions", "*", "*"),
		filepath.Join(root, "action.yaml"),
	} {
		matched, err := findYamlFiles(pattern)
		if err != nil {
			return nil, err
		}

		files = append(files, matched...)
	}

	return files, nil
}

func findYamlFiles(pattern string) ([]string, error) {
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var files []string

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		} else if info.IsDir() {
			continue
		}

		if strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") {
			files = append(files, path)
		}
	}

	return files, nil
}

func PinWorkflowFile(path string, resolver VersionResolver, warning io.Writer) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	for index, line := range lines {
		action, ok := ParseActionUse(line)
		if !ok {
			continue
		}

		hash, version, err := resolver.Resolve(action.Owner, action.Repo)
		if errors.Is(err, ErrVersionNotFound) {
			_, _ = fmt.Fprintln(warning, "warning:", err)
			continue
		} else if err != nil {
			return fmt.Errorf("resolving %s: %w", action.ActionPath(), err)
		}

		lines[index] = action.Prefix + action.ActionPath() + "@" + hash + " # " + version
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), info.Mode())
}
