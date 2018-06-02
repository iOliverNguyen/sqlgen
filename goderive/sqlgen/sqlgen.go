package sqlgen

import (
	"fmt"
	"go/types"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	. "github.com/ng-vu/sqlgen/core/strs"
	"github.com/ng-vu/sqlgen/goderive/substruct"

	"github.com/awalterschulze/goderive/derive"
)

// NewPlugin creates a new sqlgen plugin.
// This function returns the plugin name, default prefix and a constructor for the clone code generator.
func NewPlugin() derive.Plugin {
	return derive.NewPlugin("sqlgen", "sqlgen", New)
}

var gt derive.TypesMap

// New is a constructor for the clone code generator.
// This generator should be reconstructed for each package.
func New(typesMap derive.TypesMap, p derive.Printer, deps map[string]derive.Dependency) derive.Generator {
	gt = typesMap
	return &gen{
		TypesMap: typesMap,
		printer:  p,
		mapBase:  make(map[string]bool),
		mapType:  make(map[string]*typeDef),
	}
}

const sqlTag = "sq"

type gen struct {
	derive.TypesMap
	printer derive.Printer

	bases   []types.Type
	mapBase map[string]bool
	mapType map[string]*typeDef

	init bool
}

type typeDef struct {
	typ   types.Type
	base  types.Type
	cols  []*colDef
	joins []*joinDef

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
	fieldName  string
	fieldType  types.Type
	fieldTag   string
	columnName string
	columnType string
	timeLevel  timeLevel

	pathElems
}

func (c *colDef) String() string {
	return c.fieldName
}

type pathElems []pathElem

func (p pathElems) String() string {
	return p.Path()
}

func (p pathElems) Path() string {
	if p == nil {
		return "<nil>"
	}
	return p[len(p)-1].path
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
	path string
	name string
	ptr  bool
	typ  types.Type

	basePath string
	typName  string
}

func (p pathElems) append(field *types.Var) pathElems {
	name := field.Name()
	typ := field.Type()
	pStr := gt.TypeString(typ)
	ptr := pStr[0] == '*'
	Str := pStr
	if ptr {
		Str = pStr[1:]
	}

	elem := pathElem{
		name:    name,
		ptr:     ptr,
		typ:     typ,
		typName: Str,
	}
	if p == nil {
		elem.path = name
		elem.basePath = ""
		return []pathElem{elem}
	}

	elem.path = p.Path() + "." + name
	elem.basePath = p.Path()
	pdef := make([]pathElem, 0, len(p)+1)
	pdef = append(pdef, p...)
	pdef = append(pdef, elem)
	return pdef
}

type joinDef struct {
	joinTyp types.Type
	base    types.Type
}

func (g *gen) Add(name string, typs []types.Type) (string, error) {
	if len(typs) == 0 {
		return "", fmt.Errorf("%s must have at least one argument", name)
	}
	sTyp, ok := pointerToStruct(typs[0])
	if !ok {
		return "", fmt.Errorf("Type must be pointer to struct (got %v)", typs[0].String())
	}

	cols, err := parseColumnsFromType(nil, typs[0], sTyp)
	if err != nil {
		return "", err
	}

	typ := typs[0]
	def := &typeDef{
		typ:     typs[0],
		all:     true,
		cols:    cols,
		structs: getStructsFromCols(cols),
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
			return "", fmt.Errorf(
				"JOIN %v: The third param must be sq.AS (got %v)",
				g.TypeString(typs[0]), g.TypeString(typs[2]))
		}

		var err error
		def.joins, err = g.parseJoin(typs[3:])
		if err != nil {
			fmt.Println(helpJoin)
			return "", fmt.Errorf("JOIN %v: %v", g.TypeString(typs[0]), err)
		}
	}

	if def.base != nil {
		def.tableName = ToSnake(gt.TypeString(def.base)[1:])
	} else {
		def.tableName = ToSnake(gt.TypeString(typ)[1:])
	}
	g.mapType[typs[0].String()] = def
	return g.SetFuncName(name, typs[0])
}

