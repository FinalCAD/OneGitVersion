package gitVersion

import (
	"DotnetGitHubVersion/internal/git-version/ugit"
	"fmt"
	"github.com/blang/semver/v4"
)

func saveVersion(gitVersion GitVersion, environment *Environment, name string, version semver.Version, push bool) error {
	if gitVersion.GetService().TagType == TagTypeGit {
		return saveGitVersion(gitVersion, version, push)
	}
	return saveWikiVersion(gitVersion, environment, name, version, push)
}

func saveWikiVersion(gitVersion GitVersion, environment *Environment, name string, version semver.Version, push bool) error {
	var preReleaseName string
	if environment.IsPrerelease {
		preReleaseName = gitVersion.GetBranchName()
	}
	page, err := gitVersion.GetWikiRepository().ReadPage(name, gitVersion.GetService().TagName, preReleaseName)
	if err != nil {
		return err
	}
	head, err := gitVersion.GetRepository().Head()
	if err != nil {
		return err
	}
	page.AddVersion(version, head.Hash().String())
	err = page.Write()
	if err != nil {
		return err
	}
	wikiRepo := gitVersion.GetWikiRepository()
	err = page.GitAdd(wikiRepo)
	if push {
		err = wikiRepo.Push()
		if err != nil {
			return err
		}
	}
	return nil
}

func saveGitVersion(gitVersion GitVersion, version semver.Version, push bool) error {
	tag := fmt.Sprintf("%s-%s", gitVersion.GetService().TagName, version.String())
	created, err := ugit.Tag(gitVersion.GetRepository(), tag)
	if err != nil {
		return err
	}
	if created && push {
		err = ugit.Push(gitVersion.GetRepository())
		if err != nil {
			return err
		}
	}
	return nil
}
