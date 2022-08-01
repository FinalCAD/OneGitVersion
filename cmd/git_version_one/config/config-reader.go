package config

import (
	gitVersion "DotnetGitHubVersion/internal/git-version"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func ReadVersionFile(path string, filename string) (*gitVersion.VersionConfig, error) {
	r := versionConfigModel{}
	filePath := filepath.Join(path, filename)
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(fileContent, &r)
	if err != nil {
		return nil, err
	}
	dest := gitVersion.VersionConfig{}
	err = setDefaultValue(&r, &dest)
	if err != nil {
		return nil, err
	}
	return &dest, nil
}

func setDefaultValue(config *versionConfigModel, dest *gitVersion.VersionConfig) error {
	if dest.Services == nil {
		dest.Services = make(map[string]*gitVersion.Service)
	}
	if config.Version == "1.0" {
		return setDefaultValueV1(config, dest)
	} else {
		return errors.New("invalid version")
	}
}

func setDefaultValueV1(config *versionConfigModel, dest *gitVersion.VersionConfig) error {
	for name, source := range config.Services {
		service := gitVersion.Service{
			Version:       "",
			DefaultConfig: gitVersion.DefaultConfigMain,
			Environments:  map[string]*gitVersion.Environment{},
			Path:          "",
			TargetType:    gitVersion.TargetTypeDotnet,
			TagName:       name,
			VersionType:   gitVersion.VersionTypeGlobal,
			TagType:       gitVersion.TagTypeGit,
		}
		if source.DefaultConfig != "" {
			service.DefaultConfig = gitVersion.DefaultConfig(source.DefaultConfig)
		}
		if source.TagName != "" {
			service.TagName = source.TagName
		}
		if source.Path == "" {
			return errors.New(fmt.Sprintf("Missing path for service %s", name))
		}
		if source.Version == "" {
			return errors.New(fmt.Sprintf("Missing version for service %s", name))
		}
		if source.TagType != "" {
			service.TagType = gitVersion.TagType(source.TagType)
		}
		if source.VersionType != "" {
			service.VersionType = gitVersion.VersionType(source.VersionType)
		}
		service.Path = source.Path
		service.Version = source.Version
		if source.TargetType != "" {
			service.TargetType = gitVersion.TargetType(source.TargetType)
		}
		if source.DefaultConfig == string(gitVersion.DefaultConfigMain) {
			setDefaultMainEnvironment(source, &service)
		} else if source.DefaultConfig == string(gitVersion.DefaultConfigDevelop) {
			setDefaultDevelopEnvironment(source, &service)
		} else {
			return errors.New(fmt.Sprintf("Invalid default-service for service %s", name))
		}
		dest.Services[name] = &service
	}
	return nil
}

func setDefaultDevelopEnvironment(source *serviceModel, service *gitVersion.Service) {
	mergeEnvironment(source, service, "main", &gitVersion.Environment{
		IsPrerelease:  false,
		PrereleaseTag: "",
		AutoTag:       true,
		Branches: gitVersion.BranchFilter{
			Only:   []string{"develop"},
			Ignore: nil,
		},
	})
	mergeEnvironment(source, service, "others", &gitVersion.Environment{
		IsPrerelease:  true,
		AutoTag:       false,
		PrereleaseTag: "",
		Branches: gitVersion.BranchFilter{
			Only:   nil,
			Ignore: []string{"develop"},
		},
	})
}

func setDefaultMainEnvironment(source *serviceModel, service *gitVersion.Service) {
	mergeEnvironment(source, service, "main", &gitVersion.Environment{
		IsPrerelease:  false,
		PrereleaseTag: "",
		AutoTag:       true,
		Branches: gitVersion.BranchFilter{
			Only:   []string{"main", "master"},
			Ignore: nil,
		},
	})
	mergeEnvironment(source, service, "staging", &gitVersion.Environment{
		IsPrerelease:  true,
		PrereleaseTag: "beta",
		AutoTag:       false,
		Branches: gitVersion.BranchFilter{
			Only:   []string{"staging"},
			Ignore: nil,
		},
	})
	mergeEnvironment(source, service, "develop", &gitVersion.Environment{
		IsPrerelease:  true,
		PrereleaseTag: "alpha",
		AutoTag:       false,
		Branches: gitVersion.BranchFilter{
			Only:   []string{"develop"},
			Ignore: nil,
		},
	})
	mergeEnvironment(source, service, "others", &gitVersion.Environment{
		IsPrerelease: true,
		AutoTag:      false,
		Branches: gitVersion.BranchFilter{
			Only:   nil,
			Ignore: []string{"main", "master", "staging", "develop"},
		},
	})
}

func mergeEnvironment(source *serviceModel, service *gitVersion.Service, name string, defaultEnv *gitVersion.Environment) {
	s := source.Environments[name]
	var d gitVersion.Environment
	if s == nil {
		s = &environmentModel{}
	}
	d.IsPrerelease = useDefault(s.IsPrerelease, defaultEnv.IsPrerelease)
	d.PrereleaseTag = useDefault(s.PrereleaseTag, defaultEnv.PrereleaseTag)
	d.AutoTag = useDefault(s.AutoTag, defaultEnv.AutoTag)
	if s.Branches != nil {
		if s.Branches.Only != nil {
			d.Branches.Only = s.Branches.Only
		} else {
			d.Branches.Only = defaultEnv.Branches.Only
		}
		if s.Branches.Ignore != nil {
			d.Branches.Ignore = s.Branches.Ignore
		} else {
			d.Branches.Ignore = defaultEnv.Branches.Ignore
		}
	} else {
		d.Branches.Only = defaultEnv.Branches.Only
		d.Branches.Ignore = defaultEnv.Branches.Ignore
	}
	service.Environments[name] = &d
}

func useDefault[K interface{}](value *K, defaultValue K) K {
	if value == nil {
		return defaultValue
	}
	return *value
}
