package wiki

import (
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"os"
	"path/filepath"
	"time"
)

type VersionPage struct {
	filePath       string
	path           string
	name           string
	preReleaseName string
	versions       []VersionLine
}

type VersionLine struct {
	Version semver.Version
	Commit  string
}

func NewVersionPage(name string, preReleaseName string, pageDir string) (*VersionPage, error) {
	fileName := getOrCreatePageFileName(name, preReleaseName, pageDir)
	v, err := readVersionPage(fileName)
	if err != nil {
		return nil, err
	}
	return &VersionPage{
		path:           pageDir,
		name:           name,
		preReleaseName: preReleaseName,
		versions:       v,
		filePath:       fileName,
	}, nil
}

func getOrCreatePageFileName(name string, preReleaseName string, pageDir string) string {
	var fileName string
	if preReleaseName == "" {
		fileName = fmt.Sprintf("%s.md", name)
	} else {
		fileName = fmt.Sprintf("%s-%s.md", name, preReleaseName)
	}
	return filepath.Join(pageDir, fileName)
}

func (s *VersionPage) GetVersions() []semver.Version {
	return nil
}

func (s *VersionPage) AddVersion(version semver.Version, commitHash string) {
	c := false
	for _, versionLine := range s.versions {
		if versionLine.Version.Equals(version) {
			c = true
		}
	}
	if !c {
		s.versions = append(s.versions, VersionLine{
			Version: version,
			Commit:  commitHash,
		})
	}
}

func (s *VersionPage) Write() error {
	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	_, err = file.WriteString(fmt.Sprintf("# %s - %s\n\n", s.name, s.preReleaseName))
	if err != nil {
		return err
	}

	for _, versionLine := range s.versions {
		_, err = file.WriteString(fmt.Sprintf("- %s (%s)\n", versionLine.Version.String(), versionLine.Commit))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *VersionPage) GitAdd(gitWiki *GitWiki) error {
	repo := gitWiki.repo
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(s.filePath)
	if err != nil {
		return err
	}
	_, err = w.Commit(fmt.Sprintf("Update %s", s.name), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "BotMCS",
			Email: "maxime.charles@finalcad.com",
			When:  time.Now(),
		},
	})
	return err
}
