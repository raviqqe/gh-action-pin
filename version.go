package main

import "golang.org/x/mod/semver"

// ParseSemver validates a tag as a full semantic version (vMAJOR.MINOR.PATCH)
// and returns it, rejecting partial versions and pre-release tags.
func ParseSemver(tag string) (string, bool) {
	if !semver.IsValid(tag) {
		return "", false
	}

	if semver.Prerelease(tag) != "" || semver.Build(tag) != "" {
		return "", false
	}

	if semver.Canonical(tag) != tag {
		return "", false
	}

	return tag, true
}
