package gosrc

import (
	"strings"
	"testing"

	. "github.com/ng-vu/sqlgen/mock"
)

func split(src string) []string {
	return strings.Split(src, "\n")
}

func join(lines []string) string {
	return strings.Join(lines, "|")
}

func TestParseComment(t *testing.T) {
	t.Run("Single line comment", func(t *testing.T) {
		cmts := split("sqlgen: Hello")
		output, err := ParseComment(cmts)
		AssertNoError(t, err)
		AssertEqual(t, join(output), "Hello")
	})

	t.Run("Multiple line comment", func(t *testing.T) {
		cmts := split(`
This is a comment about sqlgen:
  sqlgen:Hello
sqlgen: World
Another comment about sqlgen: This line should not be included
sqlgen:!`)
		output, err := ParseComment(cmts)
		AssertNoError(t, err)
		AssertEqual(t, join(output), "Hello|World|!")
	})

	t.Run("Block comment", func(t *testing.T) {
		m := split(`
sqlgen:
  Hello
    World
!`)
		output, err := ParseComment(m)
		AssertNoError(t, err)
		AssertEqual(t, join(output), "Hello|World|!")
	})

	t.Run("No comment", func(t *testing.T) {
		cmts := split(`
No comment about sqlgen:
`)
		output, err := ParseComment(cmts)
		AssertNoError(t, err)
		AssertEqual(t, join(output), "")
	})

	t.Run("Error: Mix block", func(t *testing.T) {
		cmts := split(`sqlgen:
Hello
sqlgen:
World`)
		_, err := ParseComment(cmts)
		AssertErrorEqual(t, err, ErrComment1.Error())
	})

	t.Run("Error: Mix block and line", func(t *testing.T) {
		cmts := split(`sqlgen:Hello
sqlgen:
World`)
		_, err := ParseComment(cmts)
		AssertErrorEqual(t, err, ErrComment2.Error())
	})

	t.Run("Error: Mix block and line 2", func(t *testing.T) {
		cmts := split(`sqlgen:
Hello
sqlgen:World`)
		_, err := ParseComment(cmts)
		AssertErrorEqual(t, err, ErrComment2.Error())
	})
}
