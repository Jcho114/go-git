package cmd

import (
	"fmt"

	"github.com/Jcho114/go-git/obj"
	"github.com/Jcho114/go-git/repo"
	"github.com/spf13/cobra"
)

var objtype string

func init() {
	revParseCmd.Flags().StringVar(&objtype, "type", "", "specify expected type")
	rootCmd.AddCommand(revParseCmd)
}

var revParseCmd = &cobra.Command{
	Use:   "rev-parse",
	Short: "a very attempt at solving references",
	Long:  "a very very bad attempt at solving references from scratch",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:  runRevParse,
}

func runRevParse(cmd *cobra.Command, args []string) error {
	if objtype != "blob" && objtype != "commit" && objtype != "tag" && objtype != "tree" {
		return fmt.Errorf("invalid object type specified")
	}

	name := args[0]

	repository, err := repo.FindRepository(".", true)
	if err != nil {
		return err
	}

	objname, err := obj.ObjectFind(repository, name, objtype, true)
	if err != nil {
		return err
	}

	fmt.Println(objname)

	return nil
}
