package sqlgen

var tplStr = `
type {{.TypeNames}} []*{{.TypeName}}

{{if .IsSimple}}
const {{._Table}} = {{.TableName | go}}
const {{._ListCols}} = {{.ColsList | go}}
const {{._Insert}} = "INSERT INTO {{.TableName | quote}} (" + {{._ListCols}} + ") VALUES"
const {{._Select}} = "SELECT " + {{._ListCols}} + " FROM {{.TableName | quote}}"
const {{._Select}}_history = "SELECT " + {{._ListCols}} + " FROM history.{{.TableName | quote}}"
const {{._UpdateAll}} = "UPDATE {{.TableName | quote}} SET (" + {{._ListCols}} + ")"
{{else}}
var {{._JoinTypes}} = []sq.JOIN_TYPE{ {{.JoinTypes | join}} }
var {{._As}} sq.AS = "{{.As}}"
var {{._JoinAs}} = []sq.AS{ {{.JoinAs | join}} }
var {{._JoinConds}} = []string{ {{.JoinConds | join}} }
{{end}}

func (m *{{.TypeName}}) SQLTableName() string { return {{.TableName | go}} }
func (m {{.TypeNames}}) SQLTableName() string { return {{.TableName | go}} }

{{if or .IsAll .IsInsert .IsUpdate}}
func (m *{{.TypeName}}) SQLArgs(opts core.Opts, create bool) []interface{} {
	{{if gt .TimeLevel 0 -}}
	now := time.Now()
	{{end -}}
	return []interface{}{
		{{range .QueryArgs -}}
		{{.}},
		{{end -}}
	}
}
{{end}}

{{if or .IsAll .IsSelect}}
func (m *{{.TypeName}}) SQLScanArgs(opts core.Opts) []interface{} {
	{{range .PtrElems -}}
	m.{{.Path}} = new({{.TypeName}})
	{{end -}}
	return []interface{}{
		{{range .ScanArgs -}}
		{{.}},
		{{end -}}
	}
}
{{end}}

{{if or .IsAll .IsSelect .IsJoin}}
func (m *{{.TypeName}}) SQLScan(opts core.Opts, row *sql.Row) error {
	return row.Scan(m.SQLScanArgs(opts)...)
}

func (ms *{{.TypeNames}}) SQLScan(opts core.Opts, rows *sql.Rows) error {
	res := make({{.TypeNames}}, 0, 128)
	for rows.Next() {
		m := new({{.TypeName}})
		args := m.SQLScanArgs(opts)
		if err := rows.Scan(args...); err != nil {
			return err
		}
		res = append(res, m)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	*ms = res
	return nil
}
{{end}}

{{if or .IsAll .IsSelect}}
func (_ *{{.TypeName}}) SQLSelect(w SQLWriter) error {
	w.WriteQueryString({{._Select}})
		return nil
	}

func (_ {{.TypeNames}}) SQLSelect(w SQLWriter) error {
	w.WriteQueryString({{._Select}})
	return nil
}
{{end}}

{{if or .IsAll .IsInsert}}
func (m *{{.TypeName}}) SQLInsert(w SQLWriter) error {
	w.WriteQueryString({{._Insert}})
	w.WriteRawString(" (")
	w.WriteMarkers({{.NumCols}})
	w.WriteByte(')')
	w.WriteArgs(m.SQLArgs(w.Opts(), true))
	return nil
}

func (ms {{.TypeNames}}) SQLInsert(w SQLWriter) error {
	w.WriteQueryString({{._Insert}})
	w.WriteRawString(" (")
	for i := 0; i < len(ms); i++ {
		w.WriteMarkers({{.NumCols}})
		w.WriteArgs(ms[i].SQLArgs(w.Opts(), true))
		w.WriteRawString("),(")
	}
	w.TrimLast(2)
	return nil
}
{{end}}

{{if or .IsAll .IsUpdate}}
func (m *{{.TypeName}}) SQLUpdate(w SQLWriter) error {
	now, opts := time.Now(), w.Opts()
	_, _ = now, opts // suppress unused error
	var flag bool
	w.WriteRawString("UPDATE ")
	w.WriteName({{.TableName | go}})
	w.WriteRawString(" SET ")
	{{range .Cols -}}
	if {{nonzero .}} {
		flag = true
		w.WriteName({{.ColumnName | go}})
		w.WriteByte('=')
		w.WriteMarker()
		w.WriteByte(',')
		w.WriteArg({{updateArg .}})
	}
	{{end -}}
	if !flag {
		return core.ErrNoColumn
	}
	w.TrimLast(1)
	return nil
}

func (m *{{.TypeName}}) SQLUpdateAll(w SQLWriter) error {
	w.WriteQueryString({{._UpdateAll}})
	w.WriteRawString(" = (")
	w.WriteMarkers({{.NumCols}})
	w.WriteByte(')')
	w.WriteArgs(m.SQLArgs(w.Opts(), false))
	return nil
}
{{end}}

{{if .IsJoin}}
func (m *{{.TypeName}}) SQLSelect(w SQLWriter) error {
	(*{{.TypeName}})(nil).__sqlSelect(w)
	w.WriteByte(' ')
	(*{{.TypeName}})(nil).__sqlJoin(w, {{._JoinTypes}})
	return nil
	}

	func (m {{.TypeNames}}) SQLSelect(w SQLWriter) error {
	return (*{{.TypeName}})(nil).SQLSelect(w)
	}

	func (m *{{.TypeName}}) SQLJoin(w SQLWriter, types []sq.JOIN_TYPE) error {
	if len(types) == 0 {
	types = {{._JoinTypes}}
	}
	m.__sqlJoin(w, types)
	return nil
	}

	func (m {{.TypeNames}}) SQLJoin(w SQLWriter, types []sq.JOIN_TYPE) error {
	return (*{{.TypeName}})(nil).SQLJoin(w, types)
	}

	func (m *{{.TypeName}}) __sqlSelect(w SQLWriter) {
	w.WriteRawString("SELECT ")
	core.WriteCols(w, string({{._As}}), {{.BaseType | listColsForType}})
	{{range $i, $join := .Joins -}}
		w.WriteByte(',')
		core.WriteCols(w, string({{$._JoinAs}}[{{$i}}]), {{$join.JoinType | listColsForType}})
	{{end -}}
	}

	func (m *{{.TypeName}}) __sqlJoin(w SQLWriter, types []sq.JOIN_TYPE) {
	if len(types) != {{.NumJoins}} {
	panic("sqlgen: expect {{plural .NumJoins "type"}} to join")
	}
	w.WriteRawString("FROM ")
	w.WriteName({{.TableName | go}})
	w.WriteRawString(" AS ")
	w.WriteRawString(string({{._As}}))
	{{range $i, $join := .Joins -}}
		w.WriteByte(' ')
		w.WriteRawString(string(types[{{$i}}]))
		w.WriteRawString(" JOIN ")
		w.WriteName({{$join.JoinType | tableForType}})
		w.WriteRawString(" AS ")
		w.WriteRawString(string({{$._JoinAs}}[{{$i}}]))
		w.WriteRawString(" ON ")
		w.WriteQueryString({{$._JoinConds}}[{{$i}}])
	{{end -}}
	}

	func (m *{{.TypeName}}) SQLScanArgs(opts core.Opts) []interface{} {
	args := make([]interface{}, 0, 64) // TODO: pre-calculate length
	m.{{.BaseType | typeName}} = new({{.BaseType | typeName}})
	args = append(args, m.{{.BaseType | typeName}}.SQLScanArgs(opts)...)
	{{range $i, $join := .Joins -}}
		m.{{$join.JoinType | typeName}} = new({{$join.JoinType | typeName}})
		args = append(args, m.{{$join.JoinType | typeName}}.SQLScanArgs(opts)...)
	{{end}}
	return args
	}
{{end}}

{{if .IsPreload}}
func (m *{{.TypeName}}) SQLPreload(table string) *core.PreloadDesc {
	switch table {
	{{range .Preloads -}}
	case {{.TableName | go}}:
		var items {{.PluralTypeStr}}
		return &core.PreloadDesc{
			Fkey: {{.Fkey | go}},
			IDs: []interface{}{m.ID},
			Items: &items,
		}
	{{end -}}
	default:
		return nil
	}
}

func (m {{.TypeNames}}) SQLPreload(table string) *core.PreloadDesc {
	switch table {
	{{range .Preloads -}}
	case {{.TableName | go}}:
		ids := make([]interface{}, len(m))
		for i, item := range m {
			ids[i] = item.ID
		}
		var items {{.PluralTypeStr}}
		return &core.PreloadDesc{
			Fkey: {{.Fkey | go}},
			IDs: ids,
			Items: &items,
		}
	{{end -}}
	default:
		return nil
	}
}

func (m *{{.TypeName}}) SQLPopulate(items core.IFind) error {
	switch items := items.(type) {
	{{range .Preloads -}}
		case *{{.PluralTypeStr}}:
		m.{{.FieldName}} = *items
		return nil
	{{end -}}
	default:
		return core.Errorf("can not populate %%T into %%T", items, m)
	}
}

func (m {{.TypeNames}}) SQLPopulate(items core.IFind) error {
	mapID := make(map[int64]*{{.TypeName}})
	for _, item := range m {
		mapID[item.ID] = item
	}

	switch items := items.(type) {
	{{range .Preloads -}}
	case *{{.PluralTypeStr}}:
		for _, item := range *items {
			mitem := mapID[item.{{.Fkey | toTitle}}]
			if mitem == nil {
				return core.Errorf("can not populate id %%v", item.{{.Fkey | toTitle}})
			}
			mitem.{{.FieldName}} = append(mitem.{{.FieldName}}, item)
		}
		return nil
	{{end -}}
	default:
		return core.Errorf("can not populate %%T into %%T", items, m)
	}
}
{{end}}
`
