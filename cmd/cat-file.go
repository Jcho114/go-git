package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(catFileCmd)
}

var catFileCmd = &cobra.Command{
	Use:   "cat-file",
	Short: "a very attempt at printing a git object to stdout",
	Long:  "a very very bad attempt at printing a git object to stdout from scratch",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:  runCatFile,
}

func runCatFile(cmd *cobra.Command, args []string) error {
	return nil
}
