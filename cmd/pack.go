package cmd

import (
	"fmt"
	"os"

	"github.com/appscode/baler/baler"
	term "github.com/appscode/go-term"
	"github.com/spf13/cobra"
)

func NewCmdPack() *cobra.Command {
	var (
		destDir string
		cwd, _  = os.Getwd()
	)
	cmd := &cobra.Command{
		Use:   "pack MANIFEST_PATH",
		Short: "Create a baler archive from manifest",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("ERROR : Provide a manifest path")
				os.Exit(1)
			}
			err := baler.Pack(args[0], destDir)
			term.ExitOnError(err)
		},
	}
	cmd.Flags().StringVarP(&destDir, "dest-dir", "d", cwd, "Specify the location where baler archive will be stored(default: cwd).")
	return cmd
}
