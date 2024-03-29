package wiki

import (
	"DotnetGitHubVersion/internal/git-version/ugit"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"os"
	"path/filepath"
	"strings"
)

type GitWiki struct {
	repo        *git.Repository
	path        string
	accessToken string
}

func (s *GitWiki) Push() error {
	return s.repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "go",
			Password: s.accessToken,
		},
	})
}
func (s *GitWiki) ReadPage(name string, appName string) (*VersionPage, error) {
	pageDir := filepath.Join(s.path, appName)
	err := assertDirectoryExists(pageDir)
	if err != nil {
		return nil, err
	}
	return NewVersionPage(name, pageDir, s.path)
}

func assertDirectoryExists(path string) error {
	e, err := directoryExists(path)
	if err != nil {
		return err
	}
	if e {
		return nil
	}
	err = os.Mkdir(path, 0755)
	return err
}

func directoryExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func NewGitWiki(repositoryPath string, repo *git.Repository, accessToken string) (*GitWiki, error) {
	wikiPath := filepath.Join(repositoryPath, "wiki")
	_, err := os.Stat(wikiPath)
	if err == nil {
		r, err := git.PlainOpen(wikiPath)
		return &GitWiki{
			repo: r,
			path: wikiPath,
		}, err
	}
	remoteUrl, err := ugit.GetRemoteUrl(repo)
	if err != nil {
		return nil, err
	}
	if !(strings.HasPrefix(remoteUrl, "https://github.com/") || strings.HasPrefix(remoteUrl, "git@github.com")) {
		return nil, errors.New("only github supported")
	}
	if strings.HasPrefix(remoteUrl, "git@github.com") {
		remoteUrl = "https://github.com/" + remoteUrl[len("git@github.com:"):]
		fmt.Printf("Replace ssh version to https: %s\n", remoteUrl)
	}
	if !os.IsNotExist(err) {
		remoteUrl = remoteUrl[:len(remoteUrl)-4]
		remoteUrl = remoteUrl + ".wiki.git"
		r, err := git.PlainClone(wikiPath, false, &git.CloneOptions{
			URL: remoteUrl,
			Auth: &http.BasicAuth{
				Username: "go",
				Password: accessToken,
			},
		})
		return &GitWiki{
			repo:        r,
			path:        wikiPath,
			accessToken: accessToken,
		}, err
	}
	return nil, err
}
