package test

import (
	"time"
)

/*
sqlgen:
  generate User
  generate UserSubset from "user"
  generate UserInfo
  generate UserUnion
    from "user"      as u
    full join "user_info" as ui on u.id = ui.user_id
  generate UserUnionMore
    from "user"        as u
    full join "user_info"   as ui on u.id = ui.user_id
    right join "user" (UserSubset) as us on u.id = us.id
  generate ComplexInfo
  generate UserTag
  generate UserInline
*/

//go:generate bash -c "rm sql.gen.go || true"
//go:generate go install github.com/ng-vu/sqlgen/cmd/sqlgen
//go:generate sqlgen -o sql.gen.go
//go:generate goimports -w sql.gen.go

type User struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt *time.Time

	Bool    bool
	Float64 float64
	Int     int
	Int64   int64
	String  string

	PBool    *bool
	PFloat64 *float64
	PInt     *int
	PInt64   *int64
	PString  *string
}

type UserSubset struct {
	ID string

	Bool    bool
	Float64 float64
	Int     int
	Int64   int64
	String  string

	PBool    *bool
	PFloat64 *float64
	PInt     *int
	PInt64   *int64
	PString  *string
}

type UserInfo struct {
	UserID   string
	Metadata string

	Bool    bool
	Float64 float64
	Int     int
	Int64   int64
	String  string

	PBool    *bool
	PFloat64 *float64
	PInt     *int
	PInt64   *int64
	PString  *string
}

type UserUnion struct {
	User     *User
	UserInfo *UserInfo
}

type UserUnionMore struct {
	User       *User
	UserInfo   *UserInfo
	UserSubset *UserSubset
}

type ComplexInfo struct {
	ID string

	Address  Address
	PAddress *Address
	Metadata map[string]string

	Ints    []int
	Int64s  []int64
	Strings []string
	Times   []time.Time
	TimesP  []*time.Time

	// PInts    *[]int
	// PStrings *[]string
	// PTimes  *[]time.Time
	// PTimesP *[]*time.Time

	AliasString  AliasString
	AliasInt64   AliasInt64
	AliasInt     AliasInt
	AliasBool    AliasBool
	AliasFloat64 AliasFloat64
	// AliasTime    AliasTime

	AliasPString  AliasPString
	AliasPInt64   AliasPInt64
	AliasPInt     AliasPInt
	AliasPBool    AliasPBool
	AliasPFloat64 AliasPFloat64
	// AliasPTime    AliasPTime
}

type Address struct {
	Province string `json:"province"`
}

type AliasString string

type AliasInt64 int64

type AliasInt int

type AliasBool bool

type AliasFloat64 float64

type AliasTime time.Time

type AliasPString *string

type AliasPInt64 *int64

type AliasPInt *int

type AliasPBool *bool

type AliasPFloat64 *float64

type AliasPTime *time.Time

type UserTag struct {
	Skip   string  `sq:"-" json:"skip"`
	Inline Address `sq:"inline" json:"address"`
	Rename string  `sq:"'new_name'" json:"json_name"`
}

type UserInline struct {
	Inline    Address  `sq:"inline"`
	PtrInline *Address `sq:"inline"`
}