func (g *gen) Generate(typs []types.Type) error {
	if err := g.validateTypes(); err != nil {
		return err
	}

	g.generateCommon()
	return g.genQueryFor(typs[0])
}

func (g *gen) validateTypes() error {
	for _, def := range g.mapType {
		if def.base != nil {
			if !g.mapBase[def.base.String()] {
				return fmt.Errorf(
					"Type %v is based on %v but the latter is not defined as a table",
					gt.TypeString(def.typ), gt.TypeString(def.base))
			}
		}
	}

	// TODO: Validate join
	return nil
}

func (g *gen) generateCommon() {
	if g.init {
		return
	}
	g.init = true

	p := g.printer
	p.NewImport("sq", "github.com/ng-vu/sqlgen")()
	p.NewImport("core", "github.com/ng-vu/sqlgen/core")()
	p.NewImport("", "database/sql")()

	tpl := `
type (
	Array   = core.Array
	Bool    = core.Bool
	Float   = core.Float
	Int     = core.Int
	Int64   = core.Int64
	JSON    = core.JSON
	String  = core.String
	Time    = core.Time
	IState  = core.IState
)

var __zeroTime = time.Unix(0,0)
`
	p.P(strings.TrimRight(tpl, " \n"))
}

func (g *gen) genQueryFor(typ types.Type) error {
	p := g.printer
	g.Generating(typ)
	def := g.mapType[typ.String()]
	pStr := gt.TypeString(typ)
	Str := pStr[1:]
	Strs := ToPlural(Str)
	tableName := def.tableName

	_ListCols := fmt.Sprintf("__sql%v_ListCols", Str)
	_Table := fmt.Sprintf("__sql%v_Table", Str)
	_Insert := fmt.Sprintf("__sql%v_Insert", Str)
	_Select := fmt.Sprintf("__sql%v_Select", Str)
	_UpdateAll := fmt.Sprintf("__sql%v_UpdateAll", Str)
	_JoinTypes := fmt.Sprintf("__sql%v_JoinTypes", Str)
	_Join := fmt.Sprintf("__sql%v_Join", Str)
	_JoinConds := fmt.Sprintf("__sql%v_JoinConds", Str)
	_As := fmt.Sprintf("__sql%v_As", Str)
	_JoinAs := fmt.Sprintf("__sql%v_JoinAs", Str)

	_ListColsForType := func(typ types.Type) string {
		return fmt.Sprintf("__sql%v_ListCols", gt.TypeString(typ)[1:])
	}
	_TableForType := func(typ types.Type) string {
		return fmt.Sprintf("__sql%v_Table", gt.TypeString(typ)[1:])
	}

	{
		extra := ""
		if def.base != nil {
			extra = ", _ " + gt.TypeString(def.base)
		}
		if len(def.joins) == 0 {
			p.P("")
			p.P(`// Type %v represents table "%v"`, pStr, tableName)
			p.P("func %v(_ %v%v) bool { return true }", g.GetFuncName(typ), pStr, extra)

		} else {
			extra += ", as sq.AS"
			joinTypes := make([]string, len(def.joins))
			joinAs := make([]string, len(def.joins))
			joinConds := make([]string, len(def.joins))
			for i, join := range def.joins {
				extra += fmt.Sprintf(
					", t%v sq.JOIN_TYPE, _ %v, a%v sq.AS, c%v string",
					i, gt.TypeString(join.joinTyp), i, i)
				joinTypes[i] = fmt.Sprintf("t%v", i)
				joinAs[i] = fmt.Sprintf(`a%v+"."`, i)
				joinConds[i] = fmt.Sprintf("c%v", i)
			}

			p.P("")
			p.P(`// Type %v represents a join`, pStr)
			p.P("func %v(_ %v%v) bool {", g.GetFuncName(typ), pStr, extra)
			p.In()
			p.P("%v = []sq.JOIN_TYPE{%v}", _JoinTypes, strings.Join(joinTypes, ","))
			p.P(`%v = as + "."`, _As)
			p.P("%v = []sq.AS{%v}", _JoinAs, strings.Join(joinAs, ","))
			p.P("%v = []string{%v}",
				_JoinConds, strings.Join(joinConds, ","))
			p.P("%v = (%v)(nil).__sqlSelect(make([]byte, 0, 1024))",
				_Select, pStr)
			p.P("%v = (%v)(nil).__sqlJoin(make([]byte, 0, 1024), %v)",
				_Join, pStr, _JoinTypes)
			p.P("return true")
			p.Out()
			p.P("}")
		}
		p.P("")
		p.P("type %v []%v", Strs, pStr)
		p.P("")
	}
	if def.base != nil && len(def.joins) == 0 {
		if err := g.genConvertMethodsFor(def.typ, def.base); err != nil {
			return err
		}
	}
	if len(def.joins) == 0 {
		p.P("const %v = `%v`", _Table, tableName)

		cols := g.listColumns("", def.cols)
		p.P("const %v = `%v`", _ListCols, string(cols))

		p.P("const %v = `INSERT INTO \"%v\" (` + %v + `) VALUES`",
			_Insert, tableName, _ListCols)
		p.P("const %v = `SELECT ` + %v + ` FROM \"%v\"`", _Select, _ListCols, tableName)
		p.P("const %v = `UPDATE \"%v\" SET (` + %v + `)`", _UpdateAll, tableName, _ListCols)
	}
	if len(def.joins) > 0 {
		p.P("var %v []sq.JOIN_TYPE", _JoinTypes)
		p.P("var %v sq.AS", _As)
		p.P("var %v []sq.AS", _JoinAs)
		p.P("var %v []string", _JoinConds)
		p.P("var %v, %v []byte", _Select, _Join)
	}
	{
		p.P("")
		p.P("func (m %v) SQLTableName() string {", pStr)
		p.In()
		p.P(`return "%v"`, tableName)
		p.Out()
		p.P("}")

		p.P("")
		p.P("func (m %v) SQLTableName() string {", Strs)
		p.In()
		p.P(`return "%v"`, tableName)
		p.Out()
		p.P("}")
	}
	if def.all || def.insert || def.update {
		args := g.listQueryArgs("m", def.cols)
		p.P("")
		p.P("func (m %v) SQLArgs(args []interface{}, create bool) []interface{} {", pStr)
		p.In()
		if def.timeLevel > 0 {
			p.P("now := time.Now()")
		}
		p.P("return append(args,")
		p.In()
		for _, arg := range args {
			p.P("%v,", arg)
		}
		p.Out()
		p.P(")")
		p.Out()
		p.P("}")
	}
	if def.all || def.selecT {
		p.P("")
		p.P("func (_ %v) SQLSelect(b []byte) []byte {", pStr)
		p.In()
		p.P("return append(b, %v...)", _Select)
		p.Out()
		p.P("}")
	}
	if def.all || def.selecT {
		p.P("")
		p.P("func (_ *%v) SQLSelect(b []byte) []byte {", Strs)
		p.In()
		p.P("return append(b, %v...)", _Select)
		p.Out()
		p.P("}")
	}
	if def.all || def.insert {
		p.P("")
		p.P("func (m %v) SQLInsert(s IState, b []byte, args []interface{}) ([]byte, []interface{}, error) {", pStr)
		p.In()
		p.P("b = append(b, %v...)", _Insert)
		p.P("b = append(b, ` (`...)")
		p.P("b = s.AppendMarker(b, %v)", len(def.cols))
		p.P("b = append(b, ')')")
		p.P("return b, m.SQLArgs(args, true), nil")
		p.Out()
		p.P("}")
	}
	if def.all || def.insert {
		p.P("")
		p.P("func (ms %v) SQLInsert(s IState, b []byte, args []interface{}) ([]byte, []interface{}, error) {", Strs)
		p.In()
		p.P("b = append(b, %v...)", _Insert)
		p.P("b = append(b, ` (`...)")
		p.P("for i := 0; i < len(ms); i++ {")
		p.In()
		p.P("if i > 0 {")
		p.In()
		p.P("b = append(b, `),(`...)")
		p.Out()
		p.P("}")
		p.P("b = s.AppendMarker(b, %v)", len(def.cols))
		p.P("args = ms[i].SQLArgs(args, true)")
		p.Out()
		p.P("}")
		p.P("b = append(b, ')')")
		p.P("return b, args, nil")
		p.Out()
		p.P("}")
	}
	if def.all || def.selecT {
		var ptrStructs []pathElem
		for _, s := range def.structs {
			if s.ptr {
				ptrStructs = append(ptrStructs, s)
			}
		}

		args := g.listScanArgs("m", def.cols)
		p.P("")
		p.P("func (m %v) SQLScanArgs(args []interface{}) []interface{} {", pStr)
		p.In()
		for _, item := range ptrStructs {
			p.P("m.%v = new(%v)", item.path, item.typName)
		}
		p.P("return append(args,")
		p.In()
		for _, arg := range args {
			p.P("%v,", arg)
		}
		p.Out()
		p.P(")")
		p.Out()
		p.P("}")
	}
	if def.all || def.selecT || len(def.joins) > 0 {
		p.P("")
		p.P("func (m %v) SQLScan(row *sql.Row) error {", pStr)
		p.In()
		p.P("args := make([]interface{}, 0, 64)")
		p.P("return row.Scan(m.SQLScanArgs(args)...)")
		p.Out()
		p.P("}")

		p.P("")
		p.P("func (ms *%v) SQLScan(rows *sql.Rows) error {", Strs)
		p.In()
		p.P("res := make(%v, 0, 128)", Strs)
		p.P("args := make([]interface{}, 0, 64)")
		p.P("for rows.Next() {")
		p.In()
		p.P("m := new(%v)", Str)
		p.P("args = args[:0]")
		p.P("args = m.SQLScanArgs(args)")
		p.P("if err := rows.Scan(args...); err != nil {")
		p.In()
		p.P("return err")
		p.Out()
		p.P("}")
		p.P("res = append(res, m)")
		p.Out()
		p.P("}")
		p.P("if err := rows.Err(); err != nil {")
		p.In()
		p.P("return err")
		p.Out()
		p.P("}")
		p.P("*ms = res")
		p.P("return nil")
		p.Out()
		p.P("}")
	}
	if def.all || def.update {
		now := false
		for _, col := range def.cols {
			if col.timeLevel == timeUpdate {
				now = true
				break
			}
		}
		p.P("")
		p.P("func (m %v) SQLUpdate(s IState, b []byte, args []interface{}) ([]byte, []interface{}, error) {", pStr)
		p.In()
		p.P("var flag bool")
		if now {
			p.P("now := time.Now()")
		}
		p.P("b = append(b, `UPDATE \"%v\" SET `...)", tableName)
		for _, col := range def.cols {
			arg := g.appendUpdateArg(nil, "m", col)
			p.P(`if %v {`, g.nonZero("m", col))
			p.In()
			p.P("flag = true")
			p.P("b = append(b, `\"%v\"=`...)", col.columnName)
			p.P("b = s.AppendMarker(b, 1)")
			p.P("b = append(b, ',')")
			p.P("args = append(args, %v)", string(arg))
			p.Out()
			p.P("}")
		}
		p.P("if !flag {")
		p.In()
		p.P(`return nil, nil, core.ErrNoColumn`)
		p.Out()
		p.P("}")
		p.P("return b[:len(b)-1], args, nil")
		p.Out()
		p.P("}")
	}
	if def.all || def.update {
		p.P("")
		p.P("func (m %v) SQLUpdateAll(s IState, b []byte, args []interface{}) ([]byte, []interface{}, error) {", pStr)
		p.In()
		p.P("b = append(b, %v...)", _UpdateAll)
		p.P("b = append(b, ` = (`...)")
		p.P("b = s.AppendMarker(b, %v)", len(def.cols))
		p.P("b = append(b, ')')")
		p.P("return b, m.SQLArgs(args, false), nil")
		p.Out()
		p.P("}")
	}
	if len(def.joins) > 0 {
		p.P("")
		p.P("func (m %v) SQLSelect(b []byte) []byte {", pStr)
		p.In()
		p.P("b = append(b, %v...)", _Select)
		p.P("b = append(b, ' ')")
		p.P("return append(b, %v...)", _Join)
		p.Out()
		p.P("}")

		p.P("")
		p.P("func (m %v) SQLSelect(b []byte) []byte {", Strs)
		p.In()
		p.P("return (%v)(nil).SQLSelect(b)", pStr)
		p.Out()
		p.P("}")

		p.P("")
		p.P("func (m %v) SQLJoin(b []byte, types []sq.JOIN_TYPE) []byte {", pStr)
		p.In()
		p.P(`if len(types) == 0 {`)
		p.In()
		p.P("return append(b, %v...)", _Join)
		p.Out()
		p.P("}")
		p.P("return m.__sqlJoin(b, types)")
		p.Out()
		p.P("}")

		p.P("")
		p.P("func (m %v) SQLJoin(b []byte, types []sq.JOIN_TYPE) []byte {", Strs)
		p.In()
		p.P("return (%v)(nil).SQLJoin(b, types)", pStr)
		p.Out()
		p.P("}")

		p.P("")
		p.P("func (m %v) __sqlSelect(b []byte) []byte {", pStr)
		p.In()
		p.P("b = append(b, `SELECT `...)")
		p.P(`b = core.AppendCols(b, string(%v), %v)`, _As, _ListColsForType(def.base))
		for i, join := range def.joins {
			p.P("b = append(b, ',')")
			p.P(`b = core.AppendCols(b, string(%v[%v]), %v)`, _JoinAs, i, _ListColsForType(join.joinTyp))
		}
		p.P("return b")
		p.Out()
		p.P("}")

		p.P("")
		p.P("func (m %v) __sqlJoin(b []byte, types []sq.JOIN_TYPE) []byte {", pStr)
		p.In()
		p.P("if len(types) != %v {", len(def.joins))
		p.In()
		p.P(`panic("Expect %v %v to join")`,
			len(def.joins), Plural(len(def.joins), "type", "types"))
		p.Out()
		p.P("}")
		p.P("b = append(b, `FROM \"`...)")
		p.P("b = append(b, %v...)", _TableForType(def.base))
		p.P("b = append(b, `\" AS `...)")
		p.P("b = append(b, %v[:len(%v)-1]...)", _As, _As)
		for i, join := range def.joins {
			p.P("b = append(b, ' ')")
			p.P("b = append(b, types[%v]...)", i)
			p.P("b = append(b, ` JOIN \"`...)")
			p.P("b = append(b, %v...)", _TableForType(join.joinTyp))
			p.P("b = append(b, `\" AS `...)")
			p.P("b = append(b, %v[%v][:len(%v[%v])-1]...)", _JoinAs, i, _JoinAs, i)
			p.P("b = append(b, ` ON `...)")
			p.P("b = append(b, %v[%v]...)", _JoinConds, i)
		}
		p.P("return b")
		p.Out()
		p.P("}")

		p.P("")
		p.P("func (m %v) SQLScanArgs(args []interface{}) []interface{} {", pStr)
		p.In()
		baseTypStr := gt.TypeString(def.base)[1:]
		p.P("m.%v = new(%v)", baseTypStr, baseTypStr)
		p.P("args = m.%v.SQLScanArgs(args)", gt.TypeString(def.base)[1:])
		for _, join := range def.joins {
			joinTypStr := gt.TypeString(join.joinTyp)[1:]
			p.P("m.%v = new(%v)", joinTypStr, joinTypStr)
			p.P("args = m.%v.SQLScanArgs(args)", joinTypStr)
		}
		p.P("return args")
		p.Out()
		p.P("}")
	}
	return nil
}

