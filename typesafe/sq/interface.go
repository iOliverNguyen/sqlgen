package sq

import (
	"context"
	"database/sql"
	"time"

	core "github.com/ng-vu/sqlgen/core"
)

type DBInterface interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
}

type dbInterface interface {
	DBInterface
	log(*LogEntry) error
}

type QueryInterface interface {
	Get(obj core.IGet) (bool, error)
	Find(objs core.IFind) error
	Insert(objs ...core.IInsert) (int64, error)
	Update(objs ...core.IUpdate) (int64, error)
	UpdateMap(m map[string]interface{}) (int64, error)
	Delete(obj core.ITableName) (int64, error)
	Count(obj core.ITableName) (uint64, error)

	Table(name string) *Query
	Prefix(sql string, args ...interface{}) *Query
	Select(cols ...string) *Query
	From(table string) *Query
	SQL(sql string, args ...interface{}) *Query
	Where(cond string, args ...interface{}) *Query
	OrderBy(orderBys ...string) *Query
	GroupBy(groupBys ...string) *Query
	Limit(limit uint64) *Query
	Offset(offset uint64) *Query
	Suffix(sql string, args ...interface{}) *Query
	UpdateAll() *Query
	In(column string, args ...interface{}) *Query
}

func (db *Database) Get(obj core.IGet) (bool, error) {
	return db.NewQuery().Get(obj)
}

func (db *Database) Find(objs core.IFind) error {
	return db.NewQuery().Find(objs)
}

func (db *Database) Insert(objs ...core.IInsert) (int64, error) {
	return db.NewQuery().Insert(objs...)
}

func (db *Database) Update(objs ...core.IUpdate) (int64, error) {
	return db.NewQuery().Update(objs...)
}

func (db *Database) UpdateMap(m map[string]interface{}) (int64, error) {
	return db.NewQuery().UpdateMap(m)
}

func (db *Database) Delete(obj core.ITableName) (int64, error) {
	return db.NewQuery().Delete(obj)
}

func (db *Database) Count(obj core.ITableName) (uint64, error) {
	return db.NewQuery().Count(obj)
}

func (db *Database) Table(sql string) *Query {
	return db.NewQuery().Table(sql)
}

// Prefix adds an expression to the start of the query
func (db *Database) Prefix(sql string, args ...interface{}) *Query {
	return db.NewQuery().Prefix(sql, args...)
}

func (db *Database) Select(cols ...string) *Query {
	return db.NewQuery().Select(cols...)
}

func (db *Database) From(table string) *Query {
	return db.NewQuery().From(table)
}

func (db *Database) SQL(sql string, args ...interface{}) *Query {
	return db.NewQuery().SQL(sql, args...)
}

func (db *Database) Where(cond string, args ...interface{}) *Query {
	return db.NewQuery().Where(cond, args...)
}

// OrderBy adds ORDER BY expressions to the query.
func (db *Database) OrderBy(orderBys ...string) *Query {
	return db.NewQuery().OrderBy(orderBys...)
}

// GroupBy adds GROUP BY expressions to the query.
func (db *Database) GroupBy(groupBys ...string) *Query {
	return db.NewQuery().GroupBy(groupBys...)
}

// Limit sets a LIMIT clause on the query.
func (db *Database) Limit(limit uint64) *Query {
	return db.NewQuery().Limit(limit)
}

// Offset sets a OFFSET clause on the query.
func (db *Database) Offset(offset uint64) *Query {
	return db.NewQuery().Offset(offset)
}

// Suffix adds an expression to the end of the query
func (db *Database) Suffix(sql string, args ...interface{}) *Query {
	return db.NewQuery().Suffix(sql, args...)
}

func (db *Database) UpdateAll() *Query {
	return db.NewQuery().UpdateAll()
}

func (db *Database) In(column string, args ...interface{}) *Query {
	return db.NewQuery().In(column, args...)
}

type Tx interface {
	Commit() error
	Rollback() error

	DBInterface
	QueryInterface
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
		row: tx.tx.QueryRow(query, args...),
		log: func(err error) error {
			entry.Error = err
			return tx.db.log(entry)
		},
	}
}

func (tx *tx) QueryRow(query string, args ...interface{}) Row {
	return tx.QueryRowContext(tx.ctx, query, args...)
}

func (tx *tx) Get(obj core.IGet) (bool, error) {
	return tx.NewQuery().Get(obj)
}

func (tx *tx) Find(objs core.IFind) error {
	return tx.NewQuery().Find(objs)
}

func (tx *tx) Insert(objs ...core.IInsert) (int64, error) {
	return tx.NewQuery().Insert(objs...)
}

func (tx *tx) Update(objs ...core.IUpdate) (int64, error) {
	return tx.NewQuery().Update(objs...)
}

func (tx *tx) UpdateMap(m map[string]interface{}) (int64, error) {
	return tx.NewQuery().UpdateMap(m)
}

func (tx *tx) Delete(obj core.ITableName) (int64, error) {
	return tx.NewQuery().Delete(obj)
}

func (tx *tx) Count(obj core.ITableName) (uint64, error) {
	return tx.NewQuery().Count(obj)
}

func (tx *tx) Table(sql string) *Query {
	return tx.NewQuery().Table(sql)
}

// Prefix adds an expression to the start of the query
func (tx *tx) Prefix(sql string, args ...interface{}) *Query {
	return tx.NewQuery().Prefix(sql, args...)
}

func (tx *tx) Select(cols ...string) *Query {
	return tx.NewQuery().Select(cols...)
}

func (tx *tx) From(table string) *Query {
	return tx.NewQuery().From(table)
}

func (tx *tx) SQL(sql string, args ...interface{}) *Query {
	return tx.NewQuery().SQL(sql, args...)
}

func (tx *tx) Where(cond string, args ...interface{}) *Query {
	return tx.NewQuery().Where(cond, args...)
}

// OrderBy adds ORDER BY expressions to the query.
func (tx *tx) OrderBy(orderBys ...string) *Query {
	return tx.NewQuery().OrderBy(orderBys...)
}

// GroupBy adds GROUP BY expressions to the query.
func (tx *tx) GroupBy(groupBys ...string) *Query {
	return tx.NewQuery().GroupBy(groupBys...)
}

// Limit sets a LIMIT clause on the query.
func (tx *tx) Limit(limit uint64) *Query {
	return tx.NewQuery().Limit(limit)
}

// Offset sets a OFFSET clause on the query.
func (tx *tx) Offset(offset uint64) *Query {
	return tx.NewQuery().Offset(offset)
}

// Suffix adds an expression to the end of the query
func (tx *tx) Suffix(sql string, args ...interface{}) *Query {
	return tx.NewQuery().Suffix(sql, args...)
}

func (tx *tx) UpdateAll() *Query {
	return tx.NewQuery().UpdateAll()
}

func (tx *tx) In(column string, args ...interface{}) *Query {
	return tx.NewQuery().In(column, args...)
}
