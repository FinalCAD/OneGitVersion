package golang

import (
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
	"os"
	"path/filepath"
)

func SetVersion(version semver.Version, path string) error {
	versionDir := filepath.Join(path, "versioning")
	err := assertCompatibility(versionDir, "version.go")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(versionDir, "version.go"), os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("package versioning\n\n")
	if err != nil {
		return err
	}
	_, err = file.WriteString(fmt.Sprintf("const Version = \"%s\"\n", version.String()))
	if err != nil {
		return err
	}
	return nil
}

func assertCompatibility(path string, fileName string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("Project not supported")
		}
		return err
	}
	_, err = os.Stat(filepath.Join(path, fileName))
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("Project not supported")
		}
		return err
	}
	return nil
}
