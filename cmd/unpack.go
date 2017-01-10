package cmd

import (
	"fmt"
	"os"

	"github.com/appscode/baler/baler"
	term "github.com/appscode/go-term"
	"github.com/spf13/cobra"
)

func NewCmdUnpack() *cobra.Command {
	var (
		destDir string
		cwd, _  = os.Getwd()
	)
	cmd := &cobra.Command{
		Use:   "unpack ARCHIVE_PATH",
		Short: "Unpack a baler archive into a directory",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("ERROR : Provide a CharName")
				os.Exit(1)
			}
			err := baler.Unpack(args[0], destDir)
			term.ExitOnError(err)
		},
	}
	cmd.Flags().StringVarP(&destDir, "dest-dir", "d", cwd, "Specify the location where baler package will be unpacked(default: cwd).")
	return cmd
}
