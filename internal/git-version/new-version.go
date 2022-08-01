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

func getLastVersionFromTag(gitVersion GitVersion, name string, defaultVersion semver.Version) (*semver.Version, *object.Tag, error) {
	tags, err := gitVersion.GetRepository().TagObjects()
	if err != nil {
		return nil, nil, err
	}

	latestVersion := defaultVersion
	var tagRef *object.Tag
	if err != nil {
		return nil, nil, err
	}

	err = tags.ForEach(func(t *object.Tag) error {
		tagName := t.Name
		if strings.Index(tagName, name) == 0 {
			versionString := tagName[len(name)+1:]
			version, errs := semver.Make(versionString)
			if errs != nil {
				return errs
			}
			if version.GT(latestVersion) {
				latestVersion = version
				tagRef = t
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return &latestVersion, tagRef, nil
}

func getLastVersionFromWiki(gitVersion GitVersion, name string, environment *Environment, defaultVersion semver.Version) (*semver.Version, error) {
	var preReleaseName string
	if environment.IsPrerelease {
		preReleaseName = gitVersion.GetBranchName()
	}
	page, err := gitVersion.GetWikiRepository().ReadPage(name, gitVersion.GetService().TagName, preReleaseName)
	if err != nil {
		return nil, err
	}
	versions := page.GetVersions()
	if versions == nil {
		latestVersion := defaultVersion
		return &latestVersion, nil
	}
	latestVersion := defaultVersion
	for _, version := range versions {
		if version.GT(latestVersion) {
			latestVersion = version
		}
	}

	return &latestVersion, nil
}

func createNewVersion(gitVersion GitVersion, environment *Environment, name string, bumpPatch bool) (*semver.Version, error) {
	var latestVersion *semver.Version
	var tagRef *object.Tag
	var err error
	defaultVersion, err := semver.Make(gitVersion.GetService().Version)
	if err != nil {
		return nil, err
	}
	if gitVersion.GetService().TagType == TagTypeGit {
		latestVersion, tagRef, err = getLastVersionFromTag(gitVersion, name, defaultVersion)
	} else {
		latestVersion, err = getLastVersionFromWiki(gitVersion, name, environment, defaultVersion)
	}
	head, _ := gitVersion.GetRepository().Head()

	if tagRef != nil && tagRef.Target == head.Hash() {
		return latestVersion, nil
	}
	if bumpPatch && !latestVersion.Equals(defaultVersion) {
		err = latestVersion.IncrementPatch()
		if err != nil {
			return nil, err
		}
	}

	if environment.IsPrerelease {
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

func countCommitSinceTag(repo *git.Repository, tagRef *object.Tag) (int, error) {
	head, err := repo.Head()
	if err != nil {
		return 0, err
	}

	if tagRef != nil && head.Hash() == tagRef.Target {
		return 0, nil
	}
	var tagCommits []plumbing.Hash
	if tagRef != nil {
		commits, err := repo.Log(&git.LogOptions{
			From: tagRef.Target,
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
