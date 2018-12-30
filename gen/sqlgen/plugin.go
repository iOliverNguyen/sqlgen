package sqlgen

import (
	"fmt"
	"go/types"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	ggen "github.com/ng-vu/sqlgen/gen"
	"github.com/ng-vu/sqlgen/gen/strs"
)

type Interface = ggen.Interface

type SubstructInterface interface {
	GenSubstruct(typ, base types.Type) error
}

var g Interface

// New is a constructor for the clone code generator.
// This generator should be reconstructed for each package.
func New(iface Interface, ss SubstructInterface) *Gen {
	g = iface
	return &Gen{
		Interface: iface,
		ss:        ss,
		mapBase:   make(map[string]bool),
		mapType:   make(map[string]*typeDef),
	}
}

const sqlTag = "sq"

type Gen struct {
	Interface
	ss SubstructInterface

	init    bool
	bases   []types.Type
	mapBase map[string]bool
	mapType map[string]*typeDef

	nAdd int
	nGen int
}

type typeDef struct {
	typ      types.Type
	base     types.Type
	cols     []*colDef
	joins    []*joinDef
	preloads []*preloadDef

	tableName string
	structs   pathElems

	all    bool
	selecT bool
	insert bool
	update bool
	delete bool

	timeLevel timeLevel
}

type colDef struct {
	ColumnName string
	FieldName  string

	fieldType  types.Type
	fieldTag   string
	columnType string
	timeLevel  timeLevel
	fkey       string
	pathElems

	exclude     bool
	_nonNilPath string
}

func (c *colDef) GenNonNilPath() string {
	if c._nonNilPath == "" {
		c._nonNilPath = genNonNilPath("m", c.pathElems)
	}
	return c._nonNilPath
}

func genNonNilPath(prefix string, path pathElems) string {
	var v string
	for _, elem := range path.BasePath() {
		if elem.ptr {
			v += prefix + "." + elem.Path + ` != nil && `
		}
	}
	if v == "" {
		return ""
	}
	return v[:len(v)-4] // remove the last " && "
}

func (c *colDef) String() string {
	return c.FieldName
}

type pathElems []pathElem

func (p pathElems) String() string {
	return p.Path()
}

func (p pathElems) Path() string {
	if p == nil {
		return "<nil>"
	}
	return p[len(p)-1].Path
}

func (p pathElems) Last() pathElem {
	return p[len(p)-1]
}

func (p pathElems) BasePath() pathElems {
	if len(p) == 0 {
		return nil
	}
	return p[:len(p)-1]
}

type pathElem struct {
	Path string
	name string
	ptr  bool
	typ  types.Type

	basePath string
	TypeName string
}

func (p pathElems) append(g *Gen, field *types.Var) pathElems {
	name := field.Name()
	typ := field.Type()
	pStr := g.TypeString(typ)
	ptr := pStr[0] == '*'
	Str := pStr
	if ptr {
		Str = pStr[1:]
	}

	elem := pathElem{
		name:     name,
		ptr:      ptr,
		typ:      typ,
		TypeName: Str,
	}
	if p == nil {
		elem.Path = name
		elem.basePath = ""
		return []pathElem{elem}
	}

	elem.Path = p.Path() + "." + name
	elem.basePath = p.Path()
	pdef := make([]pathElem, 0, len(p)+1)
	pdef = append(pdef, p...)
	pdef = append(pdef, elem)
	return pdef
}

type joinDef struct {
	JoinType types.Type
	BaseType types.Type
}

type preloadDef struct {
	FieldType     types.Type
	FieldName     string
	TableName     string
	PluralTypeStr string
	BaseType      types.Type
	Fkey          string
}

