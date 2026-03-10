package main

import (
	"regexp"
	"strings"
)

var usesPattern = regexp.MustCompile(`^(\s*-?\s*uses:\s+)(\S+)(.*)$`)

type actionReference struct {
	prefix  string
	owner   string
	repo    string
	path    string
	version string
}

func parseUses(line string) (actionReference, bool) {
	matches := usesPattern.FindStringSubmatch(line)
	if matches == nil {
		return actionReference{}, false
	}

	prefix := matches[1]
	spec := matches[2]

	parts := strings.SplitN(spec, "@", 2)
	if len(parts) != 2 || parts[1] == "" {
		return actionReference{}, false
	}

	actionPath := parts[0]
	version := parts[1]

	if strings.HasPrefix(actionPath, "./") || strings.HasPrefix(actionPath, "docker://") {
		return actionReference{}, false
	}

	segments := strings.SplitN(actionPath, "/", 3)
	if len(segments) < 2 {
		return actionReference{}, false
	}

	ref := actionReference{
		prefix:  prefix,
		owner:   segments[0],
		repo:    segments[1],
		version: version,
	}

	if len(segments) == 3 {
		ref.path = "/" + segments[2]
	}

	return ref, true
}

func (ref actionReference) needsPin() bool {
	if len(ref.version) == 40 && isHexString(ref.version) {
		return false
	}

	_, ok := parseVersionSpec(ref.version)
	return ok
}

func (ref actionReference) actionPath() string {
	return ref.owner + "/" + ref.repo + ref.path
}

func isHexString(str string) bool {
	for _, char := range str {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}
