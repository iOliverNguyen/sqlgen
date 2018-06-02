package test

import (
	"time"

	sq "github.com/ng-vu/sqlgen"
)

//go:generate ../../scripts/goderive.sh
var _ = sqlgenUser(&User{})

// User ...
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

// UserSubset ...
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

// UserInfo ...
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

// UserUnion ...
type UserUnion struct {
	User     *User
	UserInfo *UserInfo
}

var _ = sqlgenUserUnionMore(
	&UserUnionMore{}, &User{}, sq.AS("u"),
	sq.FULL_JOIN, &UserInfo{}, sq.AS("ui"), `u.id = ui.user_id`,
	sq.RIGHT_JOIN, &UserSubset{}, sq.AS("us"), `u.id = us.id`,
)

// UserUnionMore ...
type UserUnionMore struct {
	User       *User
	UserInfo   *UserInfo
	UserSubset *UserSubset
}

var _ = sqlgenComplexInfo(&ComplexInfo{})

// ComplexInfo ...
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

// Address ...
type Address struct {
	Province string `json:"province"`
}

// AliasString ...
type AliasString string

// AliasInt64 ...
type AliasInt64 int64

// AliasInt ...
type AliasInt int

// AliasBool ...
type AliasBool bool

// AliasFloat64 ...
type AliasFloat64 float64

// AliasTime ...
type AliasTime time.Time

// AliasPString ...
type AliasPString *string

// AliasPInt64 ...
type AliasPInt64 *int64

// AliasPInt ...
type AliasPInt *int

// AliasPBool ...
type AliasPBool *bool

// AliasPFloat64 ...
type AliasPFloat64 *float64

// AliasPTime ...
type AliasPTime *time.Time

var _ = sqlgenUserTag(&UserTag{})

// UserTag ...
type UserTag struct {
	Skip   string  `sq:"-" json:"skip"`
	Inline Address `sq:"inline" json:"address"`
	Rename string  `sq:"'new_name'" json:"json_name"`
}

var _ = sqlgenUserInline(&UserInline{})

// UserInline ...
type UserInline struct {
	Inline    Address  `sq:"inline"`
	PtrInline *Address `sq:"inline"`
}
