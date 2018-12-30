package sq

import (
	"context"
	"database/sql"
	"time"

	"github.com/ng-vu/sqlgen/core"
)

type SQLWriter = core.SQLWriter

type WriterTo interface {
	WriteSQLTo(w core.SQLWriter) error
}

type WriterToFunc func(w core.SQLWriter) error

func (fn WriterToFunc) WriteSQLTo(w core.SQLWriter) error {
	return fn(w)
}

type Row = core.Row
type CommonQuery = core.CommonQuery
type Query = core.Query
type DBInterface = core.DBInterface

type dbInterface interface {
	DBInterface
	log(*LogEntry) error
}

type BeforeInsertInterface interface {
	BeforeInsert() error
}

type BeforeUpdateInterface interface {
	BeforeUpdate() error
}

// Get ...
func (db *Database) Get(obj core.IGet, preds ...interface{}) (bool, error) {
	return db.NewQuery().Get(obj, preds...)
}

// Find ...
func (db *Database) Find(objs core.IFind, preds ...interface{}) error {
	return db.NewQuery().Find(objs, preds...)
}

// Insert ...
func (db *Database) Insert(objs ...core.IInsert) (int64, error) {
	return db.NewQuery().Insert(objs...)
}

// Update ...
func (db *Database) Update(objs ...core.IUpdate) (int64, error) {
	return db.NewQuery().Update(objs...)
}

// UpdateMap ...
func (db *Database) UpdateMap(m map[string]interface{}) (int64, error) {
	return db.NewQuery().UpdateMap(m)
}

// Delete ...
func (db *Database) Delete(obj core.ITableName) (int64, error) {
	return db.NewQuery().Delete(obj)
}

// Count ...
func (db *Database) Count(obj core.ITableName, preds ...interface{}) (uint64, error) {
	return db.NewQuery().Count(obj, preds...)
}

// Table ...
func (db *Database) Table(sql string) Query {
	return db.NewQuery().Table(sql)
}

// Prefix adds an expression to the start of the query
func (db *Database) Prefix(sql string, args ...interface{}) Query {
	return db.NewQuery().Prefix(sql, args...)
}

// Select ...
func (db *Database) Select(cols ...string) Query {
	return db.NewQuery().Select(cols...)
}

// From ...
func (db *Database) From(table string) Query {
	return db.NewQuery().From(table)
}

// SQL ...
func (db *Database) SQL(args ...interface{}) Query {
	return db.NewQuery().SQL(args...)
}

// Where ...
func (db *Database) Where(args ...interface{}) Query {
	return db.NewQuery().Where(args...)
}

// OrderBy adds ORDER BY expressions to the query.
func (db *Database) OrderBy(orderBys ...string) Query {
	return db.NewQuery().OrderBy(orderBys...)
}

// GroupBy adds GROUP BY expressions to the query.
func (db *Database) GroupBy(groupBys ...string) Query {
	return db.NewQuery().GroupBy(groupBys...)
}

// Limit sets a LIMIT clause on the query.
func (db *Database) Limit(limit uint64) Query {
	return db.NewQuery().Limit(limit)
}

// Offset sets a OFFSET clause on the query.
func (db *Database) Offset(offset uint64) Query {
	return db.NewQuery().Offset(offset)
}

// Suffix adds an expression to the end of the query
func (db *Database) Suffix(sql string, args ...interface{}) Query {
	return db.NewQuery().Suffix(sql, args...)
}

// UpdateAll ...
func (db *Database) UpdateAll() Query {
	return db.NewQuery().UpdateAll()
}

// In ...
func (db *Database) In(column string, args ...interface{}) Query {
	return db.NewQuery().In(column, args...)
}

// NotIn ...
func (db *Database) NotIn(column string, args ...interface{}) Query {
	return db.NewQuery().NotIn(column, args...)
}

// Exists ...
func (db *Database) Exists(column string, exists bool) Query {
	return db.NewQuery().Exists(column, exists)
}

// IsNull ...
func (db *Database) IsNull(column string, null bool) Query {
	return db.NewQuery().IsNull(column, null)
}

func (db *Database) Preload(table string, preds ...interface{}) Query {
	return db.NewQuery().Preload(table, preds...)
}

func (db *Database) Apply(funcs ...func(CommonQuery)) Query {
	return db.NewQuery().Apply(funcs...)
}

// Tx ...
type Tx interface {
	Commit() error
	Rollback() error

	DBInterface
	CommonQuery
}

type tx struct {
	tx  *sql.Tx
	db  *Database
	t0  time.Time
	qs  []*LogEntry
	ctx context.Context
}

func (tx *tx) log(e *LogEntry) error {
	e.Flags = e.Flags | FlagTx
	return tx.db.log(e)
}

// Commit ...
func (tx *tx) Commit() (err error) {
	defer func() {
		// Only log once per tx
		if err == sql.ErrTxDone {
			return
		}
		entry := &LogEntry{
			Ctx:       tx.ctx,
			Error:     err,
			Time:      tx.t0,
			Flags:     Flags(TypeCommit) | FlagTx,
			TxQueries: tx.qs,
		}
		err = tx.db.log(entry)
	}()
	return tx.tx.Commit()
}

// Rollback ...
func (tx *tx) Rollback() (err error) {
	defer func() {
		// Only log once per tx
		if err == sql.ErrTxDone {
			return
		}
		entry := &LogEntry{
			Ctx:       tx.ctx,
			Error:     err,
			Time:      tx.t0,
			Flags:     Flags(TypeRollback) | FlagTx,
			TxQueries: tx.qs,
		}
		err = tx.db.log(entry)
	}()
	return tx.tx.Rollback()
}

