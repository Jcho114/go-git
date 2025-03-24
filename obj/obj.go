package obj

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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

type kvlmap = map[string][]string

type Commit struct {
	Kvlm kvlmap
}

func NewCommit(buffer []byte) *Commit {
	kvlm := kvlmap{}
	commit := &Commit{Kvlm: kvlm}
	if buffer != nil {
		commit.Deserialize(string(buffer))
	}
	return commit
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
		spaceindex := bytes.IndexByte(content[start:], ' ') + start
		newlineindex := bytes.IndexByte(content[start:], '\n') + start

		if spaceindex == -1 || newlineindex < spaceindex {
			message := string(content[start+1:])
			dct[""] = []string{message}
			break
		}

		key := string(content[start:spaceindex])
		end := start
		for {
			end = bytes.IndexByte(content[end+1:], '\n') + end + 1
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

		start = end + 1
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

type treeLeaf struct {
	Mode string
	Path string
	Sha  string
}

func (l *treeLeaf) Key() string {
	if strings.HasPrefix(l.Mode, "10") {
		return l.Path
	}
	return l.Path + "/"
}

func parseTreeOne(content string, start int) (int, *treeLeaf) {
	spaceindex := strings.Index(content[start:], " ") + start
	mode := content[:spaceindex]
	if len(mode) == 5 {
		mode = "0" + mode
	}

	nullindex := strings.Index(content[spaceindex:], "\x00") + spaceindex
	path := content[spaceindex+1 : nullindex]

	sha := content[nullindex+1 : nullindex+21]

	leaf := &treeLeaf{
		Mode: mode,
		Path: path,
		Sha:  sha,
	}
	return nullindex + 21, leaf
}

func parseTree(content string) []*treeLeaf {
	curr := 0
	res := []*treeLeaf{}

	for curr < len(content) {
		var leaf *treeLeaf
		curr, leaf = parseTreeOne(content, curr)
		res = append(res, leaf)
	}

	return res
}

type Tree struct {
	Items []*treeLeaf
}

func NewTree(buffer []byte) *Tree {
	tree := &Tree{}
	if buffer != nil {
		tree.Deserialize(string(buffer))
	}
	return tree
}

func (t *Tree) Serialize(repository *repo.Repository) string {
	sort.SliceStable(t.Items, func(i, j int) bool {
		return t.Items[i].Key() < t.Items[j].Key()
	})

	res := ""
	for _, item := range t.Items {
		res += item.Mode + " " + item.Path + "\x00" + item.Sha
	}
	return res
}

func (t *Tree) Deserialize(content string) {
	t.Items = parseTree(content)
}

func (t *Tree) Type() string {
	return "tree"
}

type Blob struct {
	data []byte
}

func NewBlob(buffer []byte) *Blob {
	blob := &Blob{}
	if buffer != nil {
		blob.Deserialize(string(buffer))
	}
	return blob
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
