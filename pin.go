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
	workflowDir := filepath.Join(root, ".github", "workflows")

	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	var files []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			files = append(files, filepath.Join(workflowDir, name))
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
		if errors.Is(err, VersionNotFoundError) {
			fmt.Fprintln(warning, "warning:", err)
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
