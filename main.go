package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/appscode/baler/cmd"
	v "github.com/appscode/go/version"
	"github.com/spf13/cobra"
)

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
