package main

import (
	"strings"

	"golang.org/x/mod/semver"
)

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

type VersionSpec struct {
	raw        string
	components int
}

func ParseVersionSpec(version string) (VersionSpec, bool) {
	if !semver.IsValid(version) {
		return VersionSpec{}, false
	}

	if semver.Prerelease(version) != "" || semver.Build(version) != "" {
		return VersionSpec{}, false
	}

	return VersionSpec{
		raw:        version,
		components: strings.Count(version, ".") + 1,
	}, true
}

func (spec VersionSpec) Matches(version string) bool {
	switch spec.components {
	case 1:
		return semver.Major(version) == semver.Major(spec.raw)
	case 2:
		return semver.MajorMinor(version) == semver.MajorMinor(spec.raw)
	case 3:
		return semver.Compare(version, spec.raw) == 0
	}

	return false
}

func (spec VersionSpec) IsFullSemver() bool {
	return spec.components == 3
}
