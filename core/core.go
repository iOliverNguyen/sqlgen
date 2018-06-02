package core

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/lib/pq"
)

// Error ...
type Error string

func (e Error) Error() string {
	return string(e)
}

// Errorf ...
func Errorf(format string, args ...interface{}) Error {
	return Error(fmt.Sprintf(format, args...))
}

// Errors
var (
	ErrNoColumn     = Error("common/sql: No column to update")
	ErrNoAction     = Error("common/sql: Must provide an action")
	ErrNoRows       = sql.ErrNoRows
	ErrNoRowsUpdate = Error("common/sql: No row was updated")
	ErrNoRowsInsert = Error("common/sql: No row was inserted")
	ErrNoRowsDelete = Error("common/sql: No row was deleted")
)

// JoinType ...
type JoinType string

// ITableName ...
type ITableName interface {
	SQLTableName() string
}

// IScan ...
type IScan interface {
	ITableName
	SQLScan(*sql.Rows) error
}

// IScanRow ...
type IScanRow interface {
	ITableName
	SQLScan(*sql.Row) error
}

// ISelect ...
type ISelect interface {
	ITableName
	SQLSelect([]byte) []byte
}

// IInsert ...
type IInsert interface {
	ITableName
	SQLInsert(IState, []byte, []interface{}) ([]byte, []interface{}, error)
}

// IUpdate ...
type IUpdate interface {
	ITableName
	SQLUpdate(IState, []byte, []interface{}) ([]byte, []interface{}, error)
	SQLUpdateAll(IState, []byte, []interface{}) ([]byte, []interface{}, error)
}

// IGet ...
type IGet interface {
	ITableName
	SQLSelect([]byte) []byte
	SQLScan(*sql.Row) error
}

// IFind ...
type IFind interface {
	ITableName
	SQLSelect([]byte) []byte
	SQLScan(*sql.Rows) error
}

// IJoin  ...
type IJoin interface {
	ITableName
	SQLJoin([]byte, []JoinType) []byte
}

// IState ...
type IState interface {
	AppendMarker([]byte, int) []byte
	AppendQuery([]byte, []byte) []byte
	AppendQueryStr([]byte, string) []byte
}

// Marker ...
type Marker func() IState

// Now ...
func Now(t, now time.Time, update bool) time.Time {
	if update && (t.IsZero() || t.Equal(zeroTime)) {
		return now
	}
	return t
}

// NowP ...
func NowP(t *time.Time, now time.Time, update bool) time.Time {
	if update && t == nil {
		return now
	}
	return *t
}

// String handles null as empty string
type String string

func (s String) String() string {
	return string(s)
}

// Scan implements the Scanner interface.
func (s *String) Scan(src interface{}) error {
	var ns sql.NullString
	err := ns.Scan(src)
	if err == nil && ns.Valid {
		*s = String(ns.String)
	}
	return err
}

// Value implements the driver Valuer interface.
func (s String) Value() (driver.Value, error) {
	if s == "" {
		return nil, nil
	}
	return string(s), nil
}

// Int handles null as 0 and stores 0 as is.
type Int int

