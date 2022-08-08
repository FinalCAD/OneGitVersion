package dotnet

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
)

type projectXml struct {
	Sdk        string         `xml:"Sdk,attr"`
	ItemGroups []itemGroupXml `xml:"ItemGroup"`
}

type projectReferenceXml struct {
	Include string `xml:"Include,attr"`
}
type packageReferenceXml struct {
	Include string `xml:"Include,attr"`
	Version string `xml:"Version,attr"`
}
type itemGroupXml struct {
	ProjectReferences []projectReferenceXml `xml:"ProjectReference"`
	Packages          []packageReferenceXml `xml:"PackageReference"`
}

func readCsproj(path string) (Project, error) {
	byteValue, err := os.ReadFile(path)
	if err != nil {
		return Project{}, err
	}
	csproj := projectXml{}
	err = xml.Unmarshal(byteValue, &csproj)
	if err != nil {
		return Project{}, err
	}
	project := Project{
		Path:         path,
		Directory:    filepath.Dir(path),
		Sdk:          csproj.Sdk,
		Dependencies: Dependencies{},
	}
	for _, itemGroup := range csproj.ItemGroups {
		if itemGroup.ProjectReferences != nil {
			for _, reference := range itemGroup.ProjectReferences {
				include := reference.Include
				if strings.Contains(include, "\\") {
					include = strings.ReplaceAll(reference.Include, "\\", "/")
				}
				depPath := filepath.Join(project.Directory, include)
				depProject, err := readCsproj(depPath)
				if err != nil {
					continue
				}
				project.Dependencies.Projects = append(project.Dependencies.Projects, depProject)
			}
		}
		if itemGroup.Packages != nil {
			for _, packageRef := range itemGroup.Packages {
				project.Dependencies.Packages = append(project.Dependencies.Packages, Packages{
					Include: packageRef.Include,
					Version: packageRef.Version,
				})
			}
		}
	}
	return project, nil
}
