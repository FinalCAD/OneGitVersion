package ugit

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"os/exec"
	"strings"
	"time"
)

func Tag(r *git.Repository, tag string) (bool, error) {
	if tagExists(tag, r) {
		log.Printf("tag %s already exists", tag)
		return false, nil
	}
	log.Printf("Set tag %s", tag)
	h, err := r.Head()
	if err != nil {
		log.Printf("get HEAD error: %s", err)
		return false, err
	}
	_, err = r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Message: tag,
		Tagger: &object.Signature{
			Name:  "BotMCS",
			Email: "maxime.charles@finalcad.com",
			When:  time.Now(),
		},
	})

	if err != nil {
		log.Printf("create tag error: %s", err)
		return false, err
	}

	return true, nil
}

func Push(r *git.Repository, accessToken string, tags bool) error {
	remoteUrl, err := GetRemoteUrl(r)
	if err != nil {
		return nil
	}
	var url string
	if strings.HasPrefix(remoteUrl, "https://") {
		tmp := remoteUrl[len("https://"):]
		url = fmt.Sprintf("https://%s@%s", accessToken, tmp)
	} else if strings.HasPrefix(remoteUrl, "http://") {
		tmp := remoteUrl[len("http://"):]
		url = fmt.Sprintf("http://%s@%s", accessToken, tmp)
	} else if strings.HasPrefix(remoteUrl, "git@") {
		tmp := strings.ReplaceAll(remoteUrl[len("git@"):], ":", "/")
		url = fmt.Sprintf("https://%s@%s", accessToken, tmp)
	} else {
		fmt.Printf("%s\n", remoteUrl)
		return errors.New(fmt.Sprintf("Unknown url %s", remoteUrl))
	}
	fmt.Printf("Push with url %s\n", url)
	args := []string{
		"push",
	}
	if tags {
		args = append(args, "--tags")
	}
	args = append(args, url)
	cmd := exec.Command("git", args[:]...)
	var out bytes.Buffer
	var outErr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &outErr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("git push %s\n", outErr.String())
		return err
	}
	fmt.Printf("git push %s\n", out.String())
	return nil
}

func tagExists(tag string, r *git.Repository) bool {
	tagFoundErr := "tag was found"
	tags, err := r.TagObjects()
	if err != nil {
		log.Printf("get tags error: %s", err)
		return false
	}
	res := false
	err = tags.ForEach(func(t *object.Tag) error {
		if t.Name == tag {
			res = true
			return fmt.Errorf(tagFoundErr)
		}
		return nil
	})
	if err != nil && err.Error() != tagFoundErr {
		log.Printf("iterate tags error: %s", err)
		return false
	}
	return res
}
