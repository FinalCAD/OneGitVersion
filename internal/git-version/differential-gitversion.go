package gitVersion

import (
	"DotnetGitHubVersion/internal/git-version/dotnet"
	"DotnetGitHubVersion/internal/git-version/wiki"
	"DotnetGitHubVersion/internal/utils/uarray"
	"bytes"
	"fmt"
	"github.com/go-git/go-git/v5"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DifferentialGitVersion struct {
	repo           *git.Repository
	wikiRepository *wiki.GitWiki
	service        *Service
	branchName     string
	repoPath       string
	parameters     Parameters
}
type projectPath struct {
	RelativeDirectory string
	Directory         string
	CsProj            string
}

var (
	excludeFolders []string = []string{"bin", "obj", ".git", ".circleci", ".idea"}
)

func (s *projectPath) Name() string {
	idx := strings.LastIndex(s.RelativeDirectory, "/")
	if idx == -1 {
		return s.RelativeDirectory
	}
	name := s.RelativeDirectory[idx+1:]
	return name
}

func (s *DifferentialGitVersion) GetRepository() *git.Repository {
	return s.repo
}

func (s *DifferentialGitVersion) GetService() *Service {
	return s.service
}

func (s *DifferentialGitVersion) GetBranchName() string {
	return s.branchName
}

func (s *DifferentialGitVersion) GetParameters() Parameters {
	return s.parameters
}

func (s *DifferentialGitVersion) GetWikiRepository() *wiki.GitWiki {
	if s.wikiRepository == nil {
		wikiRepo, err := wiki.NewGitWiki(s.repoPath, s.repo, s.parameters.AccessToken)
		if err != nil {
			panic(err)
		}
		s.wikiRepository = wikiRepo
	}
	return s.wikiRepository
}

func (s *DifferentialGitVersion) ApplyVersioning(environment *Environment) error {
	projectPaths, err := s.findProjFiles(filepath.Join(s.repoPath, s.service.Path), s.repoPath)
	if err != nil {
		return err
	}
	fmt.Printf("Find %d projects\n", len(projectPaths))
	changes, err := s.findGitChanges(s.repoPath)
	if err != nil {
		return err
	}
	projectChanges := s.findProjectWithChanges(projectPaths, changes)

	if len(projectChanges) == 0 {
		fmt.Printf("No project changes found\n")
	} else {
		fmt.Printf("%d project changes found\n", len(projectChanges))
	}
	for _, project := range projectPaths {
		bumpVersion := projectPathContains(projectChanges, project)
		err = s.versionProject(project.CsProj, project.Name(), environment, bumpVersion)
		if err != nil {
			return err
		}
	}

	if environment.AutoTag && !s.parameters.NoPush {
		err = s.wikiRepository.Push()
		if err != nil {
			return err
		}
	}
	return nil
}

func projectPathContains(projectPaths []projectPath, element projectPath) bool {
	for _, path := range projectPaths {
		if path.CsProj == element.CsProj {
			return true
		}
	}
	return false
}

func (s *DifferentialGitVersion) versionProject(csProjPath string, name string, environment *Environment, bumpVersion bool) error {
	version, err := createNewVersion(s, environment, name, bumpVersion)
	if err != nil {
		return err
	}
	if s.service.TargetType == TargetTypeDotnet {
		err = dotnet.SetVersionOnProject(*version, csProjPath)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Set version %s for project %s\n", version.String(), name)
	if environment.AutoTag {
		err = saveVersion(s, environment, name, *version, false)
		if err != nil {
			return nil
		}
	}
	return nil
}

func (s *DifferentialGitVersion) findProjectWithChanges(paths []projectPath, changes []string) []projectPath {
	var projectChanges []projectPath
	for _, path := range paths {
		if uarray.StartWith(changes, path.RelativeDirectory) {
			projectChanges = append(projectChanges, path)
		}
	}
	return projectChanges
}

func (s *DifferentialGitVersion) findGitChanges(repoPath string) ([]string, error) {
	head, err := s.repo.Head()
	if err != nil {
		return nil, err
	}
	commits, err := s.repo.Log(&git.LogOptions{
		From: head.Hash(),
	})
	if err != nil {
		return nil, err
	}
	currentCommit, err := commits.Next()
	if err != nil {
		return nil, err
	}
	previousCommit, err := commits.Next()
	if err != nil {
		return nil, err
	}
	gitDir := fmt.Sprintf("--git-dir=%s\\.git", repoPath)
	fmt.Printf("gitDir %s\n", gitDir)
	cmd := exec.Command("git", gitDir, "diff", "--name-only", currentCommit.Hash.String(), previousCommit.Hash.String())
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	outStr := out.String()
	lines := strings.Split(outStr, "\n")
	return lines, nil
}

func (s *DifferentialGitVersion) findProjFiles(seekingPath string, repositoryPath string) ([]projectPath, error) {
	elements, err := os.ReadDir(seekingPath)
	if err != nil {
		return nil, err
	}
	var results []projectPath
	for _, element := range elements {
		if !element.IsDir() && strings.HasSuffix(element.Name(), ".csproj") {
			relativePath := seekingPath[len(repositoryPath)+1:]
			if strings.Contains(relativePath, "\\") {
				relativePath = strings.ReplaceAll(relativePath, "\\", "/")
			}
			results = append(results, projectPath{
				RelativeDirectory: relativePath,
				Directory:         seekingPath,
				CsProj:            filepath.Join(seekingPath, element.Name()),
			})
		} else if element.IsDir() && !uarray.Contains(excludeFolders, element.Name()) {
			files, err := s.findProjFiles(filepath.Join(seekingPath, element.Name()), repositoryPath)
			if err != nil {
				return nil, err
			}
			results = append(results, files[:]...)
		}
	}
	return results, nil
}

func NewDifferentialGitVersion(repo *git.Repository, service *Service, branchName string, repoPath string, parameters Parameters) *DifferentialGitVersion {
	return &DifferentialGitVersion{
		repo:       repo,
		service:    service,
		branchName: branchName,
		parameters: parameters,
		repoPath:   repoPath,
	}
}
