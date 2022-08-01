package ugit

import (
	"errors"
	"github.com/go-git/go-git/v5"
)

func GetRemoteUrl(repo *git.Repository) (string, error) {
	remotes, err := repo.Remotes()
	if err != nil {
		return "", err
	}
	if len(remotes) <= 0 {
		return "", errors.New("no remote found")
	}
	remote := remotes[0]
	remoteUrl := remote.Config().URLs[0]
	return remoteUrl, nil
}
