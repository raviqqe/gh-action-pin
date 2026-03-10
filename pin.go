package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func findWorkflowFiles(root string) ([]string, error) {
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

func pinWorkflowFile(path string, resolver versionResolver) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	changed := false

	for index, line := range lines {
		action, ok := parseUses(line)
		if !ok || !action.needsPin() {
			continue
		}

		hash, fullVersion, err := resolver.resolve(action.owner, action.repo, action.version)
		if err != nil {
			return fmt.Errorf("resolving %s@%s: %w", action.actionPath(), action.version, err)
		}

		lines[index] = action.prefix + action.actionPath() + "@" + hash + " # " + fullVersion
		changed = true
	}

	if !changed {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), info.Mode())
}
