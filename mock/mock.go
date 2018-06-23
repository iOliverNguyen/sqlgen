package mock

import (
	"fmt"

	"reflect"
	"testing"

	sq "github.com/ng-vu/sqlgen/typesafe/sq"
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

func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Expect no error. Got: %v", err)
		t.FailNow()
	}
}

func AssertErrorEqual(t *testing.T, err error, expect string) {
	if err == nil || err.Error() != expect {
		t.Errorf("Expect error equal to `%v`. Got: %v", expect, err)
		t.FailNow()
	}
}

func AssertEqual(t *testing.T, actual, expect interface{}) {
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("\nExpect:\n`%v`\nGot:\n`%v`", expect, actual)
		t.FailNow()
	}
}
