package sqlgen

import (
	"go/types"
)

type timeLevel int

const (
	timeUpdate timeLevel = 1
	timeCreate timeLevel = 2
)

func (g *Gen) appendQueryArg(b []byte, prefix string, col *colDef) []byte {
	nonNil := pathNonNil(prefix, col.pathElems)
	if nonNil != "" {
		b = appends(b, "core.Ternary(", nonNil, ", ")
	}

	path := col.Path()
	typ := col.fieldType
	typStr := g.TypeString(typ)
	zero := ""
	timeComp := ""
	if col.timeLevel == timeUpdate {
		timeComp = "true"
	} else if col.timeLevel == timeCreate {
		timeComp = "create"
	}

	b = func() []byte {
		switch typStr {
		case "time.Time":
			if col.timeLevel > 0 {
				zero = "now"
				return appends(b, "core.Now(", prefix, '.', path, ", now, ", timeComp, ")")
			}
			zero = "nil"
			return appends(b, "Time(", prefix, '.', path, ")")
		case "*time.Time":
			if col.timeLevel > 0 {
				return appends(b, "core.NowP(", prefix, '.', path, ", now, ", timeComp, ")")
			}
			zero = "nil"
			return appends(b, prefix, '.', path)
		case "json.RawMessage":
			return appends(b, "JSON{", prefix, ".", path, "}")
		}

		switch g.TypeString(typ.Underlying()) {
		case "bool":
			zero = "false"
			return appends(b, "Bool(", prefix, '.', path, ")")
		case "float64":
			zero = "0"
			return appends(b, "Float(", prefix, '.', path, ")")
		case "int":
			zero = "0"
			return appends(b, "Int(", prefix, '.', path, ")")
		case "int64":
			zero = "nil"
			return appends(b, "Int64(", prefix, '.', path, ")")
		case "string":
			zero = "nil"
			return appends(b, "String(", prefix, '.', path, ")")
		case "*bool":
			zero = "nil"
			return appends(b, prefix, '.', path)
		case "*float64":
			zero = "nil"
			return appends(b, prefix, '.', path)
		case "*int":
			zero = "nil"
			return appends(b, prefix, '.', path)
		case "*int64":
			zero = "nil"
			return appends(b, prefix, '.', path)
		case "*string":
			zero = "nil"
			return appends(b, prefix, '.', path)
		}

		switch genericType(typ, typStr) {
		case typeArray:
			zero = "nil"
			return appends(b, "Array{", prefix, '.', path, "}")
		case typeStruct:
			zero = "nil"
			return appends(b, "JSON{&", prefix, ".", path, "}")
		case typeMap, typePtrStruct, typeArrayStruct:
			zero = "nil"
			return appends(b, "JSON{", prefix, ".", path, "}")
		case typePtrArray:
			// panic
		}
		panic("Unsupported type: " + typStr)
	}()
	if nonNil != "" {
		b = appends(b, ", ", zero, ")")
	}
	return b
}

func (g *Gen) appendScanArg(b []byte, prefix string, col *colDef) []byte {
	path := col.Path()
	typ := col.fieldType
	typStr := g.TypeString(typ)

	switch typStr {
	case "time.Time":
		return appends(b, "(*Time)(&", prefix, '.', path, ")")
	case "*time.Time":
		return appends(b, '&', prefix, '.', path)
	case "json.RawMessage":
		return appends(b, "JSON{&", prefix, '.', path, "}")
	}

	switch g.TypeString(typ.Underlying()) {
	case "bool":
		return appends(b, "(*Bool)(&", prefix, '.', path, ")")
	case "float64":
		return appends(b, "(*Float)(&", prefix, '.', path, ")")
	case "int":
		return appends(b, "(*Int)(&", prefix, '.', path, ")")
	case "int64":
		return appends(b, "(*Int64)(&", prefix, '.', path, ")")
	case "string":
		return appends(b, "(*String)(&", prefix, '.', path, ")")
	case "*bool":
		return appends(b, '&', prefix, '.', path)
	case "*float64":
		return appends(b, '&', prefix, '.', path)
	case "*int":
		return appends(b, '&', prefix, '.', path)
	case "*int64":
		return appends(b, '&', prefix, '.', path)
	case "*string":
		return appends(b, '&', prefix, '.', path)
	}

	switch genericType(typ, typStr) {
	case typeArray:
		return appends(b, "Array{&", prefix, '.', path, "}")
	case typeMap, typeStruct, typePtrStruct, typeArrayStruct:
		return appends(b, "JSON{&", prefix, ".", path, "}")
	case typePtrArray:
		// panic
	}
	panic("Unsupported type: " + typStr)
}

