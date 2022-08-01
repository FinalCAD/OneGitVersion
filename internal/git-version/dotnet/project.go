package dotnet

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
