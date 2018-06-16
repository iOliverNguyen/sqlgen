package gosrc

import (
	"errors"
	"strings"
	"unicode"
)

const (
	CommentPrefix = "sqlgen:"

	modeBlock = 1
	modeLine  = 2
)

var (
	ErrComment1 = errors.New("sqlgen: Must not mix block declaration")
	ErrComment2 = errors.New("sqlgen: Must not mix block declaration and line declaration")
)

func ParseComment(lines []string) (res []string, _ error) {
	mode := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
			// skip

		case line == CommentPrefix:
			switch mode {
			case modeBlock:
				return nil, ErrComment1
			case modeLine:
				return nil, ErrComment2
			}
			mode = modeBlock

		case strings.HasPrefix(line, CommentPrefix):
			if mode == modeBlock {
				return nil, ErrComment2
			}
			mode = modeLine
			line = line[len(CommentPrefix):]
			line = strings.TrimLeftFunc(line, unicode.IsSpace)
			res = append(res, line)

		default:
			if mode == modeBlock {
				res = append(res, line)
			}
		}
	}
	return res, nil
}
