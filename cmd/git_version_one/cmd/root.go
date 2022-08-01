package cmd

import (
	"DotnetGitHubVersion/cmd/git_version_one/config"
	git_version "DotnetGitHubVersion/internal/git-version"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	path        string
	services    []string
	noPush      bool
	envPath     string
	accessToken string
)
var rootCmd = &cobra.Command{
	Use:   "GitVersionOne",
	Short: "Apply version to different libraries and applications in repository",
	Run: func(cmd *cobra.Command, args []string) {
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			fmt.Printf("Invalid path %s\n", path)
			return
		}
		fmt.Printf("%s\n", absolutePath)
		c, errc := config.ReadVersionFile(absolutePath, "version.yml")
		if errc != nil {
			fmt.Printf("Invalid path %s\n", path)
			fmt.Fprintln(os.Stderr, errc)
			return
		}
		err = run(services, c, absolutePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	},
}

func run(services []string, config *git_version.VersionConfig, repoPath string) error {
	for _, serviceName := range services {
		service := config.Services[serviceName]
		if service == nil {
			return errors.New(fmt.Sprintf("Missing service named %s", serviceName))
		}
		err := git_version.Apply(service, repoPath, git_version.Parameters{
			NoPush:      noPush,
			EnvPath:     envPath,
			AccessToken: accessToken,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&path, "path", ".", "The path of the repository")
	rootCmd.MarkFlagRequired("path")
	rootCmd.Flags().StringVar(&accessToken, "access-token", "", "The access token for git")
	rootCmd.MarkFlagRequired("access-token")
	rootCmd.Flags().StringArrayVar(&services, "service", []string{}, "Limit to this service")
	rootCmd.Flags().BoolVar(&noPush, "no-push", false, "Disable pushing tag into remote")
	rootCmd.Flags().StringVar(&envPath, "export-path", "", "Destination file to export bash environment variable")
}
