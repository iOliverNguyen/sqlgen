// This file is not built and used for testing only.

package test

import "time"

type (
	TypeByte   byte
	TypeInt    int
	TypeString string
	TypeStruct struct{}
	TypeMap    map[string]int

	TypePByte   *byte
	TypePInt    *int
	TypePString *string
	TypePStruct *struct{}
	TypePMap    *map[string]int

	TypeInterface interface{}

	TypeSliceInt       []int
	TypeSliceAliasInt  []TypeInt
	TypeSlicePInt      []*int
	TypeSlicePAliasInt []*TypeInt

	TypePSlicePInt      *[]*int
	TypePSliceAliasPInt *[]TypePInt

	TypeSliceStruct       []struct{}
	TypeSliceAliasStruct  []TypeStruct
	TypeSlicePStruct      []*struct{}
	TypeSlicePAliasStruct []*TypeStruct

	TypeTime *time.Time
)

var (
	// basic type
	Byte   byte
	Int    int
	String string
	Struct struct{}
	Map    map[string]int

	// alias to basic type
	AliasByte   TypeByte
	AliasInt    TypeInt
	AliasString TypeString
	AliasStruct TypeStruct
	AliasMap    TypeMap

	// pointer of basic type
	PByte   *byte
	PInt    *int
	PString *string
	PStruct *struct{}
	PMap    *map[string]int

	// pointer to alias of basic type
	PAliasByte   *TypeByte
	PAliasInt    *TypeInt
	PAliasString *TypeString
	PAliasStruct *TypeStruct
	PAliasMap    *TypeMap

	// alias of pointer to basic type
	AliasPByte   TypePByte
	AliasPInt    TypePInt
	AliasPString TypePString
	AliasPStruct TypePStruct
	AliasPMap    TypePMap

	// unsupported double pointer for type: *test.TypePInt
	PAliasPInt *TypePInt

	// slice of basic type
	SliceInt    []int
	SliceStruct []struct{}
	SliceMap    []map[string]int

	// slice of alias of basic type
	SliceAliasInt    []TypeInt
	SliceAliasStruct []TypeStruct
	SliceAliasMap    []TypeMap

	// slice of pointer to basic type
	SlicePInt    []*int
	SlicePStruct []*struct{}
	SlicePMap    []*map[string]int

	// slice of pointer to alias of basic type
	SlicePAliasInt []*TypeInt

	// slice of alias of pointer to basic type
	SliceAliasPInt []TypePInt

	// unsupported double pointer for type: []*test.TypePInt
	SlicePAliasPInt []*TypePInt

	// pointer of slice of pointer to basic type
	PSlicePInt *[]*int

	// alias of slice of basic type
	AliasSliceInt TypeSliceInt

	// interface
	Interface       interface{}
	PInterface      *interface{}
	SliceInterface  []interface{}
	SlicePInterface []*interface{}
)
