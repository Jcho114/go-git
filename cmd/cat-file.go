package cmd

import (
	"fmt"

	"github.com/Jcho114/go-git/obj"
	"github.com/Jcho114/go-git/repo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(catFileCmd)
}

var catFileCmd = &cobra.Command{
	Use:   "cat-file",
	Short: "a very attempt at printing a git object to stdout",
	Long:  "a very very bad attempt at printing a git object to stdout from scratch",
	Args:  validateCatFileCmdArgs,
	RunE:  runCatFile,
}

func validateCatFileCmdArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(2)(cmd, args); err != nil {
		return err
	}

	if err := cobra.OnlyValidArgs(cmd, args); err != nil {
		return err
	}

	format := args[0]
	if format != "blob" && format != "commit" && format != "tag" && format != "tree" {
		return fmt.Errorf("invalid object type")
	}

	return nil
}

func runCatFile(cmd *cobra.Command, args []string) error {
	objtype, objname := args[0], args[1]

	repository, err := repo.FindRepository(".", true)
	if err != nil {
		return err
	}

	objname, err = obj.ObjectFind(repository, objname, objtype, true)
	if err != nil {
		return err
	}
	object, err := obj.ObjectRead(repository, objname)
	if err != nil {
		return err
	}

	fmt.Println(object.Serialize(repository))
	return nil
}
