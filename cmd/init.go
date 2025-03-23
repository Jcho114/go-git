package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Jcho114/go-git/repo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "a very bad attempt at creating a new git repository",
	Long:  "a very very bad attempt at creating a new git repository from scratch",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	path := args[0]
	repository, err := repo.NewRepository(path, true)
	if err != nil {
		return err
	}

	info, err := os.Stat(repository.Worktree)
	pathexists := !errors.Is(err, os.ErrNotExist)
	_, err = os.Stat(repository.Gitdir)
	gitexists := !errors.Is(err, os.ErrNotExist)
	if !pathexists {
		err := os.Mkdir(repository.Worktree, 0755)
		if err != nil {
			return err
		}
	} else if err != nil && !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	} else if gitexists {
		return fmt.Errorf("%s is not empty", path)
	}

	objectsdir := filepath.Join(repository.Gitdir, "objects")
	err = os.MkdirAll(objectsdir, 0755)
	if err != nil {
		return err
	}

	branchesdir := filepath.Join(repository.Gitdir, "branches")
	err = os.MkdirAll(branchesdir, 0755)
	if err != nil {
		return err
	}

	tagrefsdir := filepath.Join(repository.Gitdir, "refs", "tags")
	err = os.MkdirAll(tagrefsdir, 0755)
	if err != nil {
		return err
	}

	headrefsdir := filepath.Join(repository.Gitdir, "refs", "heads")
	err = os.MkdirAll(headrefsdir, 0755)
	if err != nil {
		return err
	}

	descfilepath := filepath.Join(repository.Gitdir, "description")
	err = os.WriteFile(descfilepath, []byte("Unnamed repository; edit this file 'description' to name the repository.\n"), 0644)
	if err != nil {
		return err
	}

	headfilepath := filepath.Join(repository.Gitdir, "HEAD")
	err = os.WriteFile(headfilepath, []byte("ref: refs/heads/master\n"), 0644)
	if err != nil {
		return err
	}

	configfilepath := filepath.Join(repository.Gitdir, "config")
	err = repository.Config.Write(configfilepath)
	if err != nil {
		return err
	}

	return nil
}