// ExecContext ...
func (tx *tx) ExecContext(ctx context.Context, query string, args ...interface{}) (_ sql.Result, err error) {
	entry := &LogEntry{
		Ctx:   ctx,
		Query: query,
		Args:  args,
		Time:  time.Now(),
		Flags: Flags(TypeExec) | FlagTx,
	}
	tx.qs = append(tx.qs, entry)
	defer func() {
		entry.Error = err
		err = tx.db.log(entry)
	}()
	return tx.tx.Exec(query, args...)
}

// Exec ...
func (tx *tx) Exec(query string, args ...interface{}) (_ sql.Result, err error) {
	return tx.ExecContext(tx.ctx, query, args...)
}

func (tx *tx) QueryContext(ctx context.Context, query string, args ...interface{}) (_ *sql.Rows, err error) {
	entry := &LogEntry{
		Ctx:   ctx,
		Query: query,
		Args:  args,
		Time:  time.Now(),
		Flags: Flags(TypeQuery) | FlagTx,
	}
	tx.qs = append(tx.qs, entry)
	defer func() {
		entry.Error = err
		err = tx.db.log(entry)
	}()
	return tx.tx.Query(query, args...)
}

func (tx *tx) Query(query string, args ...interface{}) (_ *sql.Rows, err error) {
	return tx.QueryContext(tx.ctx, query, args...)
}

func (tx *tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	entry := &LogEntry{
		Ctx:   ctx,
		Query: query,
		Args:  args,
		Time:  time.Now(),
		Flags: Flags(TypeQueryRow) | FlagTx,
	}
	tx.qs = append(tx.qs, entry)
	return Row{
		Row: tx.tx.QueryRow(query, args...),
		Log: func(err error) error {
			entry.Error = err
			return tx.db.log(entry)
		},
	}
}

func (tx *tx) QueryRow(query string, args ...interface{}) Row {
	return tx.QueryRowContext(tx.ctx, query, args...)
}

// Get ...
func (tx *tx) Get(obj core.IGet, preds ...interface{}) (bool, error) {
	return tx.NewQuery().Get(obj, preds...)
}

// Find ...
func (tx *tx) Find(objs core.IFind, preds ...interface{}) error {
	return tx.NewQuery().Find(objs, preds...)
}

// Insert ...
func (tx *tx) Insert(objs ...core.IInsert) (int64, error) {
	return tx.NewQuery().Insert(objs...)
}

// Update ...
func (tx *tx) Update(objs ...core.IUpdate) (int64, error) {
	return tx.NewQuery().Update(objs...)
}

// UpdateMap ...
func (tx *tx) UpdateMap(m map[string]interface{}) (int64, error) {
	return tx.NewQuery().UpdateMap(m)
}

// Delete ...
func (tx *tx) Delete(obj core.ITableName) (int64, error) {
	return tx.NewQuery().Delete(obj)
}

// Count ...
func (tx *tx) Count(obj core.ITableName, preds ...interface{}) (uint64, error) {
	return tx.NewQuery().Count(obj, preds...)
}

// Table ...
func (tx *tx) Table(sql string) Query {
	return tx.NewQuery().Table(sql)
}

// Prefix adds an expression to the start of the query
func (tx *tx) Prefix(sql string, args ...interface{}) Query {
	return tx.NewQuery().Prefix(sql, args...)
}

// Select ...
func (tx *tx) Select(cols ...string) Query {
	return tx.NewQuery().Select(cols...)
}

// From ...
func (tx *tx) From(table string) Query {
	return tx.NewQuery().From(table)
}

// SQL ...
func (tx *tx) SQL(args ...interface{}) Query {
	return tx.NewQuery().SQL(args...)
}

// Where ...
func (tx *tx) Where(args ...interface{}) Query {
	return tx.NewQuery().Where(args...)
}

// OrderBy adds ORDER BY expressions to the query.
func (tx *tx) OrderBy(orderBys ...string) Query {
	return tx.NewQuery().OrderBy(orderBys...)
}

// GroupBy adds GROUP BY expressions to the query.
func (tx *tx) GroupBy(groupBys ...string) Query {
	return tx.NewQuery().GroupBy(groupBys...)
}

// Limit sets a LIMIT clause on the query.
func (tx *tx) Limit(limit uint64) Query {
	return tx.NewQuery().Limit(limit)
}

// Offset sets a OFFSET clause on the query.
func (tx *tx) Offset(offset uint64) Query {
	return tx.NewQuery().Offset(offset)
}

// Suffix adds an expression to the end of the query
func (tx *tx) Suffix(sql string, args ...interface{}) Query {
	return tx.NewQuery().Suffix(sql, args...)
}

// UpdateAll ...
func (tx *tx) UpdateAll() Query {
	return tx.NewQuery().UpdateAll()
}

// In ...
func (tx *tx) In(column string, args ...interface{}) Query {
	return tx.NewQuery().In(column, args...)
}

// NotIn ...
func (tx *tx) NotIn(column string, args ...interface{}) Query {
	return tx.NewQuery().NotIn(column, args...)
}

// Exists ...
func (tx *tx) Exists(column string, exists bool) Query {
	return tx.NewQuery().Exists(column, exists)
}

// IsNull ...
func (tx *tx) IsNull(column string, null bool) Query {
	return tx.NewQuery().IsNull(column, null)
}

func (tx *tx) Preload(table string, preds ...interface{}) Query {
	return tx.NewQuery().Preload(table, preds...)
}

func (tx *tx) Apply(funcs ...func(CommonQuery)) Query {
	return tx.NewQuery().Apply(funcs...)
}
