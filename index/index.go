package index

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/Jcho114/go-git/repo"
)

type IndexTimestamp struct {
	Seconds     int64
	Nanoseconds int64
}

type IndexEntry struct {
	Ctime     IndexTimestamp
	Mtime     IndexTimestamp
	Dev       int
	Ino       int
	Modetype  int
	Modeperms int
	Uid       int
	Gid       int
	Fsize     int
	Sha       string
	Flagvalid bool
	Flagstage int
	Name      string
}

type Index struct {
	Version int
	Entries []IndexEntry
}

const DEFAULT_VERSION = 2

func NewIndex(version int) *Index {
	return &Index{
		Version: version,
		Entries: []IndexEntry{},
	}
}

func IndexRead(repository *repo.Repository) (*Index, error) {
	indexpath := filepath.Join(repository.Gitdir, "index")
	_, err := os.Stat(indexpath)
	pathexists := !errors.Is(err, os.ErrNotExist)
	if !pathexists {
		return NewIndex(DEFAULT_VERSION), nil
	}

	content, err := os.ReadFile(indexpath)
	if err != nil {
		return nil, err
	}

	header := content[:12]
	signature := string(header[:4])
	if err != nil {
		return nil, err
	}
	if signature != "DIRC" {
		return nil, fmt.Errorf("provided index has an invalid signature")
	}
	version := int(binary.BigEndian.Uint32(header[4:8]))
	count := int(binary.BigEndian.Uint32(header[8:12]))

	entries := []IndexEntry{}
	content = content[12:]
	curr := 0

	for range count {
		ctimeseconds := int(binary.BigEndian.Uint32(content[curr : curr+4]))
		ctimenanoseconds := int(binary.BigEndian.Uint32(content[curr+4 : curr+8]))

		mtimeseconds := int(binary.BigEndian.Uint32(content[curr+8 : curr+12]))
		mtimenanoseconds := int(binary.BigEndian.Uint32(content[curr+12 : curr+16]))

		dev := int(binary.BigEndian.Uint32(content[curr+16 : curr+20]))

		ino := int(binary.BigEndian.Uint32(content[curr+20 : curr+24]))

		unused := int(binary.BigEndian.Uint16(content[curr+24 : curr+26]))
		if unused != 0 {
			return nil, fmt.Errorf("unused section is used for an entry the provided index")
		}

		mode := int(binary.BigEndian.Uint16(content[curr+26 : curr+28]))
		modetype := mode >> 12
		if modetype != 0b1000 && modetype != 0b1010 && modetype != 0b1110 {
			return nil, fmt.Errorf("invalid modetype found in an entry in the provided index")
		}
		modeperms := mode & 0b0000000111111111

		uid := int(binary.BigEndian.Uint32(content[curr+28 : curr+32]))

		gid := int(binary.BigEndian.Uint32(content[curr+32 : curr+36]))

		fsize := int(binary.BigEndian.Uint32(content[curr+36 : curr+40]))

		sharaw := int(binary.BigEndian.Uint32(content[curr+40 : curr+44]))
		sha := fmt.Sprintf("%04x", sharaw)

		flags := int(binary.BigEndian.Uint16(content[curr+60 : curr+62]))
		flagvalid := (flags & 0b1000000000000000) != 0
		flagextended := (flags & 0b0100000000000000) != 0
		if flagextended {
			return nil, fmt.Errorf("flag is extended in an entry of the provided index file")
		}
		flagstage := flags & 0b0011000000000000

		namelength := flags & 0b0000111111111111
		curr += 62

		var nameraw []byte
		if namelength < 0xFFF {
			if content[curr+namelength] != 0x00 {
				return nil, fmt.Errorf("invalid name in an entry of the provided index file")
			}
			nameraw = content[curr : curr+namelength]
			curr += namelength + 1
		} else {
			fmt.Printf("notice: name is 0x%X bytes long\n", namelength)
			nullindex := bytes.Index(content[curr+namelength:], []byte{'\x00'}) + curr + namelength
			nameraw = content[curr : curr+nullindex]
		}
		name := string(nameraw)

		curr = 8 * int(math.Ceil(float64(curr)/8))

		entry := IndexEntry{}
		entry.Ctime = IndexTimestamp{Seconds: int64(ctimeseconds), Nanoseconds: int64(ctimenanoseconds)}
		entry.Mtime = IndexTimestamp{Seconds: int64(mtimeseconds), Nanoseconds: int64(mtimenanoseconds)}
		entry.Dev = dev
		entry.Ino = ino
		entry.Modetype = modetype
		entry.Modeperms = modeperms
		entry.Uid = uid
		entry.Gid = gid
		entry.Fsize = fsize
		entry.Sha = sha
		entry.Flagvalid = flagvalid
		entry.Flagstage = flagstage
		entry.Name = name
		entries = append(entries, entry)
	}

	index := &Index{
		Version: version,
		Entries: entries,
	}
	return index, nil
}
