package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"golang.org/x/mod/semver"
)

type VersionResolver interface {
	Resolve(owner, repo, version string) (hash string, fullVersion string, err error)
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

func (resolver *githubResolver) Resolve(owner, repo, version string) (string, string, error) {
	key := owner + "/" + repo + "@" + version

	if cached, ok := resolver.cache[key]; ok {
		return cached.hash, cached.fullVersion, nil
	}

	fullVersion, err := resolver.resolveFullVersion(owner, repo, version)
	if err != nil {
		return "", "", err
	}

	hash, err := resolver.resolveCommitHash(owner, repo, fullVersion)
	if err != nil {
		return "", "", err
	}

	resolver.cache[key] = resolvedVersion{hash: hash, fullVersion: fullVersion}

	return hash, fullVersion, nil
}

func (resolver *githubResolver) resolveFullVersion(owner, repo, version string) (string, error) {
	spec, ok := ParseVersionSpec(version)
	if ok && spec.IsFullSemver() {
		return version, nil
	}

	if ok {
		return resolver.findHighestVersion(owner, repo, version+".", &spec)
	}

	return resolver.findHighestVersion(owner, repo, "v", nil)
}

type gitRef struct {
	Ref    string `json:"ref"`
	Object struct {
		Sha string `json:"sha"`
	} `json:"object"`
}

func (resolver *githubResolver) findHighestVersion(owner, repo, tagPrefix string, spec *VersionSpec) (string, error) {
	apiPath := fmt.Sprintf("repos/%s/%s/git/matching-refs/tags/%s", owner, repo, tagPrefix)

	out, err := exec.Command("gh", "api", apiPath, "--paginate").Output()
	if err != nil {
		return "", fmt.Errorf("listing tags for %s/%s: %w", owner, repo, err)
	}

	var refs []gitRef
	if err := json.Unmarshal(out, &refs); err != nil {
		return "", fmt.Errorf("parsing tags response: %w", err)
	}

	var versions []string

	for _, ref := range refs {
		tag := strings.TrimPrefix(ref.Ref, "refs/tags/")

		parsed, ok := ParseSemver(tag)
		if !ok {
			continue
		}

		if spec != nil && !spec.Matches(parsed) {
			continue
		}

		versions = append(versions, parsed)
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no semantic version tag found for %s/%s", owner, repo)
	}

	sort.Slice(versions, func(index, other int) bool {
		return semver.Compare(versions[index], versions[other]) < 0
	})

	return versions[len(versions)-1], nil
}

func (resolver *githubResolver) resolveCommitHash(owner, repo, version string) (string, error) {
	apiPath := fmt.Sprintf("repos/%s/%s/commits/%s", owner, repo, version)

	out, err := exec.Command("gh", "api", apiPath, "--jq", ".sha").Output()
	if err != nil {
		return "", fmt.Errorf("resolving commit hash for %s/%s@%s: %w", owner, repo, version, err)
	}

	return strings.TrimSpace(string(out)), nil
}
