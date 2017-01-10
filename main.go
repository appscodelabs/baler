package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/appscode/baler/cmd"
	v "github.com/appscode/go/version"
	"github.com/spf13/cobra"
)

var (
	Version         string
	VersionStrategy string
	Os              string
	Arch            string
	CommitHash      string
	GitBranch       string
	GitTag          string
	CommitTimestamp string
	BuildTimestamp  string
	BuildHost       string
	BuildHostOs     string
	BuildHostArch   string
)

func init() {
	v.Version.Version = Version
	v.Version.VersionStrategy = VersionStrategy
	v.Version.Os = Os
	v.Version.Arch = Arch
	v.Version.CommitHash = CommitHash
	v.Version.GitBranch = GitBranch
	v.Version.GitTag = GitTag
	v.Version.CommitTimestamp = CommitTimestamp
	v.Version.BuildTimestamp = BuildTimestamp
	v.Version.BuildHost = BuildHost
	v.Version.BuildHostOs = BuildHostOs
	v.Version.BuildHostArch = BuildHostArch
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "baler [command]",
		Short: `Save and Load a collection of Docker Images from a compact archive`,
		Run: func(c *cobra.Command, args []string) {
			c.Help()
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	rootCmd.AddCommand(cmd.NewCmdPack())
	rootCmd.AddCommand(cmd.NewCmdUnpack())
	rootCmd.AddCommand(cmd.NewCmdLoad())
	rootCmd.AddCommand(cmd.NewCmdRMI())
	rootCmd.AddCommand(v.NewCmdVersion())
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
