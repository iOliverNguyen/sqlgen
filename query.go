package sq

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"

	core "github.com/ng-vu/sqlgen/core"
)

type Query struct {
	db     dbInterface
	ctx    context.Context
	marker core.Marker

	table     string
	updateAll bool
	limit     string
	offset    string

	selects    []string
	prefixes   exprs
	sqls       exprs
	whereParts parts
	orderBys   []string
	groupBys   []string
	suffixes   exprs
}

func (db *Database) NewQuery() *Query {
	return &Query{db: db, marker: db.marker, ctx: context.Background()}
}

func (tx *tx) NewQuery() *Query {
	return &Query{db: tx, marker: tx.db.marker, ctx: tx.ctx}
}

func (q *Query) WithContext(ctx context.Context) *Query {
	q.ctx = ctx
	return q
}

func (q *Query) Clone() *Query {
	nq := &Query{
		db:     q.db,
		ctx:    q.ctx,
		marker: q.marker,

		table:     q.table,
		updateAll: q.updateAll,
		limit:     q.limit,
		offset:    q.offset,

		prefixes:   make(exprs, len(q.prefixes)),
		sqls:       make(exprs, len(q.sqls)),
		whereParts: make([]Sqlizer, len(q.whereParts)),
		orderBys:   make([]string, len(q.orderBys)),
		groupBys:   make([]string, len(q.groupBys)),
		suffixes:   make(exprs, len(q.suffixes)),
	}
	copy(nq.prefixes, q.prefixes)
	copy(nq.sqls, q.sqls)
	copy(nq.whereParts, q.whereParts)
	copy(nq.orderBys, q.orderBys)
	copy(nq.groupBys, q.groupBys)
	copy(nq.suffixes, q.suffixes)
	return nq
}

var bpool = &sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, 1024)
		return &b
	},
}

func poolGet() *[]byte {
	return bpool.Get().(*[]byte)
}

func poolPut(b *[]byte) {
	*b = (*b)[:0]
	bpool.Put(b)
}

type questionMarker struct{}

func (q questionMarker) AppendMarker(b []byte, n int) []byte {
	b = append(b, '?')
	for i := 1; i < n; i++ {
		b = append(b, ",?"...)
	}
	return b
}

func (q questionMarker) AppendQuery(b []byte, query []byte) []byte {
	return append(b, query...)
}

func (q questionMarker) AppendQueryStr(b []byte, query string) []byte {
	return append(b, query...)
}

type dollarMarker struct{ c int64 }

func (d *dollarMarker) AppendMarker(b []byte, n int) []byte {
	d.c++
	b = append(b, '$')
	b = strconv.AppendInt(b, d.c, 10)
	for i := 1; i < n; i++ {
		d.c++
		b = append(b, ",$"...)
		b = strconv.AppendInt(b, d.c, 10)
	}
	return b
}

func (d *dollarMarker) AppendQuery(b []byte, query []byte) []byte {
	for i := range query {
		ch := query[i]
		switch ch {
		case '?':
			d.c++
			b = append(b, '$')
			b = strconv.AppendInt(b, d.c, 10)
		default:
			b = append(b, ch)
		}
	}
	return b
}

func (d *dollarMarker) AppendQueryStr(b []byte, query string) []byte {
	for i := range query {
		ch := query[i]
		switch ch {
		case '?':
			d.c++
			b = append(b, '$')
			b = strconv.AppendInt(b, d.c, 10)
		default:
			b = append(b, ch)
		}
	}
	return b
}

type builder func(core.IState, []byte, []interface{}) ([]byte, []interface{}, error)

func builderSimple(fn func(b []byte) []byte) builder {
	return func(_ core.IState, b []byte, args []interface{}) ([]byte, []interface{}, error) {
		return fn(b), args, nil
	}
}

func builderQuery(query string) builder {
	return func(_ core.IState, b []byte, args []interface{}) ([]byte, []interface{}, error) {
		return append(b, query...), args, nil
	}
}

