package mock

import (
	"fmt"

	sq "github.com/ng-vu/sqlgen"
)

type ErrorMock struct {
	Err    error
	Entry  *sq.LogEntry
	Called int
}

type Error struct {
	Err   error
	Entry *sq.LogEntry
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (m *ErrorMock) Reset() {
	fmt.Println()
	m.Err = nil
	m.Entry = nil
	m.Called = 0
}

func (m *ErrorMock) Mock(err error, entry *sq.LogEntry) error {
	m.Called++
	m.Err, m.Entry = err, entry

	if err == nil {
		return nil
	}
	return &Error{err, entry}
}
