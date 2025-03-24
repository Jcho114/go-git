package cmd

import (
	"github.com/Jcho114/go-git/ref"
	"github.com/Jcho114/go-git/repo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(showRefCmd)
}

var showRefCmd = &cobra.Command{
	Use:   "show-ref",
	Short: "a very attempt at listing repository refs",
	Long:  "a very very bad attempt at listing repository refs from scratch",
	Args:  cobra.NoArgs,
	RunE:  runShowRef,
}

func runShowRef(cmd *cobra.Command, args []string) error {
	repository, err := repo.FindRepository(".", true)
	if err != nil {
		return err
	}

	refmap, err := ref.RefList(repository, "")
	if err != nil {
		return err
	}

	err = ref.RefShow(refmap, "refs", true)
	if err != nil {
		return err
	}

	return nil
}
