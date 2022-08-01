package dotnet

import (
	"errors"
	"github.com/blang/semver/v4"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func SetVersionOnProject(version semver.Version, projectDir string) error {
	var csprojPath string
	if strings.HasSuffix(projectDir, ".csproj") {
		csprojPath = projectDir
	} else {
		path, err := findCsprojFile(projectDir)
		if err != nil {
			return err
		}
		csprojPath = path
	}
	csproj, err := readCsproj(csprojPath)
	if err != nil {
		return err
	}

	err = csproj.setVersion(version, false)
	if err != nil {
		return err
	}
	return nil
}

func findCsprojFile(path string) (string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".csproj") {
			return filepath.Join(path, file.Name()), nil
		}
	}
	return "", errors.New("File not found")
}
