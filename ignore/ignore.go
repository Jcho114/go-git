package ignore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jcho114/go-git/index"
	"github.com/Jcho114/go-git/obj"
	"github.com/Jcho114/go-git/repo"
)

type IgnoreRule struct {
	Pattern string
	Ignore  bool
}

type Ignore struct {
	Absolute [][]IgnoreRule
	Scoped   map[string][]IgnoreRule
}

func IgnoreRead(repository *repo.Repository) (*Ignore, error) {
	absolute := [][]IgnoreRule{}
	scoped := make(map[string][]IgnoreRule)

	repofile := filepath.Join(repository.Gitdir, "info/exclude")
	_, err := os.Stat(repofile)
	pathexists := !errors.Is(err, os.ErrNotExist)
	if pathexists {
		content, err := os.ReadFile(repofile)
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(content), "\n")
		parsed := ignoreParse(lines)
		absolute = append(absolute, parsed)
	}

	var confighome string
	if val := os.Getenv("XDG_CONFIG_HOME"); val != "" {
		confighome = val
	} else {
		confighome = "~/.config"
	}

	globalfile := filepath.Join(confighome, "git/ignore")
	_, err = os.Stat(globalfile)
	pathexists = !errors.Is(err, os.ErrNotExist)
	if pathexists {
		content, err := os.ReadFile(globalfile)
		if err != nil {
			return nil, err
		}
		lines := strings.Split(string(content), "\n")
		parsed := ignoreParse(lines)
		absolute = append(absolute, parsed)
	}

	index, err := index.IndexRead(repository)
	if err != nil {
		return nil, err
	}

	for _, entry := range index.Entries {
		if entry.Name == ".gitignore" || strings.HasSuffix(entry.Name, "/.gitignore") {
			dirname := filepath.Dir(entry.Name)
			content, err := obj.ObjectRead(repository, entry.Sha)
			if err != nil {
				return nil, err
			}
			blob, ok := content.(*obj.Blob)
			if !ok {
				return nil, fmt.Errorf("object %s is not a blob type", entry.Name)
			}
			lines := strings.Split(string(blob.Data), "\n")
			parsed := ignoreParse(lines)
			scoped[dirname] = parsed
		}
	}

	ignore := &Ignore{
		Absolute: absolute,
		Scoped:   scoped,
	}
	return ignore, nil
}

func ignoreParseLine(content string) *IgnoreRule {
	content = strings.TrimSpace(content)

	if len(content) == 0 || content == "#" {
		return nil
	}

	if content[0] == '!' {
		return &IgnoreRule{
			Pattern: content[1:],
			Ignore:  false,
		}
	}

	if content[0] == '\\' {
		return &IgnoreRule{
			Pattern: content[1:],
			Ignore:  true,
		}
	}

	return &IgnoreRule{
		Pattern: content,
		Ignore:  true,
	}
}

func ignoreParse(lines []string) []IgnoreRule {
	res := []IgnoreRule{}

	for _, line := range lines {
		parsed := ignoreParseLine(line)
		if parsed != nil {
			res = append(res, *parsed)
		}
	}

	return res
}
