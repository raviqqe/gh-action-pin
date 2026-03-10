package main

import (
	"strconv"
	"strings"
)

type Semver struct {
	Major int
	Minor int
	Patch int
	Raw   string
}

func ParseSemver(tag string) (Semver, bool) {
	if !strings.HasPrefix(tag, "v") {
		return Semver{}, false
	}

	body := tag[1:]
	parts := strings.SplitN(body, ".", 4)

	if len(parts) != 3 {
		return Semver{}, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Semver{}, false
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Semver{}, false
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Semver{}, false
	}

	return Semver{Major: major, Minor: minor, Patch: patch, Raw: tag}, true
}

func (left Semver) Less(right Semver) bool {
	if left.Major != right.Major {
		return left.Major < right.Major
	}

	if left.Minor != right.Minor {
		return left.Minor < right.Minor
	}

	return left.Patch < right.Patch
}

type VersionSpec struct {
	Major    int
	Minor    int
	Patch    int
	HasMinor bool
	HasPatch bool
}

func ParseVersionSpec(version string) (VersionSpec, bool) {
	if !strings.HasPrefix(version, "v") {
		return VersionSpec{}, false
	}

	body := version[1:]
	parts := strings.Split(body, ".")

	if len(parts) == 0 || len(parts) > 3 {
		return VersionSpec{}, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return VersionSpec{}, false
	}

	spec := VersionSpec{Major: major}

	if len(parts) >= 2 {
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return VersionSpec{}, false
		}

		spec.Minor = minor
		spec.HasMinor = true
	}

	if len(parts) == 3 {
		patch, err := strconv.Atoi(parts[2])
		if err != nil {
			return VersionSpec{}, false
		}

		spec.Patch = patch
		spec.HasPatch = true
	}

	return spec, true
}

func (spec VersionSpec) Matches(version Semver) bool {
	if version.Major != spec.Major {
		return false
	}

	if spec.HasMinor && version.Minor != spec.Minor {
		return false
	}

	if spec.HasPatch && version.Patch != spec.Patch {
		return false
	}

	return true
}

func (spec VersionSpec) IsFullSemver() bool {
	return spec.HasMinor && spec.HasPatch
}
