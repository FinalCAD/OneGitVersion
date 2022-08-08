package gitVersion

import (
	"DotnetGitHubVersion/internal/git-version/ugit"
	"fmt"
	"github.com/blang/semver/v4"
)

func saveVersion(gitVersion GitVersion, environment *Environment, name string, version semver.Version, push bool) error {
	fmt.Printf("Auto tag with %s rule\n", gitVersion.GetService().TagType)
	if gitVersion.GetService().TagType == TagTypeGit {
		return saveGitVersion(gitVersion, version, push)
	}
	return saveWikiVersion(gitVersion, name, version, push)
}

func saveWikiVersion(gitVersion GitVersion, name string, version semver.Version, push bool) error {
	wikiRepo := gitVersion.GetWikiRepository()
	page, err := wikiRepo.ReadPage(name, gitVersion.GetService().TagName)
	if err != nil {
		return err
	}
	head, err := gitVersion.GetRepository().Head()
	if err != nil {
		return err
	}
	added := page.AddVersion(version, head.Hash().String())
	if !added {
		return nil
	}
	err = page.Write()
	if err != nil {
		return err
	}
	err = page.GitAdd(wikiRepo)
	if err != nil {
		return err
	}
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
	fmt.Printf("Set tag on git %s\n", tag)
	created, err := ugit.Tag(gitVersion.GetRepository(), tag)
	if err != nil {
		return err
	}
	if created && push {
		err = ugit.Push(gitVersion.GetRepository(), gitVersion.GetParameters().AccessToken, true)
		if err != nil {
			return err
		}
	}
	return nil
}
