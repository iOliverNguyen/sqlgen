// This package generates parse error: Must not mix declaration on type A and B.

package testpkg2

// sqlgen: X
type (
	// sqlgen: A
	A int

	// sqlgen: B
	B int
)
