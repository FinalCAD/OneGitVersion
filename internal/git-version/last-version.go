package gitVersion

import (
	"github.com/blang/semver/v4"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"strings"
)

func getLastVersionFromTag(gitVersion GitVersion, name string, defaultVersion semver.Version) (*semver.Version, *plumbing.Hash, bool, error) {
	tags, err := gitVersion.GetRepository().TagObjects()
	if err != nil {
		return nil, nil, false, err
	}

	latestVersion := defaultVersion
	var tagRef *object.Tag
	if err != nil {
		return nil, nil, false, err
	}

	err = tags.ForEach(func(t *object.Tag) error {
		tagName := t.Name
		if strings.Index(tagName, name) == 0 {
			versionString := tagName[len(name)+1:]
			version, errs := semver.Make(versionString)
			if errs != nil {
				return errs
			}
			if version.GTE(latestVersion) {
				latestVersion = version
				tagRef = t
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, false, err
	}
	if tagRef == nil {
		return &latestVersion, nil, false, nil
	}
	return &latestVersion, &tagRef.Target, tagRef != nil, nil
}

func getLastVersionFromWiki(gitVersion GitVersion, name string, environment *Environment, defaultVersion semver.Version) (*semver.Version, *plumbing.Hash, bool, error) {
	page, err := gitVersion.GetWikiRepository().ReadPage(name, gitVersion.GetService().TagName)
	if err != nil {
		return nil, nil, false, err
	}
	versions := page.GetVersions()
	if versions == nil {
		latestVersion := defaultVersion
		return &latestVersion, nil, false, nil
	}
	latestVersion := defaultVersion
	var hash *plumbing.Hash
	found := false
	for _, versionLine := range versions {
		if versionLine.Version.GTE(latestVersion) {
			latestVersion = versionLine.Version
			h := plumbing.NewHash(versionLine.Commit)
			hash = &h
			found = true
		}
	}

	return &latestVersion, hash, found, nil
}

func getLastVersion(gitVersion GitVersion, environment *Environment, name string, defaultVersion semver.Version) (*semver.Version, *plumbing.Hash, bool, error) {
	if gitVersion.GetService().TagType == TagTypeGit {
		return getLastVersionFromTag(gitVersion, name, defaultVersion)
	} else {
		return getLastVersionFromWiki(gitVersion, name, environment, defaultVersion)
	}
}
