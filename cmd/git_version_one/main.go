package main

import (
	"DotnetGitHubVersion/cmd/git_version_one/cmd"
	"DotnetGitHubVersion/cmd/git_version_one/versioning"
)

func main() {
	cmd.Execute(versioning.Version)
}
