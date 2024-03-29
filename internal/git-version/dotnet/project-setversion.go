package dotnet

import (
	"encoding/xml"
	"fmt"
	"github.com/blang/semver/v4"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type importXml struct {
	Project string `xml:"Project,attr"`
}
type projectGroupXml struct {
	Version string `xml:"Version"`
}
type directoryBuildXml struct {
	XMLName        xml.Name          `xml:"Project"`
	Import         importXml         `xml:"Import"`
	PropertyGroups []projectGroupXml `xml:"PropertyGroup"`
}

func (s *Project) setVersion(version semver.Version, setChildren bool) error {
	directoryBuildFileName := "Directory.Build.props"
	fmt.Printf("Create %s in %s\n", directoryBuildFileName, s.Directory)
	configFilePath := filepath.Join(s.Directory, directoryBuildFileName)
	file, err := os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	var directoryBuild directoryBuildXml
	if err == nil {
		fileContent, err := ioutil.ReadAll(file)
		if err != nil {
			return nil
		}
		err = xml.Unmarshal(fileContent, &directoryBuild)
		if err != nil {
			if err != io.EOF {
				return err
			}
		}
	}
	nugetVersion := strings.ReplaceAll(version.String(), "+", ".")
	var existingVersion *projectGroupXml
	for _, propertyGroup := range directoryBuild.PropertyGroups {
		if propertyGroup.Version != "" {
			existingVersion = &propertyGroup
			existingVersion.Version = nugetVersion
			break
		}
	}
	if existingVersion == nil {
		n := projectGroupXml{
			Version: nugetVersion,
		}
		directoryBuild.PropertyGroups = append(directoryBuild.PropertyGroups, n)
	}
	if directoryBuild.Import.Project == "" {
		directoryBuild.Import.Project = "$([MSBuild]::GetPathOfFileAbove('build.props'))"
	}
	fileContent, err := xml.MarshalIndent(directoryBuild, "", "    ")
	if err != nil {
		return err
	}
	err = file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = file.Write(fileContent)
	if err != nil {
		return err
	}
	defer file.Close()
	if setChildren {
		for _, project := range s.Dependencies.Projects {
			err = project.setVersion(version, true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
