package sq

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/ng-vu/sqlgen/core"
)

// Query ...
type queryImpl struct {
	db  dbInterface
	ctx context.Context

	opts      core.Opts
	quote     byte
	marker    byte
	updateAll bool
	withTable bool

	table  string
	limit  string
	offset string

	errors     []error
	selects    []string
	prefixes   Parts
	sqls       Parts
	whereParts Parts
	orderBys   []string
	groupBys   []string
	suffixes   Parts

	preloads []preloadPart
}

var _ Query = &queryImpl{}

// NewQuery ...
func (db *Database) NewQuery() Query {
	return &queryImpl{
		db:  db,
		ctx: context.Background(),

		opts:   db.opts,
		quote:  db.quote,
		marker: db.marker,
	}
}

// NewQuery ...
func (tx *tx) NewQuery() Query {
	return &queryImpl{
		db:  tx,
		ctx: tx.ctx,

		opts:   tx.db.opts,
		quote:  tx.db.quote,
		marker: tx.db.marker,
	}
}

func (q *queryImpl) NewQuery() Query {
	return &queryImpl{
		db:     q.db,
		ctx:    q.ctx,
		opts:   q.opts,
		quote:  q.quote,
		marker: q.marker,
	}
}

// WithContext ...
func (q *queryImpl) WithContext(ctx context.Context) Query {
	q.ctx = ctx
	return q
}

// Clone ...
func (q *queryImpl) Clone() Query {
	return q.cloneWithPreds(nil)
}

func (q *queryImpl) withPreds(preds []interface{}) *queryImpl {
	if len(preds) == 0 {
		return q
	}
	return q.cloneWithPreds(preds)
}

func (q *queryImpl) cloneWithPreds(preds []interface{}) *queryImpl {
	nq := &queryImpl{
		db:         q.db,
		ctx:        q.ctx,
		opts:       q.opts,
		quote:      q.quote,
		marker:     q.marker,
		updateAll:  q.updateAll,
		withTable:  q.withTable,
		table:      q.table,
		limit:      q.limit,
		offset:     q.offset,
		errors:     make([]error, len(q.errors)),
		selects:    make([]string, len(q.selects), len(q.selects)+len(preds)),
		prefixes:   make(Parts, len(q.prefixes)),
		sqls:       make(Parts, len(q.sqls)),
		whereParts: make([]WriterTo, len(q.whereParts)),
		orderBys:   make([]string, len(q.orderBys)),
		groupBys:   make([]string, len(q.groupBys)),
		suffixes:   make(Parts, len(q.suffixes)),
	}
	copy(nq.errors, q.errors)
	copy(nq.selects, q.selects)
	copy(nq.prefixes, q.prefixes)
	copy(nq.sqls, q.sqls)
	copy(nq.whereParts, q.whereParts)
	copy(nq.orderBys, q.orderBys)
	copy(nq.groupBys, q.groupBys)
	copy(nq.suffixes, q.suffixes)
	_ = nq.Where(preds...)
	return nq
}

type builderFunc func(core.SQLWriter) error

