// This is a package
//
// sqlgen: package
// this line is ignored
package testpkg1

// sqlgen:
//   Block
//

// sqlgen: A
type A int

// sqlgen: B comment
type (
	B int
)

// This comment is ignored
type (
	// sqlgen: C
	C int
)

// sqlgen: Floating 1

// sqlgen: Floating 2
type (
	// sqlgen: D
	D int

	// E is a type
	//
	// sqlgen: E
	E int
)

// This comment is also ignored
type F int

type G int

/*sqlgen: Multi-line 1*/

/*sqlgen:
  Multi-line 2.1
  Multi-line 2.2
*/

func main() {
	// Z is counted as floating

	// sqlgen: Z
	type Z int
}