func (g *Gen) Add(name string, typs []types.Type) error {
	if len(typs) == 0 {
		return fmt.Errorf("%s must have at least one argument", name)
	}
	sTyp, ok := getStructType(typs[0])
	if !ok {
		return fmt.Errorf("Type must be struct (got %v)", typs[0].String())
	}

	cols, excols, err := g.parseColumnsFromType(nil, typs[0], sTyp)
	if err != nil {
		return err
	}

	preloads := make([]*preloadDef, len(excols))
	for i, col := range excols {
		typ := col.fieldType
		desc := GetTypeDesc(typ)
		if !desc.Ptr && desc.Container == reflect.Slice &&
			desc.PtrElem && desc.Elem == reflect.Struct {
			// continue
		} else {
			return fmt.Errorf("Preload type must be slice of pointer to struct (got %v)", desc.TypeString)
		}

		if !strings.HasPrefix(desc.TypeString, "[]*") {
			return fmt.Errorf("Only support []* for preload type")
		}
		bareTypeStr := desc.TypeString[3:]

		preload := &preloadDef{
			TableName:     toSnake(bareTypeStr),
			FieldType:     col.fieldType,
			FieldName:     col.FieldName,
			PluralTypeStr: plural(bareTypeStr),
			BaseType:      nil, // TODO
			Fkey:          col.fkey,
		}
		preloads[i] = preload
	}

	typ := typs[0]
	def := &typeDef{
		typ:      typs[0],
		all:      true,
		cols:     cols,
		preloads: preloads,
		structs:  getStructsFromCols(cols),
	}
	for _, col := range cols {
		if col.timeLevel > def.timeLevel {
			def.timeLevel = col.timeLevel
			break
		}
	}
	switch len(typs) {
	case 0:
		panic("Unexpected")
	case 1:
		g.bases = append(g.bases, typs[0])
		g.mapBase[typs[0].String()] = true
	case 2:
		def.base = typs[1]
	default:
		def.base = typs[1]
		def.all = false

		if g.TypeString(typs[2]) != "sq.AS" {
			fmt.Println(helpJoin)
			return fmt.Errorf(
				"JOIN %v: The third param must be sq.AS (got %v)",
				g.TypeString(typs[0]), g.TypeString(typs[2]))
		}

		var err error
		def.joins, err = g.parseJoin(typs[3:])
		if err != nil {
			fmt.Println(helpJoin)
			return fmt.Errorf("JOIN %v: %v", g.TypeString(typs[0]), err)
		}
	}

	if def.base != nil {
		def.tableName = strs.ToSnake(bareTypeName(def.base))
	} else {
		def.tableName = strs.ToSnake(bareTypeName(typ))
	}
	g.mapType[typs[0].String()] = def
	return nil
}

func (g *Gen) validateTypes() error {
	for _, def := range g.mapType {
		if def.base != nil {
			if !g.mapBase[def.base.String()] {
				return fmt.Errorf(
					"Type %v is based on %v but the latter is not defined as a table",
					g.TypeString(def.typ), g.TypeString(def.base))
			}
		}
	}

	// TODO: Validate join
	return nil
}

var (
	reTagColumnName = regexp.MustCompile(`'[0-9A-Za-z._-]+'`)
	reTagKeyword    = regexp.MustCompile(`\b[a-z]+\b`)
	reTagSpaces     = regexp.MustCompile(`^\s*$`)
	reTagPreload    = regexp.MustCompile(`^preload,fkey:'([0-9A-Za-z._-]+)'$`)
)

