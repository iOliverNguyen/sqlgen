package sq

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"log"
	"sync"
	"time"

	core "github.com/ng-vu/sqlgen/core"
)

type Error = core.Error

// Errors
var (
	ErrNoColumn     = core.ErrNoColumn
	ErrNoAction     = core.ErrNoAction
	ErrNoRows       = core.ErrNoRows
	ErrNoRowsUpdate = core.ErrNoRowsUpdate
	ErrNoRowsInsert = core.ErrNoRowsInsert
	ErrNoRowsDelete = core.ErrNoRowsDelete
)

type Option interface {
	SQLOption(*Database)
}

type OptionFunc func(*Database)

func (fn OptionFunc) SQLOption(db *Database) {
	fn(db)
}

var QuestionMarker = OptionFunc(func(db *Database) {
	db.marker = func() core.IState {
		return questionMarker{}
	}
})

var DollarMarker = OptionFunc(func(db *Database) {
	db.marker = func() core.IState {
		return &dollarMarker{}
	}
})

type Type int

const (
	TypeExec     Type = 1
	TypeQuery    Type = 2
	TypeQueryRow Type = 3
	TypeCommit   Type = 5
	TypeRollback Type = 6

	FlagTx    = 1 << 4
	FlagBuild = 1 << 8
)

type Flags uint

func (f Flags) IsTx() bool {
	return f&FlagTx > 0
}

func (f Flags) IsBuild() bool {
	return f&FlagBuild > 0
}

func (f Flags) IsQuery() bool {
	return f.Type() <= 3
}

func (f Flags) Type() Type {
	return Type(f) & (FlagTx - 1)
}

func (f Flags) MarshalJSON() ([]byte, error) {
	b := make([]byte, 1, 4)
	switch f.Type() {
	case TypeExec:
		b[0] = 'E'
	case TypeQuery:
		b[0] = 'Q'
	case TypeQueryRow:
		b[0] = 'q'
	case TypeCommit:
		b[0] = 'C'
	case TypeRollback:
		b[0] = 'R'
	default:
		b[0] = '_'
	}
	if f.IsTx() {
		b = append(b, 'x')
	}
	if f.IsBuild() {
		b = append(b, 'B')
	}
	return b, nil
}

type LogArgs []interface{}

func (args LogArgs) ToSQLValues() (res []interface{}, _err error) {
	res = make([]interface{}, len(args))
	for i, arg := range args {
		if v, ok := arg.(driver.Valuer); ok {
			var err error
			arg, err = v.Value()
			if err != nil {
				arg = err
				_err = err
			}
		}
		res[i] = arg
	}
	return
}

func (args LogArgs) MarshalJSON() ([]byte, error) {
	res, _ := args.ToSQLValues()
	return json.Marshal(res)
}

type LogEntry struct {
	Ctx       context.Context `json:"-"`
	Query     string          `json:"query"`
	Args      LogArgs         `json:"args"`
	Error     error           `json:"error"`
	OrigError error           `json:"orig_error"`
	Time      time.Time       `json:"time"`
	Duration  time.Duration   `json:"duration"`

	Flags `json:"flags"`

	// Only be set if Type is Commit or Revert
	TxQueries []*LogEntry `json:"tx_queries"`
}

type Logger func(*LogEntry)

func (l Logger) SQLOption(db *Database) {
	db.logger = l
}

func SetLogger(logger Logger) Option {
	return logger
}

var DefaultLogger = Logger(func(entry *LogEntry) {
	log.Print("query=`", entry.Query, "` arg=", entry.Args, " error=", entry.Error, " t=", entry.Duration)
})

type DynamicLogger struct {
	logger Logger
	m      sync.RWMutex
}

func NewDynamicLogger(logger Logger) *DynamicLogger {
	return &DynamicLogger{logger: logger}
}

func (d *DynamicLogger) SQLOption(db *Database) {
	db.logger = d.log
}

func (d *DynamicLogger) SetLogger(logger Logger) {
	d.m.Lock()
	d.logger = logger
	d.m.Unlock()
}

func (d *DynamicLogger) log(entry *LogEntry) {
	d.m.RLock()
	logger := d.logger
	d.m.RUnlock()

	if logger != nil {
		logger(entry)
	}
}

type ErrorMapper func(error, *LogEntry) error

func (m ErrorMapper) SQLOption(db *Database) {
	db.mapper = m
}

func SetErrorMapper(mapper ErrorMapper) Option {
	return mapper
}
