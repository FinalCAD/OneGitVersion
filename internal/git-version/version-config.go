package gitVersion

type DefaultConfig string
type TargetType string
type VersionType string
type TagType string

const (
	DefaultConfigMain    DefaultConfig = "main"
	DefaultConfigDevelop DefaultConfig = "develop"
)

const (
	VersionTypeGlobal       VersionType = "global"
	VersionTypeDifferential VersionType = "differential"
)

const (
	TargetTypeDotnet TargetType = "dotnet"
	TargetTypeDocker TargetType = "docker"
)

const (
	TagTypeGit  TagType = "git"
	TagTypeWiki TagType = "wiki"
)

type VersionConfig struct {
	Version  string              `yaml:"version"`
	Services map[string]*Service `yaml:"services"`
}

type Service struct {
	Version       string                  `yaml:"version"`
	DefaultConfig DefaultConfig           `yaml:"default-config"`
	Environments  map[string]*Environment `yaml:"environments"`
	Path          string                  `yaml:"path"`
	TargetType    TargetType              `yaml:"target-type"`
	TagName       string                  `yaml:"tag-name"`
	VersionType   VersionType             `yaml:"version-type"`
	TagType       TagType
}

type Environment struct {
	IsPrerelease  bool
	PrereleaseTag string
	AutoTag       bool
	Branches      BranchFilter
}

type BranchFilter struct {
	Only   []string `yaml:"only"`
	Ignore []string `yaml:"ignore"`
}

func (s *Environment) MatchBranch(branchName string) bool {
	if s.Branches.Ignore != nil {
		for _, branch := range s.Branches.Ignore {
			if branch == branchName {
				return false
			}
		}
		return true
	}
	if s.Branches.Only != nil {
		for _, branch := range s.Branches.Only {
			if branch == branchName {
				return true
			}
		}
	}
	return false
}
