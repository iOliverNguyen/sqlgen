package gocmt

import (
	"strings"
	"testing"

	. "github.com/ng-vu/sqlgen/mock"
)

func split(src string) []string {
	return strings.Split(src, "\n")
}

func join(groups [][]string) string {
	var b strings.Builder
	b.WriteString("[")
	for _, group := range groups {
		b.WriteString("[")
		for j, line := range group {
			if j > 0 {
				b.WriteString("・")
			}
			b.WriteString(line)
		}
		b.WriteString("]")
	}
	b.WriteString("]")
	return b.String()
}

func TestParseComment(t *testing.T) {
	t.Run("Empty comment", func(t *testing.T) {
		cmts := split("sqlgen:")
		groups, err := ParseComment(cmts)
		AssertNoError(t, err)
		AssertEqual(t, len(groups), 1)
		AssertEqual(t, len(groups[0]), 0)
	})

	t.Run("Single line comment", func(t *testing.T) {
		cmts := split("sqlgen: Hello")
		groups, err := ParseComment(cmts)
		AssertNoError(t, err)
		AssertEqual(t, len(groups), 1)
		AssertEqual(t, join(groups), "[[Hello]]")
	})

	t.Run("Multiple line comment", func(t *testing.T) {
		cmts := split(`
This is a comment about sqlgen:
  sqlgen:Hello
sqlgen: World
  from   sqlgen

Another comment about sqlgen: This line should not be included
sqlgen:!`)
		groups, err := ParseComment(cmts)
		AssertNoError(t, err)
		AssertEqual(t, join(groups), "[[Hello][World・from   sqlgen][!]]")
	})

	t.Run("No comment", func(t *testing.T) {
		cmts := split(`
No comment about sqlgen:
`)
		groups, err := ParseComment(cmts)
		AssertNoError(t, err)
		AssertEqual(t, join(groups), "[]")
	})

	t.Run("First line empty", func(t *testing.T) {
		m := split(`
sqlgen:
  Hello
    World
!`)
		groups, err := ParseComment(m)
		AssertNoError(t, err)
		AssertEqual(t, join(groups), "[[Hello・World・!]]")
	})

	t.Run("Mix first line empty 1", func(t *testing.T) {
		cmts := split(`sqlgen:
Hello
sqlgen:
World`)
		groups, err := ParseComment(cmts)
		AssertNoError(t, err)
		AssertEqual(t, join(groups), "[[Hello][World]]")
	})

	t.Run("Mix first line empty 2", func(t *testing.T) {
		cmts := split(`sqlgen:Hello
sqlgen:
World`)
		groups, err := ParseComment(cmts)
		AssertNoError(t, err)
		AssertEqual(t, join(groups), "[[Hello][World]]")
	})

	t.Run("Mix first line empty 3", func(t *testing.T) {
		cmts := split(`sqlgen:
Hello
sqlgen:World`)
		groups, err := ParseComment(cmts)
		AssertNoError(t, err)
		AssertEqual(t, join(groups), "[[Hello][World]]")
	})
}
