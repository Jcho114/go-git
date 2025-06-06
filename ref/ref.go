package ref

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Jcho114/go-git/repo"
)

func RefResolve(repository *repo.Repository, ref string) (string, error) {
	path := ref
	if !strings.HasPrefix(ref, ".git") {
		path = filepath.Join(repository.Gitdir, ref)
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if !info.Mode().IsRegular() {
		return "", err
	}

	bytescontent, err := os.ReadFile(path)
	if err != nil {
		return "", nil
	}
	content := string(bytescontent)
	content = content[:len(content)-1]

	if strings.HasPrefix(content, "ref: ") {
		content, err := RefResolve(repository, content[5:])
		if err != nil {
			return "", err
		}
		return content, nil
	}

	return content, nil
}

type RefMap = map[string]interface{}

func RefList(repository *repo.Repository, path string) (RefMap, error) {
	if path == "" {
		path = filepath.Join(repository.Gitdir, "refs")
	}

	res := make(RefMap)

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	entrynames := []string{}
	for _, entry := range entries {
		entrynames = append(entrynames, entry.Name())
	}

	slices.Sort(entrynames)
	for _, entryname := range entrynames {
		nextpath := filepath.Join(path, entryname)
		info, err := os.Stat(nextpath)
		if err != nil {
			return nil, err
		}

		if info.Mode().IsDir() {
			refmap, err := RefList(repository, nextpath)
			if err != nil {
				return nil, err
			}
			res[entryname] = refmap
		} else {
			id, err := RefResolve(repository, nextpath)
			if err != nil {
				return nil, err
			}
			res[entryname] = id
		}
	}

	return res, nil
}

func RefShow(refmap RefMap, prefix string, showhash bool) error {
	if prefix != "" {
		prefix += "/"
	}
	for key, value := range refmap {
		switch value := value.(type) {
		case RefMap:
			err := RefShow(value, prefix+key, showhash)
			if err != nil {
				return err
			}
		case string:
			if showhash {
				fmt.Printf("%s %s%s\n", value, prefix, key)
			} else {
				fmt.Printf("%s%s\n", prefix, key)
			}
		default:
			return fmt.Errorf("refmap value is neither a refmap or a string")
		}
	}
	return nil
}
