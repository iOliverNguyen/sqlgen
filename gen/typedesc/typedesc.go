package typedesc

import (
	"fmt"
	"go/types"
	"reflect"
)

// TypeDesc describes types for generating code
type TypeDesc struct {
	TypeString string
	Underlying string
	KindTuple
}

func (d *TypeDesc) IsType(t string) bool {
	return d.TypeString == t || d.Underlying == t
}

func (d *TypeDesc) IsTime() bool {
	return d.IsBareTime() || d.IsPtrTime()
}

func (d *TypeDesc) IsBareTime() bool {
	return d.IsType("time.Time")
}

func (d *TypeDesc) IsPtrTime() bool {
	return d.IsType("*time.Time")
}

func (d *TypeDesc) IsJSON() bool {
	return d.IsType("json.RawMessage")
}

func (d *TypeDesc) IsSliceOfBasicOrTime() bool {
	return d.Container == reflect.Slice && isBasic(d.Elem) ||
		d.TypeString == "[]time.Time" ||
		d.Underlying == "[]time.Time" ||
		d.TypeString == "[]*time.Time" ||
		d.Underlying == "[]*time.Time"
}

// KindTuple represents an underlying type. We do not support double pointer.
type KindTuple struct {
	Ptr       bool
	Container reflect.Kind // Only slice or array
	PtrElem   bool
	Elem      reflect.Kind // int32, int64, string, struct, map
}

// IsPtr ...
func (k KindTuple) IsPtr() bool {
	return k.Ptr || k.PtrElem
}

// IsNillable ...
func (k KindTuple) IsNillable() bool {
	return k.IsPtr() ||
		k.Container == reflect.Slice ||
		k.Elem == reflect.Map ||
		k.Elem == reflect.Slice
}

func (k KindTuple) IsPtrNumber() bool {
	return k.Ptr && k.IsNumber()
}

func (k KindTuple) IsNumber() bool {
	return isNumber(k.Elem) && k.Container == 0
}

func isNumber(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Uint64 ||
		k == reflect.Float32 || k == reflect.Float64
}

func (k KindTuple) IsPtrBasic() bool {
	return k.Ptr && k.IsBasic()
}

func (k KindTuple) IsBasic() bool {
	return isBasic(k.Elem) && k.Container == 0
}

func isBasic(k reflect.Kind) bool {
	return k == reflect.String || k == reflect.Bool ||
		isNumber(k)
}

func (k KindTuple) IsSimple() bool {
	return k.Container == 0
}

func (k KindTuple) IsSlice() bool {
	return k.Container == reflect.Slice
}

func (k KindTuple) IsKind(kind reflect.Kind) bool {
	return k.Container == kind || (k.Elem == kind && k.Container == 0)
}

func (k KindTuple) IsSimpleKind(ptr bool, kind reflect.Kind) bool {
	return k == SimpleKind(ptr, kind) && k.Container == 0
}

func (k KindTuple) IsKindTuple(kind KindTuple) bool {
	return k == kind
}

func SimpleKind(ptr bool, elem reflect.Kind) KindTuple {
	return KindTuple{Ptr: ptr, Elem: elem}
}

func NewKindTuple(typ types.Type) (kt KindTuple, err error) {
	t := UnderlyingOf(typ)
	if pt, ok := t.(*types.Pointer); ok {
		kt.Ptr = true
		t = UnderlyingOf(pt.Elem())
	}

	switch pt := t.(type) {
	case *types.Slice:
		kt.Container = reflect.Slice
		t = UnderlyingOf(pt.Elem())
	case *types.Array:
		kt.Container = reflect.Array
		t = UnderlyingOf(pt.Elem())
	}

	if pt, ok := t.(*types.Pointer); ok {
		kt.PtrElem = true
		t = UnderlyingOf(pt.Elem())
	}
	if kt.Container == 0 && kt.Ptr && kt.PtrElem {
		err = fmt.Errorf("unsupported double pointer for type: %v", typ)
		return
	}

	switch pt := t.(type) {
	case *types.Basic:
		kt.Elem = convertBasicKindToReflectKind(pt.Kind())
	case *types.Map:
		kt.Elem = reflect.Map
	case *types.Struct:
		kt.Elem = reflect.Struct
	case *types.Interface:
		kt.Elem = reflect.Interface
	case *types.Pointer:
		err = fmt.Errorf("unsupported double pointer for type: %v", typ)
	default:
		err = fmt.Errorf("unsupported type: %v", typ)
	}
	return
}

func convertBasicKindToReflectKind(k types.BasicKind) reflect.Kind {
	if k <= types.Complex128 {
		return reflect.Kind(k)
	}
	switch k {
	case types.String:
		return reflect.String
	case types.UnsafePointer:
		return reflect.UnsafePointer
	}
	panic(fmt.Sprintf("unexpected kind: %v", k))
}

func UnderlyingOf(typ types.Type) types.Type {
	for typ != typ.Underlying() {
		typ = typ.Underlying()
	}
	return typ
}
