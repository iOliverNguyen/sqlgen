package dsl_test

import (
	"reflect"
	"testing"

	. "github.com/ng-vu/sqlgen/core/dsl"
)

func TestDSL(t *testing.T) {
	t.Run("Full simple declaration", func(t *testing.T) {
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

	t.Run("Spacing", func(t *testing.T) {
		src := `generate Account()from"account"`
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), `generate Account from "account";`+"\n")
		assertEqual(t, len(file.Declarations), 1)
	})

	t.Run("Simplified 1", func(t *testing.T) {
		src := `generate`
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), `generate {} from "{}";`+"\n")
	})

	t.Run("Simplified 2", func(t *testing.T) {
		src := `generate Account`
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), `generate Account from "{}";`+"\n")
	})

	t.Run("Simplified 3", func(t *testing.T) {
		src := `generate from account`
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), `generate {} from "account";`+"\n")
	})

	t.Run("Simplified 4", func(t *testing.T) {
		src := `generate from "account"`
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), `generate {} from "account";`+"\n")
	})

	t.Run("Simplified with options", func(t *testing.T) {
		src := `generate (plural Accounts)`
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), `generate {} (plural Accounts) from "{}";`+"\n")
	})

	t.Run("Commonly use", func(t *testing.T) {
		src := `generate Account from account`
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), `generate Account from "account";`+"\n")
	})

	t.Run("Empty option", func(t *testing.T) {
		src := `generate Account () from "account"`
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), `generate Account from "account";`+"\n")
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

	t.Run("Auto semicolon insertion", func(t *testing.T) {
		src := `generate generate Account generate from account`
		expected := `
generate {} from "{}";
generate Account from "{}";
generate {} from "account";
`[1:]
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), expected)
		assertEqual(t, len(file.Declarations), 3)
	})
}

func TestJoin(t *testing.T) {
	t.Run("Full syntax", func(t *testing.T) {
		src := `
generate UserJoinAccount
	from "user"    (User)    as u
	join "account" (Account) as a on u.id = a.user_id
`
		expected := `
generate UserJoinAccount
    from "user" (User) as u
    join "account" (Account) as a on u.id = a.user_id;
`[1:]
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), expected)
		assertEqual(t, len(file.Declarations[0].Joins), 2)
	})

	t.Run("Full syntax with 3 joins", func(t *testing.T) {
		src := `
generate UserJoinAccount
	from "user"         (User)        as u
	join "account_user" (AccountUser) as au on u.id = au.user_id
	join "account"      (Account)     as a  on a.id = au.account_id;
`
		expected := `
generate UserJoinAccount
    from "user" (User) as u
    join "account_user" (AccountUser) as au on u.id = au.user_id
    join "account" (Account) as a on a.id = au.account_id;
`[1:]
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), expected)
		assertEqual(t, len(file.Declarations[0].Joins), 3)
	})

	t.Run("Simplied, keep spacing and double quotes", func(t *testing.T) {
		src := `
generate UserJoinAccount
	from user
	join account_user on "user".id  = account_user.user_id
	join account      on account.id = account_user.account_id
`
		expected := `
generate UserJoinAccount
    from "user"
    join "account_user" on "user".id  = account_user.user_id
    join "account" on account.id = account_user.account_id;
`[1:]
		file, err := ParseString("test", src)
		assertNoError(t, err)
		assertEqual(t, file.String(), expected)
		assertEqual(t, len(file.Declarations[0].Joins), 3)
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
		t.Errorf("\nExpect:\n`%v`\nGot:\n`%v`", expect, actual)
	}
}
