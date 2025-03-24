package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Jcho114/go-git/obj"
	"github.com/Jcho114/go-git/repo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkoutCmd)
}

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "a very attempt at instantiating a commit",
	Long:  "a very very bad attempt at instantiating a commit from scratch",
	Args:  cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
	RunE:  runCheckout,
}

func runCheckout(cmd *cobra.Command, args []string) error {
	commit, path := args[0], args[1]
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	repository, err := repo.FindRepository(".", true)
	if err != nil {
		return err
	}

	commitname := obj.ObjectFind(repository, commit, "commit", true)
	object, err := obj.ObjectRead(repository, commitname)
	if err != nil {
		return err
	}
	commitobject, ok := object.(*obj.Commit)
	if !ok {
		return fmt.Errorf("object is not a commit: %s", commitname)
	}

	trees, ok := commitobject.Kvlm["tree"]
	if !ok {
		return fmt.Errorf("commit object does not have a tree value")
	}
	if len(trees) != 1 {
		return fmt.Errorf("commit object does not only have 1 tree value")
	}
	treename := commitobject.Kvlm["tree"][0]
	object, err = obj.ObjectRead(repository, treename)
	if err != nil {
		return err
	}

	treeobject, ok := object.(*obj.Tree)
	if !ok {
		return fmt.Errorf("object is not a tree: %s", treename)
	}

	info, err := os.Stat(path)
	pathexists := !errors.Is(err, os.ErrNotExist)
	if !pathexists {
		err := os.MkdirAll(path, 0750)
		if err != nil {
			return err
		}
	} else {
		if !info.Mode().IsDir() {
			return fmt.Errorf("path %s is not a directory", path)
		}
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		_, err = file.Readdirnames(1)
		if err != io.EOF {
			return fmt.Errorf("path %s is not empty", path)
		}
	}

	err = checkoutTree(repository, treeobject, path)
	if err != nil {
		return err
	}

	return nil
}

func checkoutTree(repository *repo.Repository, tree *obj.Tree, path string) error {
	for _, item := range tree.Items {
		object, err := obj.ObjectRead(repository, item.Sha)
		if err != nil {
			return err
		}

		destpath := filepath.Join(path, item.Path)

		if tree, ok := object.(*obj.Tree); ok {
			err := os.MkdirAll(destpath, 0750)
			if err != nil {
				return err
			}

			err = checkoutTree(repository, tree, destpath)
			if err != nil {
				return err
			}
		} else if blob, ok := object.(*obj.Blob); ok {
			err := os.WriteFile(destpath, blob.Data, 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