func (g *Gen) appendUpdateArg(b []byte, prefix string, col *colDef) []byte {
	path := col.Path()
	typ := col.fieldType
	typStr := g.TypeString(typ)

	switch typStr {
	case "time.Time":
		if col.timeLevel == timeUpdate {
			return appends(b, "core.Now(", prefix, ".", path, ", now, true)")
		}
		return appends(b, prefix, ".", path)
	case "*time.Time":
		if col.timeLevel == timeUpdate {
			return appends(b, "core.NowP(", prefix, ".", path, ", now, true)")
		}
		return appends(b, "*", prefix, ".", path)
	case "json.RawMessage":
		return appends(b, "JSON{", prefix, ".", path, "}")
	}

	// TODO(qv): Handle difference between primitive string and alias string

	switch genericType(typ, typStr) {
	case typeArray:
		return appends(b, "Array{", prefix, '.', path, "}")
	case typeStruct:
		return appends(b, "JSON{&", prefix, ".", path, "}")
	case typeMap, typePtrStruct, typeArrayStruct:
		return appends(b, "JSON{", prefix, ".", path, "}")
	case typePtrArray:
		panic("Unsupported type: " + typStr)
	}

	// primitive types
	if typStr[0] == '*' {
		b = append(b, '*')
	}
	return appends(b, prefix, '.', path)
}

func pathNonNil(prefix string, path pathElems) string {
	var v string
	for _, elem := range path.BasePath() {
		if elem.ptr {
			v += prefix + "." + elem.path + `!= nil && `
		}
	}
	if v == "" {
		return ""
	}
	return v[:len(v)-4]
}

func (g *Gen) nonZero(prefix string, col *colDef) string {
	nonNil := pathNonNil(prefix, col.BasePath())
	if nonNil != "" {
		nonNil += " && "
	}
	v := prefix + "." + col.pathElems.Path()

	typ := col.fieldType
	typStr := g.TypeString(typ)

	switch typStr {
	case "time.Time":
		if col.timeLevel == timeUpdate {
			return "true"
		}
		return nonNil + "!" + v + ".IsZero() && !" + v + ".Equal(__zeroTime)"
	case "*time.Time":
		if col.timeLevel == timeUpdate {
			return "true"
		}
		return nonNil + v + " != nil"
	case "json.RawMessage":
		return nonNil + v + " != nil"
	}

	switch g.TypeString(typ.Underlying()) {
	case "bool":
		return nonNil + v
	case "float64", "int", "int64":
		return nonNil + v + " != 0"
	case "string":
		return nonNil + v + ` != ""`
	case "time.Time":
		return "!time.Time(" + v + ").IsZero() && !time.Time(" + v + ").Equal(__zeroTime)"
	case "*bool", "*float64", "*int", "*int64", "*string":
		return nonNil + v + " != nil"
	}

	switch genericType(typ, typStr) {
	case typeArray, typeArrayStruct:
		return nonNil + v + " != nil"
	case typeMap, typePtrStruct:
		return nonNil + v + " != nil"
	case typeStruct:
		return "true"
	case typePtrArray:
		// panic
	}
	panic("Unsupported type: " + typStr)
}

func (g *Gen) zero(typ types.Type) string {
	typStr := g.TypeString(typ)
	switch typStr {
	case "bool":
		return "false"
	case "float64", "int", "int64":
		return "0"
	case "time.Time":
		return "time.Time{}"
	default:
		return "nil"
	}
}

type typeGeneric int

const (
	typeInvalid typeGeneric = iota
	//typeArrayByte
	typeArrayStruct
	typeArray
	typePtrArray
	typeMap
	typeStruct
	typePtrStruct
)

func genericType(typ types.Type, typStr string) typeGeneric {
	switch typStr {
	case "[]bool", "[]float64", "[]int", "[]int64", "[]string",
		"[]time.Time", "[]*time.Time":
		return typeArray
	case "*[]bool", "*[]float64", "*[]int", "*[]int64", "*[]string",
		"*[]time.Time", "*[]*time.Time":
		return typePtrArray
	}
	if slice, ok := typ.Underlying().(*types.Slice); ok {
		elem := slice.Elem().Underlying()
		if e, ok := elem.(*types.Pointer); ok {
			elem = e.Elem().Underlying()
		}
		if _, ok := elem.(*types.Struct); ok {
			return typeArrayStruct
		}
	}
	if _, ok := typ.Underlying().(*types.Map); ok {
		return typeMap
	}
	p := false
	if _typ, ok := typ.Underlying().(*types.Pointer); ok {
		p = true
		typ = _typ.Elem()
	}
	if _, ok := typ.Underlying().(*types.Struct); ok {
		if p {
			return typePtrStruct
		}
		return typeStruct
	}
	return typeInvalid
}

func (g *Gen) bareTypeName(typ types.Type) string {
	s := g.TypeString(typ)
	if s[0] == '*' {
		return s[1:]
	}
	return s
}
