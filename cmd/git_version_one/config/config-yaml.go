package config

type versionConfigModel struct {
	Version  string                   `yaml:"version"`
	Services map[string]*serviceModel `yaml:"services"`
}

type serviceModel struct {
	Version        string                       `yaml:"version"`
	DefaultConfig  string                       `yaml:"default-config"`
	Environments   map[string]*environmentModel `yaml:"environments"`
	SetEnvironment bool                         `yaml:"set-environment"`
	Path           string                       `yaml:"path"`
	TargetType     string                       `yaml:"target-type"`
	TagName        string                       `yaml:"tag-name"`
	VersionType    string                       `yaml:"version-type"`
	TagType        string                       `yaml:"tag-type"`
}

type environmentModel struct {
	IsPrerelease  *bool              `yaml:"is-prerelease"`
	PrereleaseTag *string            `yaml:"prerelease-tag"`
	AutoTag       *bool              `yaml:"auto-tag"`
	Branches      *branchFilterModel `yaml:"branches"`
}

type branchFilterModel struct {
	Only   []string `yaml:"only"`
	Ignore []string `yaml:"ignore"`
}
