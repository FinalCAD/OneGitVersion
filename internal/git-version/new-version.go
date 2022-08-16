package gitVersion

import (
	"DotnetGitHubVersion/internal/utils/uarray"
	"github.com/blang/semver/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"strconv"
	"strings"
)

func createNewVersion(gitVersion GitVersion, environment *Environment, name string, bumpPatch bool) (*semver.Version, error) {
	var latestVersion *semver.Version
	var tagRef *plumbing.Hash
	var err error
	var found bool
	defaultVersion, err := semver.Make(gitVersion.GetService().Version)
	if err != nil {
		return nil, err
	}
	latestVersion, tagRef, found, err = getLastVersion(gitVersion, environment, name, defaultVersion)
	head, _ := gitVersion.GetRepository().Head()

	if tagRef != nil && *tagRef == head.Hash() {
		return latestVersion, nil
	}
	if bumpPatch && found {
		err = latestVersion.IncrementPatch()
		if err != nil {
			return nil, err
		}
	}

	if environment.IsPrerelease && bumpPatch {
		buildNumber, err := countCommitSinceTag(gitVersion.GetRepository(), tagRef)
		if err != nil {
			return nil, err
		}
		var prVersion semver.PRVersion
		if environment.PrereleaseTag == "" {
			preReleaseTag := strings.ReplaceAll(strings.ReplaceAll(gitVersion.GetBranchName(), "/", "-"), "_", "-")
			prVersion, err = semver.NewPRVersion(preReleaseTag)
			if err != nil {
				return nil, err
			}

		} else {
			prVersion, err = semver.NewPRVersion(environment.PrereleaseTag)
			if err != nil {
				return nil, err
			}
		}
		latestVersion.Pre = []semver.PRVersion{
			prVersion,
		}
		latestVersion.Build = []string{
			strconv.Itoa(buildNumber),
		}
	}
	return latestVersion, nil
}

func countCommitSinceTag(repo *git.Repository, tagRef *plumbing.Hash) (int, error) {
	head, err := repo.Head()
	if err != nil {
		return 0, err
	}

	if tagRef != nil && head.Hash() == *tagRef {
		return 0, nil
	}
	var tagCommits []plumbing.Hash
	if tagRef != nil {
		commits, err := repo.Log(&git.LogOptions{
			From: *tagRef,
		})
		if err != nil {
			return 0, err
		}
		err = commits.ForEach(func(commit *object.Commit) error {
			tagCommits = append(tagCommits, commit.Hash)
			return nil
		})
		if err != nil {
			return 0, err
		}
	}
	c := 0
	p := false
	commits, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return 0, err
	}
	err = commits.ForEach(func(commit *object.Commit) error {
		if uarray.Contains(tagCommits, commit.Hash) {
			p = true
		}
		if p == false {
			c = c + 1
		}
		return nil
	})
	return c, nil
}