func (q *queryImpl) build(typ string, def interface{}, fn builderFunc) (_ string, _ []interface{}, err error) {
	w := NewWriter(q.opts, q.quote, q.marker, 512)
	defer func() {
		if err != nil {
			entry := &LogEntry{
				Ctx:   q.ctx,
				Query: w.String(),
				Args:  w.args,
				Error: err,
				Flags: FlagBuild,
			}
			err = q.db.log(entry)
		}
	}()

	if (typ == "UPDATE" || typ == "DELETE") && len(q.whereParts) == 0 {
		return "", nil, core.Errorf("sqlgen: %v must have WHERE", typ)
	}

	if len(q.prefixes) > 0 {
		err = q.prefixes.WriteSQLTo(w, " ")
		if err != nil {
			return
		}
		w.WriteByte(' ')
	}

	switch {
	case len(q.selects) > 0:
		w.WriteRawString(`SELECT `)
		for i, sel := range q.selects {
			if i != 0 {
				w.WriteByte(',')
			}
			w.WriteQueryName(sel)
		}
		if def, ok := def.(core.IJoin); ok {
			w.WriteByte(' ')
			def.SQLJoin(w, nil)
		} else if q.table != "" {
			w.WriteRawString(` FROM `)
			w.WriteQueryName(q.table)
		}
		w.WriteByte(' ')

	case fn != nil:
		err = fn(w)
		if err != nil {
			return
		}
		if q.withTable && q.table != "" {
			w.WriteRawString(` FROM `)
			w.WriteQueryName(q.table)
		}
		w.WriteByte(' ')
	}

	if len(q.sqls) != 0 {
		err = q.sqls.WriteSQLTo(w, " ")
		if err != nil {
			return
		}
		w.WriteByte(' ')
	}
	if len(q.whereParts) != 0 {
		w.WriteRawString("WHERE (")
		err = q.whereParts.WriteSQLTo(w, ") AND (")
		if err != nil {
			return
		}
		w.WriteRawString(") ")
	}
	if len(q.groupBys) != 0 {
		w.WriteRawString("GROUP BY ")
		for i, s := range q.groupBys {
			if i != 0 {
				w.WriteByte(',')
			}
			w.WriteQueryName(s)
		}
		w.WriteByte(' ')
	}
	if len(q.orderBys) != 0 {
		w.WriteRawString("ORDER BY ")
		for i, s := range q.orderBys {
			if i != 0 {
				w.WriteByte(',')
			}
			w.WriteQueryName(s)
		}
		w.WriteByte(' ')
	}
	if q.limit != "" {
		w.WriteRawString("LIMIT ")
		w.WriteRawString(q.limit)
		w.WriteByte(' ')
	}
	if q.offset != "" {
		w.WriteRawString("OFFSET ")
		w.WriteRawString(q.offset)
		w.WriteByte(' ')
	}
	if len(q.suffixes) != 0 {
		err = q.suffixes.WriteSQLTo(w, " ")
		if err != nil {
			return
		}
		w.WriteByte(' ')
	}
	s := w.String()
	return s[:len(s)-1], w.args, nil
}

func (q *queryImpl) assertTable(obj core.ITableName) {
	table := obj.SQLTableName()
	if table == "" {
		if q.table == "" && len(q.sqls) == 0 {
			panic("sqlgen: no table name provided")
		}
		q.withTable = true
	}
	if q.table != "" && table != "" && q.table != table {
		msg := fmt.Sprintf(
			"sqlgen: table name does not match: %v != %v",
			q.table, obj.SQLTableName())
		panic(msg)
	}
}

func (q *queryImpl) BuildPreload(table string, obj interface{}, preds ...interface{}) (string, []interface{}, error) {
	preloader, ok := obj.(core.IPreload)
	if !ok {
		return "", nil, core.Errorf("sqlgen: %T does not support preload", obj)
	}
	desc := preloader.SQLPreload(table)
	if desc == nil {
		return "", nil, core.Errorf("sqlgen: %T does not support preload table %v", obj, table)
	}

	fkey, ids, items := desc.Fkey, desc.IDs, desc.Items
	if ids == nil || fkey == "" || items == nil {
		return "", nil, core.Errorf("sqlgen: invalid preload description")
	}

	nq := q.NewQuery().In(fkey, ids).Where(preds...)
	return nq.BuildFind(items)
}

// Build ...
func (q *queryImpl) Build(preds ...interface{}) (string, []interface{}, error) {
	return q.withPreds(preds).build("", nil, nil)
}

func (q *queryImpl) doPreloads(obj interface{}) error {
	if len(q.preloads) == 0 {
		return nil
	}
	exprs := make([]ExprString, len(q.preloads))
	for i, preload := range q.preloads {
		query, args, err := q.BuildPreload(preload.table, obj, preload.preds...)
		if err != nil {
			return err
		}
		exprs[i] = ExprString{query, args}
	}

	for _, expr := range exprs {
		query, args := expr.SQL, expr.Args
		_, err := q.db.ExecContext(q.ctx, query, args)
		if err != nil {
			return err
		}
	}
	return nil
}

