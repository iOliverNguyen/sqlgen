package gocmt

import (
	"strings"
	"unicode"
)

const CommentPrefix = "sqlgen:"

func ParseComment(lines []string) (groups [][]string, _ error) {
	var flag bool
	var group []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
			if flag {
				groups = append(groups, group)
				group = nil
			}
			flag = false

		case line == CommentPrefix:
			if flag {
				groups = append(groups, group)
				group = nil
			}
			flag = true

		case strings.HasPrefix(line, CommentPrefix):
			if flag {
				groups = append(groups, group)
				group = nil
			}
			flag = true
			line = line[len(CommentPrefix):]
			line = strings.TrimLeftFunc(line, unicode.IsSpace)
			group = append(group, line)

		default:
			if flag {
				group = append(group, line)
			}
		}
	}
	if flag {
		groups = append(groups, group)
	}
	return groups, nil
}
