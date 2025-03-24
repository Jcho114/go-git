package cmd

import (
	"fmt"
	"strings"

	"github.com/Jcho114/go-git/obj"
	"github.com/Jcho114/go-git/repo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logCmd)
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "a very attempt at displaying commit history",
	Long:  "a very very bad attempt at displaying commit history from scratch",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runLog,
}

func runLog(cmd *cobra.Command, args []string) error {
	var commit string
	if len(args) == 0 {
		commit = "HEAD"
	} else {
		commit = args[0]
	}

	repository, err := repo.FindRepository(".", true)
	if err != nil {
		return err
	}

	objname := obj.ObjectFind(repository, commit, "commit", true)
	seen := make(map[string]bool)
	fmt.Println("digraph log{")
	fmt.Println("  node[shape=rect]")
	err = outputGraphViz(repository, objname, seen)
	if err != nil {
		return err
	}
	fmt.Println("}")
	return nil
}

func outputGraphViz(repository *repo.Repository, objname string, seen map[string]bool) error {
	if _, ok := seen[objname]; ok {
		return nil
	}
	seen[objname] = true

	object, err := obj.ObjectRead(repository, objname)
	if err != nil {
		return err
	}
	commit, ok := object.(*obj.Commit)
	if !ok {
		return fmt.Errorf("object %s is not a commit object", objname)
	}
	if len(commit.Kvlm[""]) == 0 {
		return fmt.Errorf("commit does not have a message")
	}

	message := strings.TrimSpace(commit.Kvlm[""][0])
	message = strings.ReplaceAll(message, "\\", "\\\\")
	message = strings.ReplaceAll(message, "\"", "\\\"")

	newlineindex := strings.Index(message, "\n")
	if newlineindex != -1 {
		message = message[:newlineindex]
	}

	fmt.Printf("  c_%s [label=\"%s: %s\"]\n", objname, objname[:7], message)

	parents, ok := commit.Kvlm["parent"]
	if !ok {
		return nil
	}

	for _, parent := range parents {
		fmt.Printf("  c_%s -> c_%s\n", objname, parent)
		err := outputGraphViz(repository, parent, seen)
		if err != nil {
			return err
		}
	}

	return nil
}