func (g *Gen) parseColumnsFromType(path pathElems, root types.Type, sTyp *types.Struct) ([]*colDef, []*colDef, error) {
	var cols, excols []*colDef
	for i, n := 0, sTyp.NumFields(); i < n; i++ {
		field := sTyp.Field(i)
		if !field.Exported() {
			continue
		}
		fieldPath := path.append(g, field)

		tag := ""
		if rawTag := reflect.StructTag(sTyp.Tag(i)); rawTag != "" {
			t, ok := rawTag.Lookup(sqlTag)
			if !ok && t != "" {
				return nil, nil, fmt.Errorf(
					"Invalid tag at `%v`.%v",
					g.TypeString(root), fieldPath)
			}
			tag = t
		}
		if strings.HasPrefix(tag, "-") {
			// Skip the field
			continue
		}

		columnName := toSnake(field.Name())
		columnType := g.TypeString(field.Type())
		inline, create, update := false, false, false
		var fkey string
		if tag != "" {
			ntag := tag
			if strings.HasPrefix(ntag, "preload") {
				parts := reTagPreload.FindStringSubmatch(ntag)
				if len(parts) == 0 {
					return nil, nil, fmt.Errorf("`preload` tag must have format \"preload,fkey:'<column>'\" (Did you forget the single quote?)")
				}
				tag = "preload"
				fkey = parts[1]
				goto endparse
			}
			if s := reTagColumnName.FindString(ntag); s != "" {
				columnName = s[1 : len(s)-1]
				ntag = strings.Replace(ntag, s, "", -1)
			}
			keywords := reTagKeyword.FindAllString(ntag, -1)
			for _, keyword := range keywords {
				switch keyword {
				case "inline":
					inline = true
				case "create", "created":
					create = true
					if columnType != "time.Time" && columnType != "*time.Time" {
						return nil, nil, fmt.Errorf("`create` flag can only be used on time.Time or *time.Time field")
					}
				case "update", "updated":
					update = true
					if columnType != "time.Time" && columnType != "*time.Time" {
						return nil, nil, fmt.Errorf("`update` flag can only be used on time.Time or *time.Time field")
					}
				default:
					return nil, nil, fmt.Errorf(
						"Unregconized keyword `%v` at `%v`.%v",
						keyword, g.TypeString(root), fieldPath)
				}
				ntag = strings.Replace(ntag, keyword, "", -1)
			}
			if !reTagSpaces.MatchString(ntag) {
				return nil, nil, fmt.Errorf(
					"Invalid tag at `%v`.%v (Did you forget the single quote?)",
					g.TypeString(root), fieldPath)
			}
		}

		if countFlags(inline, create, update) > 1 {
			return nil, nil, fmt.Errorf(
				"`inline`, `create`, `update` flags can not be used together (at `%v`.%v)", g.TypeString(root), fieldPath)
		}
		if inline {
			typ := field.Type()
			if t, ok := typ.Underlying().(*types.Pointer); ok {
				typ = t.Elem()
			}
			if t, ok := typ.Underlying().(*types.Struct); ok {
				inlineCols, inlineExCols, err := g.parseColumnsFromType(fieldPath, root, t)
				if err != nil {
					return nil, nil, err
				}
				cols = append(cols, inlineCols...)
				excols = append(excols, inlineExCols...)
				continue
			}
			return nil, nil, fmt.Errorf(
				"`inline` can only be used with struct or *struct (at `%v`.%v)", g.TypeString(root), fieldPath)
		}

	endparse:
		col := &colDef{
			FieldName:  field.Name(),
			fieldType:  field.Type(),
			fieldTag:   tag,
			ColumnName: columnName,
			columnType: columnType,
			pathElems:  fieldPath,
			fkey:       fkey,
			exclude:    tag == "preload",
		}
		if create {
			col.timeLevel = timeCreate
		} else if update {
			col.timeLevel = timeUpdate
		}
		if col.exclude {
			excols = append(excols, col)
		} else {
			cols = append(cols, col)
		}
	}
	return cols, excols, nil
}

func getStructsFromCols(cols []*colDef) (res []pathElem) {
	cpath := ""
	for _, col := range cols {
		elem := col.pathElems.Last()
		if elem.basePath == "" {
			continue
		}
		if elem.basePath == cpath {
			continue
		}
		cpath = elem.basePath
		res = append(res, col.pathElems.BasePath().Last())
	}
	return res
}

func listColumns(prefix string, cols []*colDef) string {
	b := make([]byte, 0, 1024)
	for i, col := range cols {
		if i > 0 {
			b = append(b, `,`...)
		}
		if prefix != "" {
			b = append(b, prefix...)
			b = append(b, '.')
		}
		b = append(b, '"')
		b = append(b, col.ColumnName...)
		b = append(b, '"')
	}
	return string(b)
}

func listInsertArgs(cols []*colDef) []string {
	res := make([]string, len(cols))
	for i, col := range cols {
		res[i] = genInsertArg(col)
	}
	return res
}

