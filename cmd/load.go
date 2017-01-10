package cmd

import (
	"fmt"
	"os"

	"github.com/appscode/baler/baler"
	"github.com/spf13/cobra"
)

func NewCmdLoad() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load ARCHIVE_PATH",
		Short: "Load Docker images from a baler archive",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("ERROR : Provide a archive path")
				os.Exit(1)
			}
			baler.Load(args[0])
		},
	}
	return cmd
}
