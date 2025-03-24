package obj

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

type kvlmap = map[string][]string

type Commit struct {
	Kvlm kvlmap
}

func NewCommit(buffer []byte) *Commit {
	kvlm := kvlmap{}
	return &Commit{Kvlm: kvlm}
}

func (c *Commit) Serialize(repository *repo.Repository) string {
	return serializeKVLM(c.Kvlm)
}

func (c *Commit) Deserialize(content string) {
	c.Kvlm = parseKVLM([]byte(content), kvlmap{})
}

func (c *Commit) Type() string {
	return "commit"
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

func parseKVLM(content []byte, dct kvlmap) kvlmap {
	start := 0
	for {
		spaceindex := bytes.Index(content[start:], []byte(" "))
		newlineindex := bytes.Index(content[start:], []byte("\n"))

		if spaceindex == -1 || newlineindex < spaceindex {
			message := string(content[start+1:])
			dct[""] = []string{message}
			break
		}

		key := string(content[start:spaceindex])
		end := start
		for {
			end = bytes.Index([]byte(content[end+1:]), []byte("\n"))
			if content[end+1] != ' ' {
				break
			}
		}
		value := string(content[spaceindex+1 : end])
		value = strings.ReplaceAll(value, "\n ", "\n")

		if _, ok := dct[key]; !ok {
			dct[key] = []string{}
		}
		dct[key] = append(dct[key], value)
	}

	return dct
}

func serializeKVLM(kvlm kvlmap) string {
	res := ""

	for key := range kvlm {
		if key == "" {
			continue
		}

		values := kvlm[key]
		for _, value := range values {
			res += key + " " + strings.ReplaceAll(value, "\n", "\n ") + "\n"
		}
	}

	return res
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
