package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(hashObjectCmd)
}

var hashObjectCmd = &cobra.Command{
	Use:   "hash-object",
	Short: "a very attempt at converting a file into a git object",
	Long:  "a very very bad attempt at converting a file into a git object from scratch",
	Args:  cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
	RunE:  runHashObject,
}

func runHashObject(cmd *cobra.Command, args []string) error {
	return nil
}
