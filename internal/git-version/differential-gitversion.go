package gitVersion

import (
	"DotnetGitHubVersion/internal/git-version/common"
	"DotnetGitHubVersion/internal/git-version/dotnet"
	"DotnetGitHubVersion/internal/git-version/wiki"
	"DotnetGitHubVersion/internal/utils/uarray"
	"bytes"
	"fmt"
	"github.com/blang/semver/v4"
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
	excludeFolders = []string{"bin", "obj", ".git", ".circleci", ".idea"}
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
	projectChanges, err := s.findProjectWithChanges(projectPaths, environment)
	if err != nil {
		return err
	}
	if len(projectChanges) == 0 {
		fmt.Printf("No project changes found\n")
	} else {
		projectChanges, err = s.dependencyCheckUpgrade(projectPaths, projectChanges)
		if err != nil {
			return err
		}
		fmt.Printf("%d/%d project changes found\n", len(projectChanges), len(projectPaths))
	}

	for _, project := range projectPaths {
		bumpVersion := projectPathContains(projectChanges, project.CsProj)
		err = s.versionProject(project.CsProj, project.Name(), environment, bumpVersion)
		if err != nil {
			return err
		}
	}

	err = s.writeProjectChangesIntoFile(projectChanges)
	if err != nil {
		return err
	}
	if environment.AutoTag && !s.parameters.NoPush {
		err = s.wikiRepository.Push()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *DifferentialGitVersion) writeProjectChangesIntoFile(projectChanges []projectPath) error {
	if len(projectChanges) == 0 {
		return nil
	}
	path := filepath.Join(s.repoPath, "project-changes.log")
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, changes := range projectChanges {
		_, err = file.WriteString(changes.CsProj + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *DifferentialGitVersion) dependencyCheckUpgrade(allProjects []projectPath, projectChanges []projectPath) ([]projectPath, error) {
	change := false

	for {
		change = false
		for _, p := range allProjects {
			if projectPathContains(projectChanges, p.CsProj) {
				continue
			}
			var deps []common.Dependency
			var err error
			if s.service.TargetType == TargetTypeDotnet {
				deps, err = dotnet.GetDependencies(p.CsProj)
			}
			if err != nil {
				return nil, err
			}
			for _, dep := range deps {
				if projectPathContains(projectChanges, dep.Name) {
					projectChanges = append(projectChanges, p)
					change = true
					break
				}
			}
			if change {
				break
			}
		}
		if !change {
			break
		}
	}
	return projectChanges, nil
}

func projectPathContains(projectPaths []projectPath, element string) bool {
	for _, path := range projectPaths {
		if path.CsProj == element {
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
			return err
		}
	}
	return nil
}

func (s *DifferentialGitVersion) findProjectWithChanges(paths []projectPath, environment *Environment) ([]projectPath, error) {
	var projectChanges []projectPath
	for _, path := range paths {
		changes, err := s.findGitChanges(path, environment)
		if err != nil {
			return nil, err
		}
		if changes == nil || uarray.StartWith(changes, path.RelativeDirectory) {
			projectChanges = append(projectChanges, path)
		}
	}
	return projectChanges, nil
}

func (s *DifferentialGitVersion) findGitChanges(path projectPath, environment *Environment) ([]string, error) {
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

	_, hash, found, err := getLastVersion(s, environment, path.Name(), semver.Version{
		Major: 0,
		Minor: 0,
		Patch: 0,
		Pre:   nil,
		Build: nil,
	})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	previousCommit := hash
	gitDir := fmt.Sprintf("--git-dir=%s", filepath.Join(s.repoPath, ".git"))
	cmd := exec.Command("git", gitDir, "diff", "--name-only", currentCommit.Hash.String(), previousCommit.String())
	var out bytes.Buffer
	var outErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &outErr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error output: %s", outErr.String())
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
	// todo move to package dotnet
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
