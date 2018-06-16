package gosrc

import (
	"os"
	"path/filepath"
	"testing"

	"fmt"

	"bytes"

	. "github.com/ng-vu/sqlgen/mock"
)

var testpkg string

func init() {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		panic("No GOPATH")
	}
	testpkg = filepath.Join(gopath, "src/github.com/ng-vu/sqlgen/gen/gosrc/testpkg")
}

func TestParsePkg1(t *testing.T) {
	res, err := ParseDir(testpkg + "1")
	AssertNoError(t, err)

	{
		expected := `
package
Block
Floating 1
Floating 2
Multi-line 1
Multi-line 2.1
Multi-line 2.2
Z
`[1:]
		AssertEqual(t, res.Block, expected)
	}
	{
		var b bytes.Buffer
		for _, typ := range res.Types {
			fmt.Fprintf(&b, "%v-%v\n", typ.Type.Name.Name, typ.Comment)
		}
		expected := `
A2-A2
A-A
B-B comment
C-C
D-D
E-E
`[1:]
		AssertEqual(t, b.String(), expected)
	}
}

func TestParsePkg2(t *testing.T) {
	// TODO
	ParseDir(testpkg + "2")
}

func TestParsePkg3(t *testing.T) {
	// TODO
	ParseDir(testpkg + "3")
}

func TestParsePkg4(t *testing.T) {
	_, err := ParseDir(testpkg + "4")
	AssertErrorEqual(t, err, "Multiple declaration on type A")
}
