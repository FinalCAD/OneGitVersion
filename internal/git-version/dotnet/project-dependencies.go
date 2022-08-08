package dotnet

import "DotnetGitHubVersion/internal/git-version/common"

func GetDependencies(path string) ([]common.Dependency, error) {
	project, err := openProject(path)
	if err != nil {
		return nil, err
	}

	deps := getDependenciesFlat(project)
	return deps, nil
}

func getDependenciesFlat(project Project) []common.Dependency {
	var dependencies []common.Dependency
	for _, dep := range project.Dependencies.Projects {
		dependencies = append(dependencies, common.Dependency{
			Name: dep.Path,
		})
		if dep.Dependencies.Projects != nil {
			subDeps := getDependenciesFlat(dep)
			dependencies = append(dependencies, subDeps[:]...)
		}
	}
	return dependencies
}
