// This package generates parse error: Must not mix declaration on type A and B.

package testpkg3

// sqlgen: A
type (
	A int

	B int
)