func (q *Query) build(typ string, def interface{}, fn builder) (_ string, _ []interface{}, err error) {
	buf := poolGet()
	b := *buf
	args := make([]interface{}, 0, 64)
	defer func() {
		poolPut(buf)
		if err != nil {
			entry := &LogEntry{
				Ctx:   q.ctx,
				Query: string(b),
				Args:  args,
				Error: err,
				Flags: FlagBuild,
			}
			err = q.db.log(entry)
		}
	}()

	if (typ == "UPDATE" || typ == "DELETE") && len(q.whereParts) == 0 {
		return "", nil, core.Errorf("common/sql: %v must have WHERE", typ)
	}

	s := q.marker()
	if len(q.prefixes) > 0 {
		b, args, err = q.prefixes.Append(s, b, args, " ")
		if err != nil {
			return "", nil, err
		}
	}

	if len(b) > 0 {
		b = append(b, ' ')
	}

	switch {
	case len(q.selects) > 0:
		b = append(b, `SELECT `...)
		for i, sel := range q.selects {
			if i > 0 {
				b = append(b, ',')
			}
			quote := canbeName(sel)
			if quote {
				b = append(b, '"')
			}
			b = append(b, sel...)
			if quote {
				b = append(b, '"')
			}
		}
		if def, ok := def.(core.IJoin); ok {
			b = append(b, ' ')
			b = def.SQLJoin(b, nil)
		} else if q.table != "" {
			b = append(b, ` FROM "`...)
			b = append(b, q.table...)
			b = append(b, '"')
		}

	case fn != nil:
		b, args, err = fn(s, b, args)
		if err != nil {
			return "", nil, err
		}
	default:
		return "", nil, core.ErrNoAction
	}

	if len(q.sqls) > 0 {
		if len(b) > 0 {
			b = append(b, ' ')
		}
		b, args, err = q.sqls.Append(s, b, args, " ")
		if err != nil {
			return "", nil, err
		}
	}
	if len(q.whereParts) > 0 {
		b = append(b, " WHERE ("...)
		b, args, err = q.whereParts.Append(s, b, args, ") AND (")
		if err != nil {
			return "", nil, err
		}
		b = append(b, ')')
	}
	if len(q.groupBys) > 0 {
		b = append(b, " GROUP BY "...)
		for i, s := range q.groupBys {
			if i > 0 {
				b = append(b, ',')
			}
			quote := canbeName(s)
			if quote {
				b = append(b, '"')
			}
			b = append(b, s...)
			if quote {
				b = append(b, '"')
			}
		}
	}
	if len(q.orderBys) > 0 {
		b = append(b, " ORDER BY "...)
		for i, s := range q.orderBys {
			if i > 0 {
				b = append(b, ',')
			}
			quote := canbeName(s)
			if quote {
				b = append(b, '"')
			}
			b = append(b, s...)
			if quote {
				b = append(b, '"')
			}
		}
	}
	if q.limit != "" {
		b = append(b, " LIMIT "...)
		b = append(b, q.limit...)
	}
	if q.offset != "" {
		b = append(b, " OFFSET "...)
		b = append(b, q.offset...)
	}
	if len(q.suffixes) > 0 {
		b, args, err = q.suffixes.Append(s, b, args, " ")
		if err != nil {
			return "", nil, err
		}
		b = append(b, ' ')
	}
	return string(b), args, nil
}

