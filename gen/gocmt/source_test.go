package gocmt

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	. "github.com/ng-vu/sqlgen/mock"
)

var pkg string

func init() {
	root, err := filepath.Abs("../..")
	if err != nil {
		panic(err)
	}
	if filepath.Base(root) != "sqlgen" {
		panic(fmt.Sprintf("expect package sqlgen, got %v", filepath.Base(root)))
	}
	pkg = filepath.Join(root, "gen/gocmt")
}

func TestParsePkg1(t *testing.T) {
	res, err := ParseDir(filepath.Join(pkg, "testpkg1"))
	AssertNoError(t, err)

	{
		expected := `[[package][Span・comment][Standalone 1][Multi-line 1][Multi-line 2.1・Multi-line 2.2][Standalone 2]]`
		AssertEqual(t, join(res.Block), expected)
	}
	{
		var ss []string
		for _, typ := range res.Types {
			s := fmt.Sprintf("%v-%v\n", typ.Type.Name.Name, typ.Comment)
			ss = append(ss, s)
		}
		sort.Strings(ss)
		expected := `
A-[A]
A2-[A2]
B-[B]
C-[C]
D-[D]
E-[E]
E2-[E2]
H-[H block]
`[1:]
		AssertEqual(t, strings.Join(ss, ""), expected)
	}
}

func TestParsePkgError(t *testing.T) {
	tests := []struct {
		pkg string
		err string
	}{
		{
			"e_testpkg_mix1",
			"Must not mix declaration on type A and B",
		},
		{
			"e_testpkg_mix2",
			"Must not mix declaration on type A and B",
		},
		{
			"e_testpkg_mix3",
			"Must not mix declaration on type A and B",
		},
		{
			"e_testpkg_mix4",
			"Must not mix declaration on type A and B",
		},
		{
			"e_testpkg_multi1",
			"Multiple declarations on type A",
		},
		{
			"e_testpkg_multi2",
			"Multiple declarations on type A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			_, err := ParseDir(filepath.Join(pkg, tt.pkg))
			AssertErrorEqual(t, err, tt.err)
		})
	}
}
