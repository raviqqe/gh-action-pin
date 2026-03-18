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
	workflowFiles, err := findYamlFiles(filepath.Join(root, ".github", "workflows", "*"))
	if err != nil {
		return nil, err
	}

	actionFiles, err := findYamlFiles(filepath.Join(root, ".github", "actions", "*", "*"))
	if err != nil {
		return nil, err
	}

	return append(workflowFiles, actionFiles...), nil
}

func findYamlFiles(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var files []string

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			continue
		}

		if strings.HasSuffix(match, ".yml") || strings.HasSuffix(match, ".yaml") {
			files = append(files, match)
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
