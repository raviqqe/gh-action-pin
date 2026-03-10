package main

import (
	"regexp"
	"strings"
)

var usesPattern = regexp.MustCompile(`^(\s*-?\s*uses:\s+)(\S+)(.*)$`)

type ActionReference struct {
	Prefix  string
	Owner   string
	Repo    string
	Path    string
	Version string
}

func ParseUses(line string) (ActionReference, bool) {
	matches := usesPattern.FindStringSubmatch(line)
	if matches == nil {
		return ActionReference{}, false
	}

	prefix := matches[1]
	spec := matches[2]

	parts := strings.SplitN(spec, "@", 2)
	if len(parts) != 2 || parts[1] == "" {
		return ActionReference{}, false
	}

	actionPath := parts[0]
	version := parts[1]

	if strings.HasPrefix(actionPath, "./") || strings.HasPrefix(actionPath, "docker://") {
		return ActionReference{}, false
	}

	segments := strings.SplitN(actionPath, "/", 3)
	if len(segments) < 2 {
		return ActionReference{}, false
	}

	ref := ActionReference{
		Prefix:  prefix,
		Owner:   segments[0],
		Repo:    segments[1],
		Version: version,
	}

	if len(segments) == 3 {
		ref.Path = "/" + segments[2]
	}

	return ref, true
}

func (ref ActionReference) NeedsPin() bool {
	return len(ref.Version) != 40 || !IsHexString(ref.Version)
}

func (ref ActionReference) ActionPath() string {
	return ref.Owner + "/" + ref.Repo + ref.Path
}

func IsHexString(str string) bool {
	for _, char := range str {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}
