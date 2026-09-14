package main

import (
	"regexp"

	"golang.org/x/mod/semver"
)

var commitHashPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func IsSemver(tag string) bool {
	return semver.IsValid(tag) &&
		semver.Prerelease(tag) == "" &&
		semver.Build(tag) == "" &&
		semver.Canonical(tag) == tag
}

func IsCommitHash(ref string) bool {
	return commitHashPattern.MatchString(ref)
}