var (
	reTagColumnName = regexp.MustCompile(`'[0-9A-Za-z._-]+'`)
	reTagKeyword    = regexp.MustCompile(`\b[a-z]+\b`)
	reTagSpaces     = regexp.MustCompile(`^\s*$`)
)

func parseColumnsFromType(path pathElems, root types.Type, sTyp *types.Struct) ([]*colDef, error) {
	var cols []*colDef
	for i, n := 0, sTyp.NumFields(); i < n; i++ {
		field := sTyp.Field(i)
		if !field.Exported() {
			continue
		}
		fieldPath := path.append(field)

		tag := ""
		if rawTag := reflect.StructTag(sTyp.Tag(i)); rawTag != "" {
			t, ok := rawTag.Lookup(sqlTag)
			if !ok && t != "" {
				return nil, fmt.Errorf(
					"Invalid tag at `%v`.%v",
					gt.TypeString(root), fieldPath)
			}
			tag = t
		}
		if strings.HasPrefix(tag, "-") {
			// Skip the field
			continue
		}

		columnName := ToSnake(field.Name())
		columnType := gt.TypeString(field.Type())
		inline, create, update := false, false, false
		if tag != "" {
			ntag := tag
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
						return nil, fmt.Errorf("`create` flag can only be used on time.Time or *time.Time field")
					}
				case "update", "updated":
					update = true
					if columnType != "time.Time" && columnType != "*time.Time" {
						return nil, fmt.Errorf("`update` flag can only be used on time.Time or *time.Time field")
					}
				default:
					return nil, fmt.Errorf(
						"Unregconized keyword `%v` at `%v`.%v",
						keyword, gt.TypeString(root), fieldPath)
				}
				ntag = strings.Replace(ntag, keyword, "", -1)
			}
			if !reTagSpaces.MatchString(ntag) {
				return nil, fmt.Errorf(
					"Invalid tag at `%v`.%v (Did you forget the single quote?)",
					gt.TypeString(root), fieldPath)
			}
		}
		if countFlags(inline, create, update) > 1 {
			return nil, fmt.Errorf(
				"`inline`, `create`, `update` flags can not be used together (at `%v`.%v)", gt.TypeString(root), fieldPath)
		}
		if inline {
			typ := field.Type()
			if t, ok := typ.Underlying().(*types.Pointer); ok {
				typ = t.Elem()
			}
			if t, ok := typ.Underlying().(*types.Struct); ok {
				inlineCols, err := parseColumnsFromType(fieldPath, root, t)
				if err != nil {
					return nil, err
				}
				cols = append(cols, inlineCols...)
				continue
			}
			return nil, fmt.Errorf(
				"`inline` can only be used with struct or *struct (at `%v`.%v)", gt.TypeString(root), fieldPath)
		}

		col := &colDef{
			fieldName:  field.Name(),
			fieldType:  field.Type(),
			fieldTag:   tag,
			columnName: columnName,
			columnType: columnType,
			pathElems:  fieldPath,
		}
		if create {
			col.timeLevel = timeCreate
		} else if update {
			col.timeLevel = timeUpdate
		}
		cols = append(cols, col)
	}
	return cols, nil
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

