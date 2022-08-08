package dotnet

import "strings"

type Project struct {
	Path         string
	Directory    string
	Sdk          string
	Dependencies Dependencies
}

type Dependencies struct {
	Packages []Packages
	Projects []Project
}

type Packages struct {
	Include string
	Version string
}

var (
	pathToProject map[string]Project = make(map[string]Project)
)

func openProject(path string) (Project, error) {
	var csprojPath string
	if strings.HasSuffix(path, ".csproj") {
		csprojPath = path
	} else {
		csPath, err := findCsprojFile(path)
		if err != nil {
			return Project{}, err
		}
		csprojPath = csPath
	}
	existingProject, exists := pathToProject[csprojPath]
	if exists {
		return existingProject, nil
	}
	csproj, err := readCsproj(csprojPath)
	if err != nil {
		return Project{}, err
	}
	pathToProject[csprojPath] = csproj
	return csproj, nil
}