// BuildGet ...
func (q *queryImpl) BuildGet(obj core.IGet, preds ...interface{}) (string, []interface{}, error) {
	q.assertTable(obj)
	return q.withPreds(preds).build("SELECT", obj, obj.SQLSelect)
}

// BuildFind ...
func (q *queryImpl) BuildFind(objs core.IFind, preds ...interface{}) (string, []interface{}, error) {
	q.assertTable(objs)
	return q.withPreds(preds).build("SELECT", objs, objs.SQLSelect)
}

// BuildInsert ...
func (q *queryImpl) BuildInsert(obj core.IInsert) (string, []interface{}, error) {
	q.assertTable(obj)
	return q.build("INSERT", nil, obj.SQLInsert)
}

// BuildUpdate ...
func (q *queryImpl) BuildUpdate(obj core.IUpdate) (string, []interface{}, error) {
	q.assertTable(obj)
	fn := obj.SQLUpdate
	if q.updateAll {
		fn = obj.SQLUpdateAll
	}
	return q.build("UPDATE", nil, fn)
}

// BuildDelete ...
func (q *queryImpl) BuildDelete(obj core.ITableName) (string, []interface{}, error) {
	q.assertTable(obj)
	tableName := obj.SQLTableName()
	return q.build("DELETE", nil, func(w core.SQLWriter) error {
		w.WriteRawString("DELETE FROM ")
		w.WriteName(tableName)
		return nil
	})
}

// BuildCount ...
func (q *queryImpl) BuildCount(obj core.ITableName, preds ...interface{}) (string, []interface{}, error) {
	q.assertTable(obj)
	q.Select(`COUNT(*)`).Table(obj.SQLTableName())
	return q.withPreds(preds).build("SELECT", obj, nil)
}

// Exec ...
func (q *queryImpl) Exec() (sql.Result, error) {
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return q.db.ExecContext(q.ctx, query, args...)
}

// Query ...
func (q *queryImpl) Query() (*sql.Rows, error) {
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return q.db.QueryContext(q.ctx, query, args...)
}

// QueryRow ...
func (q *queryImpl) QueryRow() (Row, error) {
	query, args, err := q.Build()
	if err != nil {
		return Row{}, err
	}
	return q.db.QueryRowContext(q.ctx, query, args...), nil
}

// Scan ...
func (q *queryImpl) Scan(dest ...interface{}) error {
	query, args, err := q.Build()
	if err != nil {
		return err
	}
	return q.db.QueryRowContext(q.ctx, query, args...).Scan(dest...)
}

// Get ...
func (q *queryImpl) Get(obj core.IGet, preds ...interface{}) (bool, error) {
	q.limit = "1"
	query, args, err := q.BuildGet(obj, preds...)
	if err != nil {
		return false, err
	}
	row := q.db.QueryRowContext(q.ctx, query, args...)
	sqlErr := obj.SQLScan(q.opts, row.Row)

	// The above SQLScan() is called with *sql.Row, therefore logging and
	// mapping come here.
	err = row.Log(sqlErr)
	if sqlErr == sql.ErrNoRows {
		return false, nil
	}
	if err == nil && len(q.preloads) > 0 {
		err = q.doPreloads(obj)
	}
	return err == nil, err
}

