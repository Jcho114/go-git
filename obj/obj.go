package obj

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"crypto/sha1"

	"github.com/Jcho114/go-git/ref"
	"github.com/Jcho114/go-git/repo"
)

type Object interface {
	Serialize(repository *repo.Repository) string
	Deserialize(content string)
	Type() string
}

var hashRegex = regexp.MustCompile("^[0-9A-Fa-f]{4,40}$")

func ObjectFind(repository *repo.Repository, name string, format string, follow bool) (string, error) {
	objnames, err := objectResolve(repository, name)
	if err != nil {
		return "", err
	}

	if len(objnames) == 0 || len(objnames) > 1 {
		return "", fmt.Errorf("%s is an ambiguous reference", name)
	}

	objname := objnames[0]

	if format == "any" {
		return objname, nil
	}

	for {
		object, err := ObjectRead(repository, objname)
		if err != nil {
			return "", nil
		}

		switch format {
		case "blob":
			if _, ok := object.(*Blob); ok {
				return objname, nil
			}
		case "commit":
			_, ok := object.(*Commit)
			if ok {
				return objname, nil
			}
		case "tag":
			_, ok := object.(*Tag)
			if ok {
				return objname, nil
			}
		case "tree":
			if _, ok := object.(*Tree); ok {
				return objname, nil
			}
		}

		if commit, ok := object.(*Commit); ok {
			value, ok := commit.Kvlm["tree"]
			if !ok || len(value) == 0 {
				return "", fmt.Errorf("commit does not have tree field in it")
			}
			objname = commit.Kvlm["tree"][0]
		} else if tag, ok := object.(*Tag); ok {
			value, ok := tag.Kvlm["object"]
			if !ok || len(value) == 0 {
				return "", fmt.Errorf("tag does not have object field in it")
			}
			objname = tag.Kvlm["object"][0]
		} else {
			return "", fmt.Errorf("unable to find object %s with type %s", name, format)
		}
	}
}

func objectResolve(repository *repo.Repository, name string) ([]string, error) {
	candidates := []string{}

	if strings.TrimSpace(name) == "" {
		err := fmt.Errorf("unable to resolve an empty name")
		return nil, err
	}

	if name == "HEAD" {
		objname, err := ref.RefResolve(repository, "HEAD")
		if err != nil {
			return nil, err
		}
		return []string{objname}, nil
	}

	if hashRegex.MatchString(name) {
		prefix := name[:2]
		path := filepath.Join(repository.Gitdir, "objects", prefix)
		info, err := os.Stat(path)
		pathexists := !errors.Is(err, os.ErrNotExist)
		if pathexists && info.Mode().IsDir() {
			rem := prefix[2:]
			files, err := os.ReadDir(path)
			if err != nil {
				return nil, err
			}
			for _, file := range files {
				if strings.HasPrefix(file.Name(), rem) {
					candidates = append(candidates, file.Name())
				}
			}
		}
	}

	tagname, err := ref.RefResolve(repository, "refs/tags/"+name)
	pathexists := errors.Is(err, os.ErrNotExist)
	if pathexists {
		return nil, err
	}
	if err == nil {
		candidates = append(candidates, tagname)
	}

	branchname, err := ref.RefResolve(repository, "refs/heads/"+name)
	pathexists = errors.Is(err, os.ErrNotExist)
	if pathexists {
		return nil, err
	}
	if err == nil {
		candidates = append(candidates, branchname)
	}

	return candidates, nil
}

func ObjectRead(repository *repo.Repository, sha string) (Object, error) {
	objfilepath := filepath.Join(repository.Gitdir, "objects", sha[0:2], sha[2:])
	info, err := os.Stat(objfilepath)
	if err != nil {
		return nil, err
	}

	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("object does not exist")
	}

	file, err := os.ReadFile(objfilepath)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(file)
	reader, err := zlib.NewReader(buffer)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	_, err = io.Copy(buffer, reader)
	if err != nil {
		return nil, err
	}

	wsindex := bytes.Index(buffer.Bytes(), []byte(" "))
	format := buffer.Bytes()[:wsindex]

	nulindex := wsindex + bytes.Index(buffer.Bytes()[wsindex:], []byte("\x00"))
	size, err := strconv.Atoi(string(buffer.Bytes()[wsindex+1 : nulindex]))
	if err != nil {
		return nil, err
	}

	if size != buffer.Len()-nulindex-1 {
		return nil, fmt.Errorf("malformed object %s: bad length", sha)
	}

	switch string(format) {
	case "commit":
		return NewCommit(buffer.Bytes()[nulindex+1:]), nil
	case "tree":
		return NewTree(buffer.Bytes()[nulindex+1:]), nil
	case "tag":
		return NewTag(buffer.Bytes()[nulindex+1:]), nil
	case "blob":
		return NewBlob(buffer.Bytes()[nulindex+1:]), nil
	default:
		return nil, fmt.Errorf("unknown type %s for object %s", format, sha)
	}
}

func ObjectWrite(repository *repo.Repository, object Object) (string, error) {
	data := object.Serialize(repository)
	result := object.Type() + " " + strconv.Itoa(len(data)) + "\x00" + data
	hash := sha1.New()
	sha := hash.Sum([]byte(result))

	if repository != nil {
		objectfilepath := filepath.Join(repository.Gitdir, "objects", string(sha[:2]), string(sha[2:]))

		var buffer bytes.Buffer
		writer := zlib.NewWriter(&buffer)
		_, err := writer.Write(sha)
		if err != nil {
			return "", err
		}
		writer.Close()

		err = os.WriteFile(objectfilepath, buffer.Bytes(), 0644)
		if err != nil {
			return "", err
		}
	}

	return string(sha), nil
}
