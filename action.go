package main

import (
	"regexp"
	"strings"
)

var usesPattern = regexp.MustCompile(`^(\s*-?\s*uses:\s+)(\S+)(.*)$`)

type ActionReference struct {
	Prefix string
	Owner  string
	Repo   string
	Path   string
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

	if strings.HasPrefix(actionPath, "./") || strings.HasPrefix(actionPath, "docker://") {
		return ActionReference{}, false
	}

	segments := strings.SplitN(actionPath, "/", 3)
	if len(segments) < 2 {
		return ActionReference{}, false
	}

	ref := ActionReference{
		Prefix: prefix,
		Owner:  segments[0],
		Repo:   segments[1],
	}

	if len(segments) == 3 {
		ref.Path = "/" + segments[2]
	}

	return ref, true
}

func (ref ActionReference) ActionPath() string {
	return ref.Owner + "/" + ref.Repo + ref.Path
}