func (g *gen) queryInsert(def *typeDef, repeat int) []byte {
	b := make([]byte, 0, 1024)
	b = appends(b, `INSERT INTO "`, def.tableName, " (")
	b = g.appendListColumns(b, "", def.cols)
	b = append(b, `) VALUES`...)
	for r := 0; r < repeat; r++ {
		if r > 0 {
			b = append(b, `, `...)
		}
		b = append(b, " ("...)
		for i := range def.cols {
			if i > 0 {
				b = append(b, ',')
			}
			b = append(b, '?')
		}
		b = append(b, ')')
	}
	return b
}

func (g *gen) appendListColumns(b []byte, prefix string, cols []*colDef) []byte {
	for i, col := range cols {
		if i > 0 {
			b = append(b, `,`...)
		}
		if prefix != "" {
			b = append(b, prefix...)
			b = append(b, '.')
		}
		b = append(b, '"')
		b = append(b, col.columnName...)
		b = append(b, '"')
	}
	return b
}

func (g *gen) listColumns(prefix string, cols []*colDef) []byte {
	b := make([]byte, 0, 1024)
	return g.appendListColumns(b, prefix, cols)
}

func (g *gen) queryArgs(prefix string, cols []*colDef) []byte {
	b := make([]byte, 0, 1024)
	for i, col := range cols {
		if i > 0 {
			b = append(b, ", "...)
		}
		b = g.appendQueryArg(b, prefix, col)
	}
	return b
}

