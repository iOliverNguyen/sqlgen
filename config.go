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

// Error ...
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

// Option ...
type Option interface {
	SQLOption(*Database)
}

// OptionFunc ...
type OptionFunc func(*Database)

// SQLOption ...
func (fn OptionFunc) SQLOption(db *Database) {
	fn(db)
}

// QuestionMarker ...
var QuestionMarker = OptionFunc(func(db *Database) {
	db.marker = func() core.IState {
		return questionMarker{}
	}
})

// DollarMarker ...
var DollarMarker = OptionFunc(func(db *Database) {
	db.marker = func() core.IState {
		return &dollarMarker{}
	}
})

// Type ...
type Type int

// Constants ...
const (
	TypeExec     Type = 1
	TypeQuery    Type = 2
	TypeQueryRow Type = 3
	TypeCommit   Type = 5
	TypeRollback Type = 6

	FlagTx    = 1 << 4
	FlagBuild = 1 << 8
)

// Flags ...
type Flags uint

// IsTx ...
func (f Flags) IsTx() bool {
	return f&FlagTx > 0
}

// IsBuild ...
func (f Flags) IsBuild() bool {
	return f&FlagBuild > 0
}

// IsQuery ...
func (f Flags) IsQuery() bool {
	return f.Type() <= 3
}

// Type ...
func (f Flags) Type() Type {
	return Type(f) & (FlagTx - 1)
}

// MarshalJSON ...
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

// LogArgs ...
type LogArgs []interface{}

// ToSQLValues ...
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

// MarshalJSON ...
func (args LogArgs) MarshalJSON() ([]byte, error) {
	res, _ := args.ToSQLValues()
	return json.Marshal(res)
}

// LogEntry ...
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

// Logger ...
type Logger func(*LogEntry)

// SQLOption ...
func (l Logger) SQLOption(db *Database) {
	db.logger = l
}

// SetLogger ...
func SetLogger(logger Logger) Option {
	return logger
}

// DefaultLogger ...
var DefaultLogger = Logger(func(entry *LogEntry) {
	log.Print("query=`", entry.Query, "` arg=", entry.Args, " error=", entry.Error, " t=", entry.Duration)
})

// DynamicLogger ...
type DynamicLogger struct {
	logger Logger
	m      sync.RWMutex
}

// NewDynamicLogger ...
func NewDynamicLogger(logger Logger) *DynamicLogger {
	return &DynamicLogger{logger: logger}
}

// SQLOption ...
func (d *DynamicLogger) SQLOption(db *Database) {
	db.logger = d.log
}

// SetLogger ...
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

// ErrorMapper ...
type ErrorMapper func(error, *LogEntry) error

// SQLOption ...
func (m ErrorMapper) SQLOption(db *Database) {
	db.mapper = m
}

// SetErrorMapper ...
func SetErrorMapper(mapper ErrorMapper) Option {
	return mapper
}
