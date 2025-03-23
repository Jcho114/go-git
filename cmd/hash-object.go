package cmd

import (
	"fmt"
	"os"

	"github.com/Jcho114/go-git/obj"
	"github.com/Jcho114/go-git/repo"
	"github.com/spf13/cobra"
)

var (
	format string
	write  bool
)

func init() {
	hashObjectCmd.Flags().StringVar(&format, "t", "blob", "git object type")
	hashObjectCmd.Flags().BoolVar(&write, "w", false, "actually write object to database")
	rootCmd.AddCommand(hashObjectCmd)
}

var hashObjectCmd = &cobra.Command{
	Use:   "hash-object",
	Short: "a very attempt at converting a file into a git object",
	Long:  "a very very bad attempt at converting a file into a git object from scratch",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:  runHashObject,
}

func runHashObject(cmd *cobra.Command, args []string) error {
	if format != "blob" && format != "commit" && format != "tag" && format != "tree" {
		return fmt.Errorf("invalid object type")
	}

	var repository *repo.Repository
	var err error
	if write {
		repository, err = repo.FindRepository(".", true)
		if err != nil {
			return err
		}
	} else {
		repository = nil
	}

	path := args[0]
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var object obj.Object
	switch format {
	case "commit":
		object = obj.NewCommit(data)
	case "tree":
		object = obj.NewTree(data)
	case "tag":
		object = obj.NewTag(data)
	case "blob":
		object = obj.NewBlob(data)
	}

	sha, err := obj.ObjectWrite(repository, object)
	if err != nil {
		return err
	}

	fmt.Println(sha)

	return nil
}
