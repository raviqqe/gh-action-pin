package main

import "golang.org/x/mod/semver"

func IsSemver(tag string) bool {
	return semver.IsValid(tag) &&
		semver.Prerelease(tag) == "" &&
		semver.Build(tag) == "" &&
		semver.Canonical(tag) == tag
}
