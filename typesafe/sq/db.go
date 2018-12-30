package sq

import (
	"context"
	"database/sql"
	"time"

	"github.com/ng-vu/sqlgen/core"
)

// Database ...
type Database struct {
	db *sql.DB

	opts   core.Opts
	quote  byte
	marker byte
	logger Logger
	mapper ErrorMapper
}

// Connect ...
func Connect(driver, connStr string, opts ...Option) (*Database, error) {
	_db, err := sql.Open(driver, connStr)
	if err != nil {
		return nil, err
	}
	db := &Database{db: _db, logger: func(_ *LogEntry) {}}

	switch driver {
	case "postgres", "cloudsqlpostgres":
		DollarMarker(db)
		DoubleQuoteEscape(db)
		UseArrayInsteadOfJSON(db, true)
	default:
		QuestionMarker(db)
		BacktickEscape(db)
	}
	for _, opt := range opts {
		opt.SQLOption(db)
	}
	return db, nil
}

// MustConnect ...
func MustConnect(driver, connStr string, opts ...Option) *Database {
	db, err := Connect(driver, connStr, opts...)
	if err != nil {
		panic(err)
	}
	return db
}

func (db *Database) DB() *sql.DB {
	return db.db
}

func (db *Database) Opts() core.Opts {
	return db.opts
}

// ExecContext ...
func (db *Database) ExecContext(ctx context.Context, query string, args ...interface{}) (_ sql.Result, err error) {
	entry := &LogEntry{
		Ctx:   ctx,
		Query: query,
		Args:  args,
		Time:  time.Now(),
		Flags: Flags(TypeExec),
	}
	defer func() {
		entry.Error = err
		err = db.log(entry)
	}()
	return db.db.ExecContext(ctx, query, args...)
}

// Exec ...
func (db *Database) Exec(query string, args ...interface{}) (_ sql.Result, err error) {
	return db.ExecContext(context.Background(), query, args...)
}

// MustExec ...
func (db *Database) MustExec(query string, args ...interface{}) sql.Result {
	res, err := db.Exec(query, args...)
	if err != nil {
		panic(err.Error())
	}
	return res
}

// QueryContext ...
func (db *Database) QueryContext(ctx context.Context, query string, args ...interface{}) (_ *sql.Rows, err error) {
	entry := &LogEntry{
		Ctx:   ctx,
		Query: query,
		Args:  args,
		Time:  time.Now(),
		Flags: Flags(TypeQuery),
	}
	defer func() {
		entry.Error = err
		err = db.log(entry)
	}()
	return db.db.QueryContext(ctx, query, args...)
}

// Query ...
func (db *Database) Query(query string, args ...interface{}) (_ *sql.Rows, err error) {
	return db.QueryContext(context.Background(), query, args...)
}

// QueryRowContext ...
func (db *Database) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	entry := &LogEntry{
		Ctx:   ctx,
		Query: query,
		Args:  args,
		Time:  time.Now(),
		Flags: Flags(TypeQueryRow),
	}
	return Row{
		Row: db.db.QueryRowContext(ctx, query, args...),
		Log: func(err error) error {
			entry.Error = err
			return db.log(entry)
		},
	}
}

// QueryRow ...
func (db *Database) QueryRow(query string, args ...interface{}) Row {
	return db.QueryRowContext(context.Background(), query, args...)
}

func (db *Database) log(entry *LogEntry) (err error) {
	err = entry.Error
	entry.Duration = time.Now().Sub(entry.Time)
	if db.mapper != nil {
		entry.OrigError = err
		err = db.mapper(err, entry)
		entry.Error = err
	}
	db.logger(entry)
	return
}

// Begin ...
func (db *Database) Begin() (Tx, error) {
	t, err := db.db.Begin()
	if err != nil {
		return nil, err
	}
	return &tx{tx: t, db: db, t0: time.Now(), ctx: context.Background()}, nil
}

// BeginContext ...
func (db *Database) BeginContext(ctx context.Context) (Tx, error) {
	t, err := db.db.Begin()
	if err != nil {
		return nil, err
	}
	return &tx{tx: t, db: db, t0: time.Now(), ctx: ctx}, nil
}