// Scan implements the Scanner interface.
func (i *Int) Scan(src interface{}) error {
	var ni sql.NullInt64
	err := ni.Scan(src)
	if err == nil && ni.Valid {
		*i = Int(ni.Int64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (i Int) Value() (driver.Value, error) {
	if i == 0 {
		return int64(0), nil
	}
	return int64(i), nil
}

// Int64 handles null as 0 but stores 0 as null. It's because int64 is usually
// used for identifier.
type Int64 int64

// Scan implements the Scanner interface.
func (i *Int64) Scan(src interface{}) error {
	var ni sql.NullInt64
	err := ni.Scan(src)
	if err == nil && ni.Valid {
		*i = Int64(ni.Int64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (i Int64) Value() (driver.Value, error) {
	if i == 0 {
		return nil, nil
	}
	return int64(i), nil
}

// Float handles null as 0
type Float float64

// Scan implements the Scanner interface.
func (f *Float) Scan(src interface{}) error {
	var nf sql.NullFloat64
	err := nf.Scan(src)
	if err == nil && nf.Valid {
		*f = Float(nf.Float64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (f Float) Value() (driver.Value, error) {
	if f == 0 {
		return float64(0), nil
	}
	return float64(f), nil
}

// Bool handles null as false
type Bool bool

// Scan implements the Scanner interface.
func (b *Bool) Scan(src interface{}) error {
	var nb sql.NullBool
	err := nb.Scan(src)
	if err == nil && nb.Valid {
		*b = Bool(nb.Bool)
	}
	return err
}

// Value implements the driver Valuer interface.
func (b Bool) Value() (driver.Value, error) {
	if !b {
		return bool(false), nil
	}
	return bool(b), nil
}

// Time ...
type Time time.Time

// Scan implements the Scanner interface.
func (t *Time) Scan(src interface{}) error {
	tt, _ := src.(time.Time)
	*t = Time(tt)
	return nil
}

var zeroTime = time.Unix(0, 0)

// Value implements the driver Valuer interface.
func (t Time) Value() (driver.Value, error) {
	tt := time.Time(t)
	if tt.IsZero() || tt.Equal(zeroTime) {
		return nil, nil
	}
	return tt, nil
}

// JSON ...
type JSON struct {
	V interface{}
}

// Scan implements the Scanner interface.
func (v JSON) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	switch value := src.(type) {
	case []byte:
		if len(value) == 0 {
			return nil
		}
		return json.Unmarshal(value, v.V)
	case string:
		if len(value) == 0 {
			return nil
		}
		return json.Unmarshal([]byte(value), v.V)
	default:
		return fmt.Errorf("common/sql: Unsupported json source %v", reflect.TypeOf(src))
	}
}

// Value implements the driver Valuer interface.
func (v JSON) Value() (driver.Value, error) {
	if v.V == nil || reflect.ValueOf(v.V).IsNil() {
		return nil, nil
	}
	if v, ok := v.V.(json.RawMessage); ok {
		if len(v) == 0 {
			return nil, nil
		}
		return []byte(v), nil
	}
	data, err := json.Marshal(v.V)
	return data, err
}

// Array ...
type Array struct {
	V interface{}
}

// Scan implements the Scanner interface.
func (a Array) Scan(src interface{}) error {

	// TODO(qv): mysql
	switch a.V.(type) {
	case *[]int:
		var int64s pq.Int64Array
		err := int64s.Scan(src)
		if err != nil {
			return err
		}
		if len(int64s) == 0 {
			return nil
		}
		ints := make([]int, len(int64s))
		for i := range int64s {
			ints[i] = int(int64s[i])
		}
		reflect.ValueOf(a.V).Elem().Set(reflect.ValueOf(ints))
		return nil

	case *[]time.Time:
		var ss pq.StringArray
		err := ss.Scan(src)
		if err != nil {
			return err
		}
		if len(ss) == 0 {
			return nil
		}
		times := make([]time.Time, len(ss))
		for i, s := range ss {
			times[i], err = parseTime(s)
			if err != nil {
				return err
			}
		}
		reflect.ValueOf(a.V).Elem().Set(reflect.ValueOf(times))
		return nil

	case *[]*time.Time:
		var ss pq.StringArray
		err := ss.Scan(src)
		if err != nil {
			return err
		}
		if len(ss) == 0 {
			return nil
		}
		times := make([]*time.Time, len(ss))
		for i, s := range ss {
			t, err := parseTime(s)
			if err != nil {
				return err
			}
			times[i] = &t
		}
		reflect.ValueOf(a.V).Elem().Set(reflect.ValueOf(times))
		return nil
	}

	return pq.Array(a.V).Scan(src)
}

var timeLayout = `2006-01-02 15:04:05.000`

func parseTime(s string) (time.Time, error) {
	s = strings.TrimSuffix(s, "+00")
	return time.Parse(timeLayout, s)
}

func pTime(t time.Time) *time.Time {
	return &t
}

// Value implements the driver Valuer interface.
func (a Array) Value() (driver.Value, error) {
	v, err := pq.Array(a.V).Value()
	return v, err
}

// Map ...
type Map struct {
	Table string
	M     map[string]interface{}
}

// SQLTableName ...
func (m Map) SQLTableName() string { return m.Table }

// SQLUpdate ...
func (m Map) SQLUpdate(s IState, b []byte, args []interface{}) ([]byte, []interface{}, error) {
	return m.SQLUpdateAll(s, b, args)
}

// SQLUpdateAll ...
func (m Map) SQLUpdateAll(s IState, b []byte, args []interface{}) ([]byte, []interface{}, error) {
	b = append(b, `UPDATE "`...)
	b = append(b, m.Table...)
	switch len(m.M) {
	case 0:
		return b, args, errors.New("common/sql: No column to update")
	case 1:
		b = append(b, `" SET `...)
		for k, v := range m.M {
			b = append(b, '"')
			b = append(b, k...)
			b = append(b, `",`...)
			args = append(args, v)
		}
		b = b[:len(b)-1]
		b = append(b, " = "...)
		b = s.AppendMarker(b, len(m.M))
	default:
		b = append(b, `" SET (`...)
		for k, v := range m.M {
			b = append(b, '"')
			b = append(b, k...)
			b = append(b, `",`...)
			args = append(args, v)
		}
		b = b[:len(b)-1]
		b = append(b, ") = ("...)
		b = s.AppendMarker(b, len(m.M))
		b = append(b, ')')
	}

	return b, args, nil
}

// AppendCols ...
func AppendCols(b []byte, prefix string, cols string) []byte {
	b = append(b, prefix...)
	for i, l := 0, len(cols); i < l; i++ {
		ch := cols[i]
		b = append(b, ch)
		if ch == ',' {
			b = append(b, prefix...)
		}
	}
	return b
}

// Ternary ...
func Ternary(cond bool, exp1, exp2 interface{}) interface{} {
	if cond {
		return exp1
	}
	return exp2
}
