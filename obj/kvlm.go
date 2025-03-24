package obj

import (
	"bytes"
	"strings"
)

type kvlmap = map[string][]string

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
