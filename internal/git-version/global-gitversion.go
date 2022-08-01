package gitVersion

import (
	"DotnetGitHubVersion/internal/git-version/dotnet"
	"DotnetGitHubVersion/internal/git-version/export"
	"DotnetGitHubVersion/internal/git-version/golang"
	"DotnetGitHubVersion/internal/git-version/wiki"
	"fmt"
	"github.com/go-git/go-git/v5"
	"path/filepath"
)

type GlobalGitVersion struct {
	repo           *git.Repository
	wikiRepository *wiki.GitWiki
	service        *Service
	branchName     string
	parameters     Parameters
	repoPath       string
}

func (s *GlobalGitVersion) GetRepository() *git.Repository {
	return s.repo
}

func (s *GlobalGitVersion) GetService() *Service {
	return s.service
}

func (s *GlobalGitVersion) GetBranchName() string {
	return s.branchName
}

func (s *GlobalGitVersion) GetParameters() Parameters {
	return s.parameters
}

func (s *GlobalGitVersion) GetWikiRepository() *wiki.GitWiki {
	if s.wikiRepository == nil {
		wikiRepo, err := wiki.NewGitWiki(s.repoPath, s.repo, s.parameters.AccessToken)
		if err != nil {
			panic(err)
		}
		s.wikiRepository = wikiRepo
	}
	return s.wikiRepository
}

func (s *GlobalGitVersion) ApplyVersioning(environment *Environment) error {
	newVersion, err := createNewVersion(s, environment, s.service.TagName, true)
	if err != nil {
		return err
	}
	fmt.Printf("New version %s\n", newVersion.String())
	if newVersion != nil {
		fmt.Printf("New version %s\n", newVersion.String())
	}

	fmt.Printf("Target type %s\n", s.service.TargetType)
	if s.service.TargetType == TargetTypeDotnet {
		err = dotnet.SetVersionOnProject(*newVersion, filepath.Join(s.repoPath, s.service.Path))
		if err != nil {
			return err
		}
	} else if s.service.TargetType == TargetTypeGolang {
		err = golang.SetVersion(*newVersion, filepath.Join(s.repoPath, s.service.Path))
		if err != nil {
			return err
		}
	}
	if environment.AutoTag {
		err = saveVersion(s, environment, s.service.TagName, *newVersion, !s.parameters.NoPush)
		if err != nil {
			return nil
		}
	}
	if s.parameters.EnvPath != "" {
		export.ExportEnv(s.parameters.EnvPath, *newVersion)
	}
	return nil
}

func NewGlobalGitVersion(repo *git.Repository, service *Service, branchName string, repoPath string, parameters Parameters) *GlobalGitVersion {
	return &GlobalGitVersion{
		repo:       repo,
		service:    service,
		branchName: branchName,
		parameters: parameters,
		repoPath:   repoPath,
	}
}
