// This is a package
//
// sqlgen: package
//
// this line is ignored
package testpkg1

// sqlgen: Span
//   comment
//

// sqlgen: A
type A int

// sqlgen: B
type (
	B int
)

// This comment is ignored
type (
	// sqlgen: C
	C int
)

// sqlgen: Standalone 1

type (
	// sqlgen: D
	D int

	// E is a type
	//
	// sqlgen: E
	E int
)

// This comment is ignored
type (
	D2 int

	// sqlgen: E2
	E2 int
)

// F and G are ignored in result list
type F int

type G int

// H is block comment
// sqlgen:
//   H block
type H int

/*sqlgen: Multi-line 1*/

/*sqlgen:
  Multi-line 2.1
  Multi-line 2.2
*/

func main() {
	// Z is counted as standalone

	// sqlgen: Standalone 2
	type Z int
}
