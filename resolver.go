package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type versionResolver interface {
	resolve(owner, repo, version string) (hash string, fullVersion string, err error)
}

type semver struct {
	major int
	minor int
	patch int
	raw   string
}

func parseSemver(tag string) (semver, bool) {
	if !strings.HasPrefix(tag, "v") {
		return semver{}, false
	}

	body := tag[1:]
	parts := strings.SplitN(body, ".", 4)

	if len(parts) != 3 {
		return semver{}, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semver{}, false
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semver{}, false
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return semver{}, false
	}

	return semver{major: major, minor: minor, patch: patch, raw: tag}, true
}

func (left semver) less(right semver) bool {
	if left.major != right.major {
		return left.major < right.major
	}

	if left.minor != right.minor {
		return left.minor < right.minor
	}

	return left.patch < right.patch
}

type versionSpec struct {
	major    int
	minor    int
	patch    int
	hasMinor bool
	hasPatch bool
}

func parseVersionSpec(version string) (versionSpec, bool) {
	if !strings.HasPrefix(version, "v") {
		return versionSpec{}, false
	}

	body := version[1:]
	parts := strings.Split(body, ".")

	if len(parts) == 0 || len(parts) > 3 {
		return versionSpec{}, false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return versionSpec{}, false
	}

	spec := versionSpec{major: major}

	if len(parts) >= 2 {
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return versionSpec{}, false
		}

		spec.minor = minor
		spec.hasMinor = true
	}

	if len(parts) == 3 {
		patch, err := strconv.Atoi(parts[2])
		if err != nil {
			return versionSpec{}, false
		}

		spec.patch = patch
		spec.hasPatch = true
	}

	return spec, true
}

func (spec versionSpec) matches(version semver) bool {
	if version.major != spec.major {
		return false
	}

	if spec.hasMinor && version.minor != spec.minor {
		return false
	}

	if spec.hasPatch && version.patch != spec.patch {
		return false
	}

	return true
}

func (spec versionSpec) isFullSemver() bool {
	return spec.hasMinor && spec.hasPatch
}

type githubResolver struct {
	cache map[string]resolvedVersion
}

type resolvedVersion struct {
	hash        string
	fullVersion string
}

func newGithubResolver() *githubResolver {
	return &githubResolver{
		cache: make(map[string]resolvedVersion),
	}
}

func (resolver *githubResolver) resolve(owner, repo, version string) (string, string, error) {
	key := owner + "/" + repo + "@" + version

	if cached, ok := resolver.cache[key]; ok {
		return cached.hash, cached.fullVersion, nil
	}

	spec, ok := parseVersionSpec(version)
	if !ok {
		return "", "", fmt.Errorf("invalid version: %s", version)
	}

	fullVersion := version

	if !spec.isFullSemver() {
		resolved, err := resolver.findFullVersion(owner, repo, version, spec)
		if err != nil {
			return "", "", err
		}

		fullVersion = resolved
	}

	hash, err := resolver.resolveCommitHash(owner, repo, fullVersion)
	if err != nil {
		return "", "", err
	}

	resolver.cache[key] = resolvedVersion{hash: hash, fullVersion: fullVersion}

	return hash, fullVersion, nil
}

type gitRef struct {
	Ref    string `json:"ref"`
	Object struct {
		Sha string `json:"sha"`
	} `json:"object"`
}

func (resolver *githubResolver) findFullVersion(owner, repo, version string, spec versionSpec) (string, error) {
	apiPath := fmt.Sprintf("repos/%s/%s/git/matching-refs/tags/%s.", owner, repo, version)

	out, err := exec.Command("gh", "api", apiPath, "--paginate").Output()
	if err != nil {
		return "", fmt.Errorf("listing tags for %s/%s@%s: %w", owner, repo, version, err)
	}

	var refs []gitRef
	if err := json.Unmarshal(out, &refs); err != nil {
		return "", fmt.Errorf("parsing tags response: %w", err)
	}

	var versions []semver

	for _, ref := range refs {
		tag := strings.TrimPrefix(ref.Ref, "refs/tags/")

		parsed, ok := parseSemver(tag)
		if !ok {
			continue
		}

		if spec.matches(parsed) {
			versions = append(versions, parsed)
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no matching version found for %s/%s@%s", owner, repo, version)
	}

	sort.Slice(versions, func(index, other int) bool {
		return versions[index].less(versions[other])
	})

	return versions[len(versions)-1].raw, nil
}

func (resolver *githubResolver) resolveCommitHash(owner, repo, version string) (string, error) {
	apiPath := fmt.Sprintf("repos/%s/%s/commits/%s", owner, repo, version)

	out, err := exec.Command("gh", "api", apiPath, "--jq", ".sha").Output()
	if err != nil {
		return "", fmt.Errorf("resolving commit hash for %s/%s@%s: %w", owner, repo, version, err)
	}

	return strings.TrimSpace(string(out)), nil
}