func canbeName(s string) bool {
	if s == "" {
		panic("common/sql: Empty!")
	}
	c := s[0]
	if !(c == '_' ||
		c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z') {
		return false
	}
	for i := range s {
		c := s[i]
		if !(c == '_' ||
			c >= 'a' && c <= 'z' ||
			c >= 'A' && c <= 'Z' ||
			c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

func (q *Query) assertTable(obj core.ITableName) {
	table := obj.SQLTableName()
	if q.table == "" && table == "" {
		panic("common/sql: No table provided")
	}
	if q.table != "" && table != "" && q.table != table {
		panic(fmt.Sprintf(
			"common/sql: Table name does not match: %v != %v",
			q.table, obj.SQLTableName()))
	}
}

func (q *Query) Build() (string, []interface{}, error) {
	return q.build("", nil, builderQuery(""))
}

func (q *Query) BuildGet(obj core.IGet) (string, []interface{}, error) {
	q.assertTable(obj)
	return q.build("SELECT", obj, builderSimple(obj.SQLSelect))
}

func (q *Query) BuildFind(objs core.IFind) (string, []interface{}, error) {
	q.assertTable(objs)
	return q.build("SELECT", objs, builderSimple(objs.SQLSelect))
}

func (q *Query) BuildInsert(obj core.IInsert) (string, []interface{}, error) {
	q.assertTable(obj)
	return q.build("INSERT", nil, obj.SQLInsert)
}

func (q *Query) BuildUpdate(obj core.IUpdate) (string, []interface{}, error) {
	q.assertTable(obj)
	fn := obj.SQLUpdate
	if q.updateAll {
		fn = obj.SQLUpdateAll
	}
	return q.build("UPDATE", nil, fn)
}

func (q *Query) BuildDelete(obj core.ITableName) (string, []interface{}, error) {
	q.assertTable(obj)
	tableName := obj.SQLTableName()
	query := `DELETE FROM "` + tableName + `"`
	return q.build("DELETE", nil, builderQuery(query))
}

func (q *Query) BuildCount(obj core.ITableName) (string, []interface{}, error) {
	q.assertTable(obj)
	q = q.Select(`COUNT(*)`).Table(obj.SQLTableName())
	return q.build("SELECT", obj, nil)
}

func (q *Query) Exec() (sql.Result, error) {
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return q.db.ExecContext(q.ctx, query, args...)
}

func (q *Query) Query() (*sql.Rows, error) {
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return q.db.QueryContext(q.ctx, query, args...)
}

func (q *Query) QueryRow() (Row, error) {
	query, args, err := q.Build()
	if err != nil {
		return Row{}, err
	}
	return q.db.QueryRowContext(q.ctx, query, args...), nil
}

func (q *Query) Scan(dest ...interface{}) error {
	query, args, err := q.Build()
	if err != nil {
		return err
	}
	return q.db.QueryRowContext(q.ctx, query, args...).Scan(dest...)
}

func (q *Query) Get(obj core.IGet) (bool, error) {
	q.limit = "1"
	query, args, err := q.BuildGet(obj)
	if err != nil {
		return false, err
	}
	row := q.db.QueryRowContext(q.ctx, query, args...)
	sqlErr := obj.SQLScan(row.row)

	// The above SQLScan() is called with *sql.Row, therefore logging and
	// mapping come here.
	err = row.log(sqlErr)
	if sqlErr == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (q *Query) Find(objs core.IFind) error {
	query, args, err := q.BuildFind(objs)
	if err != nil {
		return err
	}
	rows, err := q.db.QueryContext(q.ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	return objs.SQLScan(rows)
}

func (q *Query) Insert(objs ...core.IInsert) (int64, error) {
	switch len(objs) {
	case 0:
		return 0, nil
	case 1:
		query, args, err := q.BuildInsert(objs[0])
		if err != nil {
			return 0, err
		}
		res, err := q.db.ExecContext(q.ctx, query, args...)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	default:
		doInsert := func(tx Tx) (int64, error) {
			var count int64
			for _, obj := range objs {
				c, err := tx.Insert(obj)
				if err != nil {
					return 0, err
				}
				count += c
			}
			return count, nil
		}

		switch x := q.db.(type) {
		case Tx:
			return doInsert(x)
		case *Database:
			tx, err := x.Begin()
			if err != nil {
				return 0, err
			}
			defer func() { _ = tx.Rollback() }()
			n, err := doInsert(tx)
			if err != nil {
				return 0, err
			}
			return n, tx.Commit()
		default:
			panic("Expect Database or Tx")
		}
	}
}

func (q *Query) Update(objs ...core.IUpdate) (int64, error) {
	switch len(objs) {
	case 1:
		query, args, err := q.BuildUpdate(objs[0])
		if err != nil {
			return 0, err
		}
		res, err := q.db.ExecContext(q.ctx, query, args...)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	case 0:
		return 0, nil
	}
	return 0, errors.New("TODO")
}

func (q *Query) UpdateMap(m map[string]interface{}) (int64, error) {
	return q.Update(core.Map{Table: q.table, M: m})
}

func (q *Query) Delete(obj core.ITableName) (int64, error) {
	query, args, err := q.BuildDelete(obj)
	if err != nil {
		return 0, err
	}
	res, err := q.db.ExecContext(q.ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (q *Query) Count(obj core.ITableName) (n uint64, err error) {
	query, args, err := q.BuildCount(obj)
	if err != nil {
		return 0, err
	}
	err = q.db.QueryRowContext(q.ctx, query, args...).Scan(&n)
	return
}

func (q *Query) Table(name string) *Query {
	q.table = name
	return q
}

// Prefix adds an expression to the start of the query
func (q *Query) Prefix(sql string, args ...interface{}) *Query {
	q.prefixes = append(q.suffixes, Expr(sql, args...))
	return q
}

func (q *Query) Select(cols ...string) *Query {
	q.selects = append(q.selects, cols...)
	return q
}

func (q *Query) From(name string) *Query {
	q.table = name
	return q
}

func (q *Query) SQL(sql string, args ...interface{}) *Query {
	q.sqls = append(q.sqls, expr{sql, args})
	return q
}

// Where adds WHERE expressions to the query.
func (q *Query) Where(cond string, args ...interface{}) *Query {
	q.whereParts = append(q.whereParts, newWherePart(cond, args...))
	return q
}

// OrderBy adds ORDER BY expressions to the query.
func (q *Query) OrderBy(orderBys ...string) *Query {
	q.orderBys = append(q.orderBys, orderBys...)
	return q
}

// GroupBy adds GROUP BY expressions to the query.
func (q *Query) GroupBy(groupBys ...string) *Query {
	q.groupBys = append(q.groupBys, groupBys...)
	return q
}

// Limit sets a LIMIT clause on the query.
func (q *Query) Limit(limit uint64) *Query {
	q.limit = strconv.FormatUint(limit, 10)
	return q
}

// Offset sets a OFFSET clause on the query.
func (q *Query) Offset(offset uint64) *Query {
	q.offset = strconv.FormatUint(offset, 10)
	return q
}

// Suffix adds an expression to the end of the query
func (q *Query) Suffix(sql string, args ...interface{}) *Query {
	q.suffixes = append(q.suffixes, Expr(sql, args...))
	return q
}

func (q *Query) UpdateAll() *Query {
	q.updateAll = true
	return q
}

func (q *Query) In(column string, args ...interface{}) *Query {
	switch len(args) {
	case 0:
		q.whereParts = append(q.whereParts, newWherePart("FALSE"))
		return q
	case 1:
		if t := reflect.TypeOf(args[0]); t.Kind() == reflect.Slice {
			vArgs := reflect.ValueOf(args[0])
			if vArgs.Len() == 0 {
				q.whereParts = append(q.whereParts, newWherePart("FALSE"))
				return q
			}
			args = make([]interface{}, vArgs.Len())
			for i, n := 0, vArgs.Len(); i < n; i++ {
				args[i] = vArgs.Index(i).Interface()
			}
		}
	}

	b := make([]byte, 0, 64)
	quote := canbeName(column)
	if quote {
		b = append(b, '"')
	}
	b = append(b, column...)
	if quote {
		b = append(b, '"')
	}
	b = append(b, " IN ("...)
	for i := range args {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '?')
	}
	b = append(b, ')')
	q.whereParts = append(q.whereParts, newWherePart(b, args...))
	return q
}
