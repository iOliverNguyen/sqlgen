package test

import (
	"time"

	sq "github.com/ng-vu/sqlgen/typesafe/sq"
)

/*
sqlgen:
  generate User
  //generate UserSubset from "user" (User)
  generate UserInfo
  generate UserUnion
    from "user"      as u
    join "user_info" as ui on u.ui = ui.user_id
  generate UserUnionMore
    from "user"        as u
    join "user_info"   as ui on u.ui = ui.user_id
    join "user_subset" as us on u.ui = us.ui
  generate ComplexInfo
  generate UserTag
  generate UserInline
*/

//go:generate ../../scripts/goderive.sh
var _ = sqlgenUser(&User{})

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

// Generate a struct (B) represents part of the given table (A).
//
//    sqlgen...(B, A)
var _ = sqlgenUserSubset(&UserSubset{}, &User{})

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

var _ = sqlgenUserInfo(&UserInfo{})

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

// Generate a struct (U) which represents join between table (A) and (B)
//
//    sqlgen...(U, A, join, B, condition)
var _ = sqlgenUserUnion(
	&UserUnion{}, &User{}, sq.AS("u"),
	sq.FULL_JOIN, &UserInfo{}, sq.AS("ui"), `u.id = ui.user_id`,
)

type UserUnion struct {
	User     *User
	UserInfo *UserInfo
}

var _ = sqlgenUserUnionMore(
	&UserUnionMore{}, &User{}, sq.AS("u"),
	sq.FULL_JOIN, &UserInfo{}, sq.AS("ui"), `u.id = ui.user_id`,
	sq.RIGHT_JOIN, &UserSubset{}, sq.AS("us"), `u.id = us.id`,
)

type UserUnionMore struct {
	User       *User
	UserInfo   *UserInfo
	UserSubset *UserSubset
}

var _ = sqlgenComplexInfo(&ComplexInfo{})

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

var _ = sqlgenUserTag(&UserTag{})

type UserTag struct {
	Skip   string  `sq:"-" json:"skip"`
	Inline Address `sq:"inline" json:"address"`
	Rename string  `sq:"'new_name'" json:"json_name"`
}

var _ = sqlgenUserInline(&UserInline{})

type UserInline struct {
	Inline    Address  `sq:"inline"`
	PtrInline *Address `sq:"inline"`
}
