// This package generates parse error: Must not mix declaration on type A and B.

package testpkg2

// sqlgen:
type (
	A int

	B int
)