func listScanArgs(cols []*colDef) []string {
	res := make([]string, len(cols))
	for i, col := range cols {
		res[i] = genScanArg(col)
	}
	return res
}

func (g *Gen) tableName(def *typeDef) string {
	typ := def.typ
	if def.base != nil {
		typ = def.base
	}
	name := g.TypeString(typ)[1:]
	return toSnake(name)
}

func (g *Gen) tableNameOf(typ types.Type) string {
	def := g.mapType[typ.String()]
	return g.tableName(def)
}

func (g *Gen) genConvertMethodsFor(typ, base types.Type) error {
	return g.ss.GenSubstruct(typ, base)
}

const helpJoin = `
    JOIN must have syntax: JoinType BaseType Condition

		JoinType  : One of JOIN, FULL_JOIN, LEFT_JOIN, RIGHT_JOIN,
                    NATUAL_JOIN, SELF_JOIN, CROSS_JOIN
		BaseType  : Must be a selectable struct.
		Condition : The join condition.
                    Use $L and $R as placeholders for table name.

	Example:
        sqlgenUserFullInfo(
            &UserFullInfo{}, &User{}, sq.AS("u"),
            sq.FULL_JOIN, &UserInfo{}, sq.AS("ui"), "$L.id = $R.user_id",
        )
        type UserFullInfo struct {
            User     *User
            UserInfo *UserInfo
        }
`

func (g *Gen) parseJoin(typs []types.Type) (joins []*joinDef, err error) {
	if len(typs)%4 != 0 {
		return nil, fmt.Errorf("Invalid join definition")
	}
	for i := 0; i < len(typs); i = i + 4 {
		join, err := g.parseJoinLine(typs[i:])
		if err != nil {
			return nil, err
		}
		joins = append(joins, join)
	}
	return joins, nil
}

func (g *Gen) parseJoinLine(typs []types.Type) (*joinDef, error) {
	if g.TypeString(typs[0]) != "core.JoinType" {
		return nil, fmt.Errorf("Invalid JoinType: must be one of predefined constants (got %v)", g.TypeString(typs[0]))
	}

	base := typs[1]
	if _, ok := getStructType(base); !ok {
		return nil, fmt.Errorf(
			"Invalid base type for join: must be pointer to struct (got %v)",
			g.TypeString(base))
	}

	as := typs[2]
	if g.TypeString(as) != "sq.AS" {
		return nil, fmt.Errorf(
			"Invalid AS: must be sq.AS (got %v)", g.TypeString(as))
	}

	cond := typs[3]
	if g.TypeString(cond) != "string" {
		return nil, fmt.Errorf(
			"Invalid condition for join: must be string (got %v)",
			g.TypeString(cond))
	}

	return &joinDef{
		JoinType: base,
	}, nil
}

func getStructType(typ types.Type) (*types.Struct, bool) {
	pt, ok := typ.Underlying().(*types.Pointer)
	if ok {
		typ = pt.Elem()
	}
	st, ok := typ.Underlying().(*types.Struct)
	return st, ok
}

func isPointer(typ types.Type) bool {
	_, ok := typ.Underlying().(*types.Pointer)
	return ok
}

func capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}

func plural(s string) string {
	return strs.Plural(2, s, "")
}

func appends(b []byte, args ...interface{}) []byte {
	for _, arg := range args {
		switch arg := arg.(type) {
		case byte:
			b = append(b, arg)
		case rune:
			b = append(b, byte(arg))
		case string:
			b = append(b, arg...)
		case []byte:
			b = append(b, arg...)
		case int:
			b = strconv.AppendInt(b, int64(arg), 10)
		case int64:
			b = strconv.AppendInt(b, arg, 10)
		default:
			panic("Unsupport arg type: " + reflect.TypeOf(arg).Name())
		}
	}
	return b
}

func toSnake(s string) string {
	return strs.ToSnake(s)
}

func countFlags(args ...bool) int {
	c := 0
	for _, arg := range args {
		if arg {
			c++
		}
	}
	return c
}
