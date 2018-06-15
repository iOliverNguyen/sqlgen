package dsl_test

import (
	"reflect"
	"testing"

	. "github.com/ng-vu/sqlgen/core/dsl"
)

func TestDSL(t *testing.T) {
	t.Run("Full syntax", func(t *testing.T) {
		src := `generate Account (plural Accounts) from "account";`
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), src+"\n")
		assertEqual(t, len(file.Declarations), 1)
	})

	t.Run("Syntax error", func(t *testing.T) {
		src := `generate Account (plural Accounts from "account";`
		_, err := ParseString("test", src)
		assertErrorEqual(t, err, "Error at test:1:35: syntax error")
	})

	t.Run("Simple", func(t *testing.T) {
		src := `generate Account from account`
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), "generate Account from \"account\";\n")
		assertEqual(t, len(file.Declarations), 1)
	})

	t.Run("Multiple declarations", func(t *testing.T) {
		src := `
generate Account from account;
generate User (plural Users) from "user"
`
		expected := `
generate Account from "account";
generate User (plural Users) from "user";
`[1:]
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), expected)
		assertEqual(t, len(file.Declarations), 2)
	})
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Expect no error. Got: %v", err)
		t.FailNow()
	}
}

func assertErrorEqual(t *testing.T, err error, expect string) {
	if err == nil || err.Error() != expect {
		t.Errorf("Expect error equal to `%v`. Got: %v", expect, err)
	}
}

func assertEqual(t *testing.T, actual, expect interface{}) {
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("Expect `%v`. Got: %v", expect, actual)
	}
}
