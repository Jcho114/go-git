package obj

import (
	"encoding/hex"
	"sort"
	"strings"

	"github.com/Jcho114/go-git/repo"
)

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
	mode := content[start:spaceindex]
	if len(mode) == 5 {
		mode = "0" + mode
	}

	nullindex := strings.Index(content[spaceindex:], "\x00") + spaceindex
	path := content[spaceindex+1 : nullindex]

	binarysha := content[nullindex+1 : nullindex+21]
	sha := hex.EncodeToString([]byte(binarysha))

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
