package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"
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

func PinWorkflowFile(path string, resolver VersionResolver) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	changed := false

	for index, line := range lines {
		action, ok := ParseUses(line)
		if !ok {
			continue
		}

		version := action.Version

		if !action.NeedsPin() {
			version = action.CommentVersion()
			if version == "" {
				continue
			}
		}

		spec, ok := ParseVersionSpec(version)
		if ok && spec.IsFullSemver() {
			version = semver.Major(version)
		}

		hash, fullVersion, err := resolver.Resolve(action.Owner, action.Repo, version)
		if err != nil {
			return fmt.Errorf("resolving %s@%s: %w", action.ActionPath(), version, err)
		}

		newLine := action.Prefix + action.ActionPath() + "@" + hash + " # " + fullVersion
		if lines[index] != newLine {
			lines[index] = newLine
			changed = true
		}
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