func (g *gen) listQueryArgs(prefix string, cols []*colDef) []string {
	res := make([]string, len(cols))
	for i, col := range cols {
		res[i] = string(g.queryArg(prefix, col))
	}
	return res
}

func (g *gen) queryArg(prefix string, col *colDef) []byte {
	b := make([]byte, 0, 64)
	b = g.appendQueryArg(b, prefix, col)
	return b
}

func (g *gen) listScanArgs(prefix string, cols []*colDef) []string {
	res := make([]string, len(cols))
	for i, col := range cols {
		b := make([]byte, 0, 64)
		b = g.appendScanArg(b, prefix, col)
		res[i] = string(b)
	}
	return res
}

func (g *gen) tableName(def *typeDef) string {
	typ := def.typ
	if def.base != nil {
		typ = def.base
	}
	name := gt.TypeString(typ)[1:]
	return ToSnake(name)
}

func (g *gen) tableNameOf(typ types.Type) string {
	def := g.mapType[typ.String()]
	return g.tableName(def)
}

func (g *gen) genConvertMethodsFor(typ, base types.Type) error {
	sgen := substruct.New(g.TypesMap, g.printer, nil)
	if _, err := sgen.Add(g.GetFuncName(typ), []types.Type{typ, base}); err != nil {
		return err
	}
	return sgen.Generate([]types.Type{typ})
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

func (g *gen) parseJoin(typs []types.Type) (joins []*joinDef, err error) {
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

func (g *gen) parseJoinLine(typs []types.Type) (*joinDef, error) {
	if gt.TypeString(typs[0]) != "core.JoinType" {
		return nil, fmt.Errorf("Invalid JoinType: must be one of predefined constants (got %v)", gt.TypeString(typs[0]))
	}

	base := typs[1]
	if _, ok := pointerToStruct(base); !ok {
		return nil, fmt.Errorf(
			"Invalid base type for join: must be pointer to struct (got %v)",
			gt.TypeString(base))
	}

	as := typs[2]
	if gt.TypeString(as) != "sq.AS" {
		return nil, fmt.Errorf(
			"Invalid AS: must be sq.AS (got %v)", g.TypeString(as))
	}

	cond := typs[3]
	if gt.TypeString(cond) != "string" {
		return nil, fmt.Errorf(
			"Invalid condition for join: must be string (got %v)",
			gt.TypeString(cond))
	}

	return &joinDef{
		joinTyp: base,
	}, nil
}

func pointerToStruct(typ types.Type) (*types.Struct, bool) {
	pt, ok := typ.Underlying().(*types.Pointer)
	if !ok {
		return nil, false
	}
	st, ok := pt.Elem().Underlying().(*types.Struct)
	return st, ok
}

func isPointer(typ types.Type) bool {
	_, ok := typ.Underlying().(*types.Pointer)
	return ok
}

func capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
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

func countFlags(args ...bool) int {
	c := 0
	for _, arg := range args {
		if arg {
			c++
		}
	}
	return c
}
