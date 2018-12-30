package sqlgen

import (
	"go/types"
	"reflect"

	"github.com/ng-vu/sqlgen/gen/typedesc"
)

type TypeDesc = typedesc.TypeDesc
type timeLevel int

const (
	timeUpdate timeLevel = 1
	timeCreate timeLevel = 2
)

var basicWrappers = []string{
	reflect.Bool:    "Bool",
	reflect.Int:     "Int",
	reflect.Int8:    "Int8",
	reflect.Int16:   "Int16",
	reflect.Int32:   "Int32",
	reflect.Int64:   "Int64",
	reflect.Uint:    "Uint",
	reflect.Uint8:   "Uint8",
	reflect.Uint16:  "Uint16",
	reflect.Uint32:  "Uint32",
	reflect.Uint64:  "Uint64",
	reflect.Float32: "Float32",
	reflect.Float64: "Float64",
	reflect.String:  "String",
}

var typeMap = map[types.Type]*TypeDesc{}

func GetTypeDesc(typ types.Type) *TypeDesc {
	desc := typeMap[typ]
	if desc == nil {
		kt, err := typedesc.NewKindTuple(typ)
		if err != nil {
			panic(err)
		}
		desc = &TypeDesc{
			TypeString: g.TypeString(typ),
			Underlying: g.TypeString(typ.Underlying()),
			KindTuple:  kt,
		}
		typeMap[typ] = desc
	}
	return desc
}

func GenScanArg(path string, typ types.Type) string {
	return genScanArg2(path, typ)
}

func genScanArg(col *colDef) string {
	path := "m." + col.Path()
	return genScanArg2(path, col.fieldType)
}

func genScanArg2(path string, typ types.Type) string {
	desc := GetTypeDesc(typ)
	switch {
	case desc.IsPtrTime():
		return "&" + path

	case desc.IsBareTime():
		return "(*core.Time)(&" + path + ")"

	case desc.IsJSON():
		return "core.JSON{&" + path + "}"

	case desc.IsPtrBasic():
		return "&" + path

	case desc.IsBasic(): // && !IsPtrBasic()
		return "(*core." + basicWrappers[desc.Elem] + ")(&" + path + ")"

	case desc.IsSliceOfBasicOrTime():
		return "core.Array{&" + path + ", opts}"

	case
		desc.IsSimpleKind(false, reflect.Struct),
		desc.IsSimpleKind(true, reflect.Struct),
		desc.IsSimpleKind(false, reflect.Map),
		desc.IsSlice(): // && !desc.IsSliceOfBasicOrTime()
		return "core.JSON{&" + path + "}"
	}

	panic("unsupported type: " + desc.TypeString)
}

func genInsertArg(col *colDef) string {
	path := "m." + col.Path()
	res := genInsertArg2(path, col.fieldType, col.timeLevel)

	nonNilPath := col.GenNonNilPath()
	if nonNilPath == "" {
		return res
	}
	return "core.Ternary(" + nonNilPath + "," + res + ", nil)"
}

func genInsertArg2(path string, typ types.Type, timeLevel timeLevel) string {
	desc := GetTypeDesc(typ)
	switch {
	case desc.IsPtrTime():
		if timeLevel > 0 {
			timeComp := getTimeComp(timeLevel)
			return "core.NowP(" + path + ", now, " + timeComp + ")"
		}
		return path

	case desc.IsBareTime():
		if timeLevel > 0 {
			timeComp := getTimeComp(timeLevel)
			return "core.Now(" + path + ", now, " + timeComp + ")"
		}
		return "core.Time(" + path + ")"

	case desc.IsJSON():
		return "core.JSON{" + path + "}"

	case desc.IsPtrBasic():
		return path

	case desc.IsBasic(): // && !desc.IsBasic()
		return "core." + basicWrappers[desc.Elem] + "(" + path + ")"

	case desc.IsSliceOfBasicOrTime():
		return "core.Array{" + path + ", opts}"

	case desc.IsSimpleKind(false, reflect.Struct):
		return "core.JSON{&" + path + "}"

	case
		desc.IsSimpleKind(true, reflect.Struct),
		desc.IsSimpleKind(false, reflect.Map),
		desc.IsSlice(): // && !desc.IsSliceOfBasicOrTime()
		return "core.JSON{" + path + "}"
	}

	panic("unsupported type: " + desc.TypeString)
}

func getTimeComp(timeLevel timeLevel) string {
	switch timeLevel {
	case timeUpdate:
		return "true"
	case timeCreate:
		return "create"
	}
	panic("unexpected")
}

func genUpdateArg(col *colDef) string {
	path := "m." + col.Path()
	return genUpdateArg2(path, col.fieldType, col.timeLevel)
}

func genUpdateArg2(path string, typ types.Type, timeLevel timeLevel) string {
	desc := GetTypeDesc(typ)
	switch {
	case desc.IsPtrTime():
		if timeLevel == timeUpdate {
			return "core.NowP(" + path + ", time.Now(), true)"
		}
		return "*" + path

	case desc.IsBareTime():
		if timeLevel == timeUpdate {
			return "core.Now(" + path + ", time.Now(), true)"
		}
		return path

	case desc.IsJSON():
		return "core.JSON{" + path + "}"

	case desc.IsPtrBasic():
		if desc.TypeString == desc.Underlying {
			return "*" + path
		}
		return "(" + desc.Underlying + ")(" + path + ")"

	case desc.IsBasic(): // && !desc.IsPtrBasic()
		if desc.TypeString == desc.Underlying {
			return path
		}
		return desc.Underlying + "(" + path + ")"

	case desc.IsSliceOfBasicOrTime():
		return "core.Array{" + path + ", opts}"

	case desc.IsSimpleKind(false, reflect.Struct):
		return "core.JSON{&" + path + "}"

	case
		desc.IsSimpleKind(true, reflect.Struct),
		desc.IsSimpleKind(false, reflect.Map),
		desc.IsSlice(): // && !desc.IsSliceOfBasicOrTime()
		return "core.JSON{" + path + "}"
	}

	panic("unsupported type: " + desc.TypeString)
}

func genIfNotEqualToZero(col *colDef) string {
	path := "m." + col.pathElems.Path()
	res := genNotEqualToZero(path, col.fieldType)

	nonNilPath := col.GenNonNilPath()
	if nonNilPath == "" {
		return res
	}
	return nonNilPath + " && " + res
}

func genNotEqualToZero(path string, typ types.Type) string {
	desc := GetTypeDesc(typ)
	switch {
	case desc.IsBareTime():
		return "!" + path + ".IsZero()"
	case desc.IsNillable():
		return path + " != nil"
	case desc.IsNumber():
		return path + " != 0"
	case desc.IsKind(reflect.Bool):
		return path
	case desc.IsKind(reflect.String):
		return path + ` != ""`
	case desc.IsKind(reflect.Struct):
		return "true"
	}

	panic("unsupported type: " + desc.TypeString)
}

func genZeroValue(typ types.Type) string {
	desc := GetTypeDesc(typ)
	switch {
	case desc.IsBareTime():
		return "time.Time{}"
	case desc.IsNillable():
		return "nil"
	case desc.IsNumber():
		return "0"
	case desc.IsKind(reflect.Bool):
		return "false"
	case desc.IsKind(reflect.String):
		return `""`
	case desc.IsKind(reflect.Struct):
		return desc.TypeString + "{}"
	}

	panic("unsupported type: " + desc.TypeString)
}

func bareTypeName(typ types.Type) string {
	s := g.TypeString(typ)
	if s[0] == '*' {
		return s[1:]
	}
	return s
}