// Find ...
func (q *queryImpl) Find(objs core.IFind, preds ...interface{}) error {
	query, args, err := q.BuildFind(objs, preds...)
	if err != nil {
		return err
	}
	rows, err := q.db.QueryContext(q.ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	err = objs.SQLScan(q.opts, rows)
	if err == nil && len(q.preloads) > 0 {
		err = q.doPreloads(objs)
	}
	return err
}

// Insert ...
func (q *queryImpl) Insert(objs ...core.IInsert) (int64, error) {
	switch len(objs) {
	case 0:
		return 0, nil
	case 1:
		if err := execBeforeInsert(objs[0]); err != nil {
			return 0, err
		}
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
			for _, obj := range objs {
				if err := execBeforeInsert(obj); err != nil {
					return 0, err
				}
			}
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

// Update ...
func (q *queryImpl) Update(objs ...core.IUpdate) (int64, error) {
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

// UpdateMap ...
func (q *queryImpl) UpdateMap(m map[string]interface{}) (int64, error) {
	return q.Update(core.Map{Table: q.table, M: m})
}

// Delete ...
func (q *queryImpl) Delete(obj core.ITableName) (int64, error) {
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

// Count ...
func (q *queryImpl) Count(obj core.ITableName, preds ...interface{}) (n uint64, err error) {
	query, args, err := q.withPreds(preds).BuildCount(obj)
	if err != nil {
		return 0, err
	}
	err = q.db.QueryRowContext(q.ctx, query, args...).Scan(&n)
	return
}

// Table ...
func (q *queryImpl) Table(name string) Query {
	q.table = name
	return q
}

// Prefix adds an expression to the start of the query
func (q *queryImpl) Prefix(sql string, args ...interface{}) Query {
	q.prefixes = append(q.suffixes, ExprString{sql, args})
	return q
}

// Select ...
func (q *queryImpl) Select(cols ...string) Query {
	q.selects = append(q.selects, cols...)
	return q
}

// From ...
func (q *queryImpl) From(name string) Query {
	q.table = name
	return q
}

// SQL ...
func (q *queryImpl) SQL(preds ...interface{}) Query {
	q.sqls = append(q.sqls, NewExpr(preds...))
	return q
}

// Where adds WHERE expressions to the query.
func (q *queryImpl) Where(preds ...interface{}) Query {
	q.whereParts = q.whereParts.Append(preds...)
	return q
}

// OrderBy adds ORDER BY expressions to the query.
func (q *queryImpl) OrderBy(orderBys ...string) Query {
	q.orderBys = append(q.orderBys, orderBys...)
	return q
}

// GroupBy adds GROUP BY expressions to the query.
func (q *queryImpl) GroupBy(groupBys ...string) Query {
	q.groupBys = append(q.groupBys, groupBys...)
	return q
}

// Limit sets a LIMIT clause on the query.
func (q *queryImpl) Limit(limit uint64) Query {
	q.limit = strconv.FormatUint(limit, 10)
	return q
}

// Offset sets a OFFSET clause on the query.
func (q *queryImpl) Offset(offset uint64) Query {
	q.offset = strconv.FormatUint(offset, 10)
	return q
}

// Suffix adds an expression to the end of the query
func (q *queryImpl) Suffix(sql string, args ...interface{}) Query {
	q.suffixes = append(q.suffixes, ExprString{sql, args})
	return q
}

// UpdateAll ...
func (q *queryImpl) UpdateAll() Query {
	q.updateAll = true
	return q
}

func (q *queryImpl) In(column string, args ...interface{}) Query {
	q.whereParts = append(q.whereParts, NewInPart(true, column, args...))
	return q
}

func (q *queryImpl) NotIn(column string, args ...interface{}) Query {
	q.whereParts = append(q.whereParts, NewInPart(false, column, args...))
	return q
}

// Exists ...
func (q *queryImpl) Exists(column string, exists bool) Query {
	return q.IsNull(column, !exists)
}

// IsNull ...
func (q *queryImpl) IsNull(column string, null bool) Query {
	q.whereParts = append(q.whereParts, NewIsNullPart(column, null))
	return q
}

func (q *queryImpl) Apply(funcs ...func(query CommonQuery)) Query {
	for _, fn := range funcs {
		fn(q)
	}
	return q
}

func (q *queryImpl) Preload(table string, preds ...interface{}) Query {
	part := preloadPart{table, preds}
	q.preloads = append(q.preloads, part)
	return q
}

func execBeforeInsert(obj interface{}) error {
	if in, ok := obj.(BeforeInsertInterface); ok {
		return in.BeforeInsert()
	}
	return nil
}

func (q *queryImpl) AddError(err error) {
	q.errors = append(q.errors, err)
}
