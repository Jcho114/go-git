package obj

import (
	"github.com/Jcho114/go-git/repo"
)

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
