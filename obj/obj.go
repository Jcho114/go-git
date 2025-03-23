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
	defer reader.Close()

	_, err = io.Copy(buffer, reader)
	if err != nil {
		return nil, err
	}

	wsindex := bytes.Index(buffer.Bytes(), []byte(" "))
	format := buffer.Bytes()[:wsindex]

	nulindex := bytes.Index(buffer.Bytes(), []byte("\x00"))
	size, err := strconv.Atoi(string(buffer.Bytes()[:nulindex]))
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
		err := os.WriteFile(objectfilepath, sha, 0644)
		if err != nil {
			return "", err
		}
	}

	return string(sha), nil
}

type Commit struct {
}

func NewCommit(buffer []byte) *Commit {
	return &Commit{}
}

func (c *Commit) Serialize(repository *repo.Repository) string {
	return ""
}

func (c *Commit) Deserialize(content string) {

}

func (c *Commit) Type() string {
	return "commit"
}

type Tree struct {
}

func NewTree(buffer []byte) *Tree {
	return &Tree{}
}

func (t *Tree) Serialize(repository *repo.Repository) string {
	return ""
}

func (t *Tree) Deserialize(content string) {

}

func (t *Tree) Type() string {
	return "tree"
}

type Tag struct {
}

func NewTag(buffer []byte) *Tag {
	return &Tag{}
}

func (t *Tag) Serialize(repository *repo.Repository) string {
	return ""
}

func (t *Tag) Deserialize(content string) {

}

func (t *Tag) Type() string {
	return "tag"
}

type Blob struct {
	data []byte
}

func NewBlob(buffer []byte) *Blob {
	return &Blob{}
}

func (b *Blob) Serialize(repository *repo.Repository) string {
	return string(b.data)
}

func (b *Blob) Deserialize(content string) {
	b.data = []byte(content)
}

func (b *Blob) Type() string {
	return "blob"
}
