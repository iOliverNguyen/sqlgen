package dsl

import (
	"bytes"
	"errors"
	"io"
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

type CommentReader interface {
	ReadLine() (string, error)
}

func ParseComment(r CommentReader) (string, error) {
	mode := 0
	var buf bytes.Buffer
loop:
	for {
		line, err := r.ReadLine()
		switch err {
		case nil:
		case io.EOF:
			break loop
		default:
			return "", err
		}

		line = strings.TrimSpace(line)
		switch {
		case line == "":
			if mode == modeBlock {
				buf.WriteString("\n")
			}

		case line == CommentPrefix:
			switch mode {
			case modeBlock:
				return "", ErrComment1
			case modeLine:
				return "", ErrComment2
			}
			mode = modeBlock

		case strings.HasPrefix(line, CommentPrefix):
			if mode == modeBlock {
				return "", ErrComment2
			}
			mode = modeLine
			line = line[len(CommentPrefix):]
			line = strings.TrimLeftFunc(line, unicode.IsSpace)
			buf.WriteString(line)
			buf.WriteString("\n")

		default:
			if mode == modeBlock {
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}
	}
	return strings.TrimSpace(buf.String()), nil
}
