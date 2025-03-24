package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Jcho114/go-git/obj"
	"github.com/Jcho114/go-git/repo"
	"github.com/spf13/cobra"
)

var recursive bool

func init() {
	lsTreeCmd.Flags().BoolVar(&recursive, "r", false, "print tree contents recursively")
	rootCmd.AddCommand(lsTreeCmd)
}

var lsTreeCmd = &cobra.Command{
	Use:   "ls-tree",
	Short: "a very attempt at printing tree contents",
	Long:  "a very very bad attempt at printing tree contents from scratch",
	Args:  cobra.ExactArgs(1),
	RunE:  runLsTree,
}

func runLsTree(cmd *cobra.Command, args []string) error {
	repository, err := repo.FindRepository(".", true)
	if err != nil {
		return err
	}

	err = lsTree(repository, args[0], recursive, "")
	if err != nil {
		return err
	}

	return nil
}

func lsTree(repository *repo.Repository, ref string, recursive bool, prefix string) error {
	objname := obj.ObjectFind(repository, ref, "tree", true)
	object, err := obj.ObjectRead(repository, objname)
	if err != nil {
		return err
	}

	tree, ok := object.(*obj.Tree)
	if !ok {
		return fmt.Errorf("object is not a tree")
	}

	for _, item := range tree.Items {
		var kind string
		if len(item.Mode) == 5 {
			kind = item.Mode[0:1]
		} else {
			kind = item.Mode[0:2]
		}

		var objtype string
		switch kind {
		case "04":
			objtype = "tree"
		case "10":
			objtype = "blob"
		case "12":
			objtype = "blob"
		case "16":
			objtype = "commit"
		default:
			return fmt.Errorf("invalid tree leaf mode %s", item.Mode)
		}

		if recursive && objtype == "tree" {
			return lsTree(repository, item.Sha, recursive, filepath.Join(prefix, item.Path))
		}
		fmtmode := fmt.Sprintf("%06s", item.Mode)
		fmtpath := filepath.Join(prefix, item.Path)
		fmt.Printf("%s %s %s\t%s\n", fmtmode, objtype, item.Sha, fmtpath)
	}

	return nil
}
