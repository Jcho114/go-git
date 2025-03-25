package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Jcho114/go-git/ignore"
	"github.com/Jcho114/go-git/repo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkIgnoreCmd)
}

var checkIgnoreCmd = &cobra.Command{
	Use:   "check-ignore",
	Short: "a very attempt at outputing paths that should be ignored",
	Long:  "a very very bad attempt at outputing paths that should be ignored from scratch",
	Args:  cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
	RunE:  runCheckIgnore,
}

func runCheckIgnore(cmd *cobra.Command, args []string) error {
	repository, err := repo.FindRepository(".", true)
	if err != nil {
		return err
	}

	rules, err := ignore.IgnoreRead(repository)
	if err != nil {
		return err
	}
	for _, path := range args {
		res, err := checkIgnore(rules, path)
		if err != nil {
			return err
		}
		if *res {
			fmt.Println(path)
		}
	}
	return nil
}

func checkIgnoreOne(rules []ignore.IgnoreRule, path string) (*bool, error) {
	var res *bool
	for _, entry := range rules {
		pattern, value := entry.Pattern, entry.Ignore
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			return nil, err
		}
		if matched {
			res = &value
		}
	}
	return res, nil
}

func checkIgnoreScoped(rules map[string][]ignore.IgnoreRule, path string) (*bool, error) {
	dirname := filepath.Dir(path)
	for {
		if nestedrules, ok := rules[dirname]; ok {
			res, err := checkIgnoreOne(nestedrules, path)
			if err != nil {
				return nil, err
			}
			if res != nil && *res {
				return res, nil
			}
		}

		if dirname == "." {
			break
		}

		dirname = filepath.Dir(path)
	}

	return nil, nil
}

func checkIgnoreAbsolute(rules [][]ignore.IgnoreRule, path string) (*bool, error) {
	for _, rule := range rules {
		res, err := checkIgnoreOne(rule, path)
		if err != nil {
			return nil, err
		}
		if res != nil && *res {
			return res, nil
		}
	}
	var falseVal = false
	return &falseVal, nil
}

func checkIgnore(rules *ignore.Ignore, path string) (*bool, error) {
	if filepath.IsAbs(path) {
		return nil, fmt.Errorf("path is not absolute")
	}

	res, err := checkIgnoreScoped(rules.Scoped, path)
	if err != nil {
		return nil, err
	}
	if res != nil {
		return res, nil
	}

	res, err = checkIgnoreAbsolute(rules.Absolute, path)
	if err != nil {
		return nil, err
	}
	return res, nil
}
