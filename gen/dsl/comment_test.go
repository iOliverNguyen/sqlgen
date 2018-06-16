package dsl_test

import (
	"io"
	"strings"
	"testing"

	"fmt"

	. "github.com/ng-vu/sqlgen/gen/dsl"
)

type CommentReaderMock struct {
	lines []string
	l, i  int
}

func NewMock(text string) *CommentReaderMock {
	lines := strings.Split(text, "\n")
	return &CommentReaderMock{lines: lines, l: len(lines) - 1, i: -1}
}

func (c *CommentReaderMock) ReadLine() (string, error) {
	fmt.Println("line", c.i, c.l)
	if c.i >= c.l {
		return "", io.EOF
	}
	c.i = c.i + 1
	return c.lines[c.i], nil
}

func TestParseComment(t *testing.T) {
	t.Run("Single line comment", func(t *testing.T) {
		m := NewMock("sqlgen: Hello")
		output, err := ParseComment(m)
		assertNoError(t, err)
		assertEqual(t, output, "Hello")
	})

	t.Run("Multiple line comment", func(t *testing.T) {
		m := NewMock(`
This is a comment about sqlgen:
  sqlgen:Hello
sqlgen: World
Another comment about sqlgen: This line should not be included
sqlgen:!`)
		output, err := ParseComment(m)
		assertNoError(t, err)
		assertEqual(t, output, "Hello\nWorld\n!")
	})

	t.Run("Block comment", func(t *testing.T) {
		m := NewMock(`
sqlgen:
  Hello
    World
!`)
		output, err := ParseComment(m)
		assertNoError(t, err)
		assertEqual(t, output, "Hello\nWorld\n!")
	})

	t.Run("No comment", func(t *testing.T) {
		m := NewMock(`
No comment about sqlgen:
`)
		output, err := ParseComment(m)
		assertNoError(t, err)
		assertEqual(t, output, "")
	})

	t.Run("Error: Mix block", func(t *testing.T) {
		m := NewMock(`sqlgen:
Hello
sqlgen:
World`)
		_, err := ParseComment(m)
		assertErrorEqual(t, err, ErrComment1.Error())
	})

	t.Run("Error: Mix block and line", func(t *testing.T) {
		m := NewMock(`sqlgen:Hello
sqlgen:
World`)
		_, err := ParseComment(m)
		assertErrorEqual(t, err, ErrComment2.Error())
	})

	t.Run("Error: Mix block and line 2", func(t *testing.T) {
		m := NewMock(`sqlgen:
Hello
sqlgen:World`)
		_, err := ParseComment(m)
		assertErrorEqual(t, err, ErrComment2.Error())
	})
}
