package mock

import (
	"fmt"

	sq "github.com/ng-vu/sqlgen"
)

// ErrorMock ...
type ErrorMock struct {
	Err    error
	Entry  *sq.LogEntry
	Called int
}

// Error ...
type Error struct {
	Err   error
	Entry *sq.LogEntry
}

func (e *Error) Error() string {
	return e.Err.Error()
}

// Reset ...
func (m *ErrorMock) Reset() {
	fmt.Println()
	m.Err = nil
	m.Entry = nil
	m.Called = 0
}

// Mock ...
func (m *ErrorMock) Mock(err error, entry *sq.LogEntry) error {
	m.Called++
	m.Err, m.Entry = err, entry

	if err == nil {
		return nil
	}
	return &Error{err, entry}
}
