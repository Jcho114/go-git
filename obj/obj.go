package obj

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"crypto/sha1"

	"github.com/Jcho114/go-git/repo"
)

type Object interface {
	Serialize(repository *repo.Repository) string
	Deserialize(content string)
	Type() string
}

func ObjectFind(repository *repo.Repository, name string, format string, follow bool) string {
	return name
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
