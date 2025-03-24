package obj

import (
	"github.com/Jcho114/go-git/repo"
)

type Blob struct {
	Data []byte
}

func NewBlob(buffer []byte) *Blob {
	blob := &Blob{}
	if buffer != nil {
		blob.Deserialize(string(buffer))
	}
	return blob
}

func (b *Blob) Serialize(repository *repo.Repository) string {
	return string(b.Data)
}

func (b *Blob) Deserialize(content string) {
	b.Data = []byte(content)
}

func (b *Blob) Type() string {
	return "blob"
}
