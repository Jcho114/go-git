package obj

import (
	"github.com/Jcho114/go-git/repo"
)

type Tag struct {
	Kvlm kvlmap
}

func NewTag(buffer []byte) *Tag {
	kvlm := kvlmap{}
	tag := &Tag{Kvlm: kvlm}
	if buffer != nil {
		tag.Deserialize(string(buffer))
	}
	return tag
}

func (t *Tag) Serialize(repository *repo.Repository) string {
	return serializeKVLM(t.Kvlm)
}

func (t *Tag) Deserialize(content string) {
	t.Kvlm = parseKVLM([]byte(content), kvlmap{})
}

func (t *Tag) Type() string {
	return "tag"
}
