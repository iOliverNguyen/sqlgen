package sq

import (
	"context"
	"database/sql"
	"time"

	"github.com/ng-vu/sqlgen/core"
)

type Database struct {
	db *sql.DB

	marker core.Marker
	logger Logger
	mapper ErrorMapper
}

func Connect(driver, connStr string, opts ...Option) (*Database, error) {
	_db, err := sql.Open(driver, connStr)
	if err != nil {
		return nil, err
	}
	db := &Database{db: _db, logger: func(_ *LogEntry) {}}

	switch driver {
	case "postgres", "cloudsqlpostgres":
		DollarMarker(db)
	default:
		QuestionMarker(db)
	}
	for _, opt := range opts {
		opt.SQLOption(db)
	}
	return db, nil
}

func MustConnect(driver, connStr string, opts ...Option) *Database {
	db, err := Connect(driver, connStr, opts...)
	if err != nil {
		panic(err)
	}
	return db
}

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

func (db *Database) Exec(query string, args ...interface{}) (_ sql.Result, err error) {
	return db.ExecContext(context.Background(), query, args...)
}

func (db *Database) MustExec(query string, args ...interface{}) sql.Result {
	res, err := db.Exec(query, args...)
	if err != nil {
		panic(err.Error())
	}
	return res
}

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

func (db *Database) Query(query string, args ...interface{}) (_ *sql.Rows, err error) {
	return db.QueryContext(context.Background(), query, args...)
}

type Row struct {
	row *sql.Row
	log func(err error) error
}

func (r Row) Scan(dest ...interface{}) error {
	err := r.row.Scan(dest...)
	return r.log(err)
}

func (db *Database) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	entry := &LogEntry{
		Ctx:   ctx,
		Query: query,
		Args:  args,
		Time:  time.Now(),
		Flags: Flags(TypeQueryRow),
	}
	return Row{
		row: db.db.QueryRowContext(ctx, query, args...),
		log: func(err error) error {
			entry.Error = err
			return db.log(entry)
		},
	}
}

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

func (db *Database) Begin() (Tx, error) {
	t, err := db.db.Begin()
	if err != nil {
		return nil, err
	}
	return &tx{tx: t, db: db, t0: time.Now(), ctx: context.Background()}, nil
}

func (db *Database) BeginContext(ctx context.Context) (Tx, error) {
	t, err := db.db.Begin()
	if err != nil {
		return nil, err
	}
	return &tx{tx: t, db: db, t0: time.Now(), ctx: ctx}, nil
}
