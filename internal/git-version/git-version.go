package gitVersion

import (
	"DotnetGitHubVersion/internal/git-version/wiki"
	"errors"
	"fmt"
	git "github.com/go-git/go-git/v5"
)

type GitVersion interface {
	ApplyVersioning(environment *Environment) error
	GetRepository() *git.Repository
	GetService() *Service
	GetBranchName() string
	GetParameters() Parameters
	GetWikiRepository() *wiki.GitWiki
}

type Parameters struct {
	NoPush      bool
	EnvPath     string
	AccessToken string
}

func Apply(service *Service, repoPath string, parameters Parameters) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}
	branchName, err := getBranchName(repo)
	if err != nil {
		return err
	}
	env, err := getEnv(service, branchName)
	if err != nil {
		return err
	}

	var gitVersioning GitVersion
	if service.VersionType == VersionTypeGlobal {
		gitVersioning = NewGlobalGitVersion(repo, service, branchName, repoPath, parameters)
	} else if service.VersionType == VersionTypeDifferential {
		gitVersioning = NewDifferentialGitVersion(repo, service, branchName, repoPath, parameters)
	}
	return gitVersioning.ApplyVersioning(env)
}

func getBranchName(repo *git.Repository) (string, error) {
	headRef, err := repo.Head()
	if err != nil {
		return "", err
	}
	branchName := headRef.Name().Short()
	return branchName, nil
}

func getEnv(service *Service, branchName string) (*Environment, error) {
	var results []*Environment

	for _, environment := range service.Environments {
		if environment.MatchBranch(branchName) {
			results = append(results, environment)
		}
	}

	if len(results) == 0 {
		return nil, errors.New(fmt.Sprintf("No matching environment for branch %s", branchName))
	}
	if len(results) > 1 {
		return nil, errors.New(fmt.Sprintf("Multiple matching environments for branch %s", branchName))
	}
	return results[0], nil
}
