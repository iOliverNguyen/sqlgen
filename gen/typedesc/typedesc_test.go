package typedesc

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fset     = token.NewFileSet()
	typeInfo = types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}
	typeByIdent = map[string]types.Type{}
)

func init() {
	src, err := ioutil.ReadFile("_typedesc_test.go")
	if err != nil {
		panic(err)
	}

	file, err := parser.ParseFile(fset, "_typedesc_test.go", src, 0)
	if err != nil {
		panic(err)
	}

	config := &types.Config{
		IgnoreFuncBodies: true,

		Importer: importer.Default(),
	}
	_, err = config.Check("test", fset, []*ast.File{file}, &typeInfo)
	if err != nil {
		panic(err)
	}

	for ident, obj := range typeInfo.Defs {
		if ident.Name != "test" {
			typeByIdent[ident.Name] = obj.Type()
		}
	}
}

func getType(ident string) types.Type {
	typ := typeByIdent[ident]
	if typ == nil {
		panic(fmt.Sprintf("ident not found (%v)", ident))
	}
	return typ
}

func TestKindTuple(t *testing.T) {
	var kt KindTuple
	var err error

	t.Run("basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("Byte"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(false, reflect.Uint8), kt)

		kt, err = NewKindTuple(getType("Int"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(false, reflect.Int), kt)

		kt, err = NewKindTuple(getType("String"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(false, reflect.String), kt)

		kt, err = NewKindTuple(getType("Struct"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(false, reflect.Struct), kt)

		kt, err = NewKindTuple(getType("Map"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(false, reflect.Map), kt)
	})

	t.Run("alias of basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("AliasByte"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(false, reflect.Uint8), kt)

		kt, err = NewKindTuple(getType("AliasInt"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(false, reflect.Int), kt)

		kt, err = NewKindTuple(getType("AliasString"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(false, reflect.String), kt)

		kt, err = NewKindTuple(getType("AliasStruct"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(false, reflect.Struct), kt)

		kt, err = NewKindTuple(getType("AliasMap"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(false, reflect.Map), kt)
	})

	t.Run("pointer to basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("PByte"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Uint8), kt)

		kt, err = NewKindTuple(getType("PInt"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Int), kt)

		kt, err = NewKindTuple(getType("PString"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.String), kt)

		kt, err = NewKindTuple(getType("PStruct"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Struct), kt)

		kt, err = NewKindTuple(getType("PMap"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Map), kt)
	})

	t.Run("pointer to alias of basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("PAliasByte"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Uint8), kt)

		kt, err = NewKindTuple(getType("PAliasInt"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Int), kt)

		kt, err = NewKindTuple(getType("PAliasString"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.String), kt)

		kt, err = NewKindTuple(getType("PAliasStruct"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Struct), kt)

		kt, err = NewKindTuple(getType("PAliasMap"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Map), kt)
	})

	t.Run("alias of pointer to basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("AliasPByte"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Uint8), kt)

		kt, err = NewKindTuple(getType("AliasPInt"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Int), kt)

		kt, err = NewKindTuple(getType("AliasPString"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.String), kt)

		kt, err = NewKindTuple(getType("AliasPStruct"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Struct), kt)

		kt, err = NewKindTuple(getType("AliasPMap"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Map), kt)
	})

	t.Run("pointer to alias of pointer to basic type (error)", func(t *testing.T) {
		kt, err = NewKindTuple(getType("PAliasPInt"))
		require.EqualError(t, err, "unsupported double pointer for type: *test.TypePInt")
	})

	t.Run("slice of basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("SliceInt"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, false, reflect.Int}, kt)

		kt, err = NewKindTuple(getType("SliceStruct"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, false, reflect.Struct}, kt)

		kt, err = NewKindTuple(getType("SliceMap"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, false, reflect.Map}, kt)
	})

	t.Run("slice of alias to basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("SliceAliasInt"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, false, reflect.Int}, kt)

		kt, err = NewKindTuple(getType("SliceAliasStruct"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, false, reflect.Struct}, kt)

		kt, err = NewKindTuple(getType("SliceAliasMap"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, false, reflect.Map}, kt)
	})

	t.Run("slice of pointer to basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("SlicePInt"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, true, reflect.Int}, kt)

		kt, err = NewKindTuple(getType("SlicePStruct"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, true, reflect.Struct}, kt)

		kt, err = NewKindTuple(getType("SlicePMap"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, true, reflect.Map}, kt)
	})

	t.Run("slice of pointer to alias of basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("SlicePAliasInt"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, true, reflect.Int}, kt)
	})

	t.Run("slice of alias of pointer to basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("SliceAliasPInt"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, true, reflect.Int}, kt)
	})

	t.Run("slice of pointer to alias of pointer to basic type (error)", func(t *testing.T) {
		kt, err = NewKindTuple(getType("SlicePAliasPInt"))
		require.EqualError(t, err, "unsupported double pointer for type: []*test.TypePInt")
	})

	t.Run("pointer of slice of pointer to basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("PSlicePInt"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{true, reflect.Slice, true, reflect.Int}, kt)
	})

	t.Run("alias of slice of basic type", func(t *testing.T) {
		kt, err = NewKindTuple(getType("AliasSliceInt"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, false, reflect.Int}, kt)
	})

	t.Run("interface", func(t *testing.T) {
		kt, err = NewKindTuple(getType("Interface"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(false, reflect.Interface), kt)

		kt, err = NewKindTuple(getType("PInterface"))
		require.NoError(t, err)
		assert.Equal(t, SimpleKind(true, reflect.Interface), kt)

		kt, err = NewKindTuple(getType("SliceInterface"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, false, reflect.Interface}, kt)

		kt, err = NewKindTuple(getType("SlicePInterface"))
		require.NoError(t, err)
		assert.Equal(t, KindTuple{false, reflect.Slice, true, reflect.Interface}, kt)
	})
}
