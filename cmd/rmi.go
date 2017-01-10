package cmd

import (
	"fmt"
	"os"

	"github.com/appscode/baler/baler"
	"github.com/spf13/cobra"
)

func NewCmdRMI() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rmi MANIFEST_PATH",
		Short: "Remove images specified in manifest",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("ERROR : Provide a manifest path")
				os.Exit(1)
			}
			baler.RemoveImages(args[0])
		},
	}
	return cmd
}
