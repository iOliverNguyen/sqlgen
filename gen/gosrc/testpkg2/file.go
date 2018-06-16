// This package generates parse error: Must not mix declaration on type A and B.

package testpkg2

// sqlgen: A
type (
	A int

	// sqlgen: B
	B int
)
