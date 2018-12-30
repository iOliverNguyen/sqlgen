package core

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/lib/pq"
)

type Row struct {
	Row *sql.Row
	Log func(err error) error
}

func (r Row) Scan(dest ...interface{}) error {
	err := r.Row.Scan(dest...)
	return r.Log(err)
}

// QueryInterface ...
type CommonQuery interface {
	Get(obj IGet, preds ...interface{}) (bool, error)
	Find(objs IFind, preds ...interface{}) error
	Insert(objs ...IInsert) (int64, error)
	Update(objs ...IUpdate) (int64, error)
	UpdateMap(m map[string]interface{}) (int64, error)
	Delete(obj ITableName) (int64, error)
	Count(obj ITableName, preds ...interface{}) (uint64, error)

	Table(name string) Query
	Prefix(sql string, args ...interface{}) Query
	Select(cols ...string) Query
	From(table string) Query
	SQL(preds ...interface{}) Query
	Where(preds ...interface{}) Query
	OrderBy(orderBys ...string) Query
	GroupBy(groupBys ...string) Query
	Limit(limit uint64) Query
	Offset(offset uint64) Query
	Suffix(sql string, args ...interface{}) Query
	UpdateAll() Query
	In(column string, args ...interface{}) Query
	NotIn(column string, args ...interface{}) Query
	Exists(column string, exists bool) Query
	IsNull(column string, null bool) Query

	Preload(table string, preds ...interface{}) Query
	Apply(funcs ...func(CommonQuery)) Query
}

// DBInterface ...
type DBInterface interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
}

// Query ...
type Query interface {
	CommonQuery

	Build(preds ...interface{}) (string, []interface{}, error)
	BuildGet(obj IGet, preds ...interface{}) (string, []interface{}, error)
	BuildFind(objs IFind, preds ...interface{}) (string, []interface{}, error)
	BuildInsert(obj IInsert) (string, []interface{}, error)
	BuildUpdate(obj IUpdate) (string, []interface{}, error)
	BuildDelete(obj ITableName) (string, []interface{}, error)
	BuildCount(obj ITableName, preds ...interface{}) (string, []interface{}, error)
	Clone() Query
	Exec() (sql.Result, error)
	Query() (*sql.Rows, error)
	QueryRow() (Row, error)
	Scan(dest ...interface{}) error
	WithContext(context.Context) Query
}

// Error ...
type Error string

func (e Error) Error() string {
	return string(e)
}

// InvalidArgumentError ...
type InvalidArgumentError string

func (e InvalidArgumentError) Error() string {
	return string(e)
}

// Errorf ...
func Errorf(format string, args ...interface{}) Error {
	return Error(fmt.Sprintf(format, args...))
}

// Errors
var (
	ErrNoColumn = Error("sqlgen: no column to update")
	ErrNoRows   = sql.ErrNoRows
)

type Opts struct {
	UseArrayInsteadOfJSON bool
}

func (opts Opts) Array(v interface{}) Array {
	return Array{V: v, Opts: opts}
}

func (opts Opts) JSON(v interface{}) JSON {
	return JSON{V: v}
}

// JoinType ...
type JoinType string

// ITableName ...
type ITableName interface {
	SQLTableName() string
}

// IScan ...
type IScan interface {
	ITableName
	SQLScan(Opts, *sql.Rows) error
}

// IScanRow ...
type IScanRow interface {
	ITableName
	SQLScan(Opts, *sql.Row) error
}

// ISelect ...
type ISelect interface {
	ITableName
	SQLSelect(SQLWriter) error
}

// IInsert ...
type IInsert interface {
	ITableName
	SQLInsert(SQLWriter) error
}

// IUpdate ...
type IUpdate interface {
	ITableName
	SQLUpdate(SQLWriter) error
	SQLUpdateAll(SQLWriter) error
}

// IGet ...
type IGet interface {
	ITableName
	SQLSelect(SQLWriter) error
	SQLScan(Opts, *sql.Row) error
}

// IFind ...
type IFind interface {
	ITableName
	SQLSelect(SQLWriter) error
	SQLScan(Opts, *sql.Rows) error
}

// IJoin  ...
type IJoin interface {
	ITableName
	SQLJoin(SQLWriter, []JoinType) error
}

type IPreload interface {
	SQLPreload(name string) *PreloadDesc
	SQLPopulate(items IFind) error
}

type PreloadDesc struct {
	Fkey  string
	IDs   interface{}
	Items IFind
}

type SQLWriter interface {
	Len() int
	Opts() Opts
	TrimLast(n int)
	WriteArg(arg interface{})
	WriteArgs(args []interface{})
	WriteByte(b byte)
	WriteMarker()
	WriteMarkers(n int)
	WriteName(name string)
	WritePrefixedName(schema string, name string)
	WriteQuery(b []byte)
	WriteQueryName(name string)
	WriteQueryString(s string)
	WriteQueryStringWithPrefix(prefix string, s string)
	WriteRaw(b []byte)
	WriteRawString(s string)
	WriteScanArg(arg interface{})
	WriteScanArgs(args []interface{})
}

type Interface struct{ V interface{} }

func (i Interface) Bool() *bool {
	if i.V == nil {
		return nil
	}
	v := i.V.(bool)
	return &v
}

func (i Interface) Int() *int {
	if i.V == nil {
		return nil
	}
	v := int(i.V.(int64))
	return &v
}

func (i Interface) Int32() *int32 {
	if i.V == nil {
		return nil
	}
	v := int32(i.V.(int64))
	return &v
}

func (i Interface) Int64() *int64 {
	if i.V == nil {
		return nil
	}
	v := i.V.(int64)
	return &v
}

func (i Interface) String() *string {
	if i.V == nil {
		return nil
	}
	switch v := i.V.(type) {
	case string:
		return &v
	case []byte:
		res := unsafeBytesToString(v)
		return &res
	default:
		panic(fmt.Sprintf("sqlgen: unknown type %v", reflect.TypeOf(i.V)))
	}
}

func (i Interface) Time() time.Time {
	if i.V == nil {
		return time.Time{}
	}
	v := i.V.(time.Time)
	return v
}

func (i Interface) JSON() json.RawMessage {
	if i.V == nil {
		return nil
	}
	v := i.V.([]byte)
	return v
}

func (i Interface) Unmarshal(v interface{}) error {
	if i.V == nil {
		return nil
	}
	err := json.Unmarshal(i.V.([]byte), v)
	if err != nil {
		log.Println("sqlgen: error unmarshalling", err)
	}
	return err
}

func (i Interface) Map() map[string]interface{} {
	if i.V == nil {
		return nil
	}
	v := i.V.(map[string]interface{})
	return v
}

// Now ...
func Now(t, now time.Time, update bool) time.Time {
	if update && t.IsZero() {
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
	return int64(i), nil
}

// Int8 ...
type Int8 int8

// Scan ...
func (i *Int8) Scan(src interface{}) error {
	var ni sql.NullInt64
	err := ni.Scan(src)
	if err == nil && ni.Valid {
		*i = Int8(ni.Int64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (i Int8) Value() (driver.Value, error) {
	return int64(i), nil
}

// Int16 ...
type Int16 int16

// Scan ...
func (i *Int16) Scan(src interface{}) error {
	var ni sql.NullInt64
	err := ni.Scan(src)
	if err == nil && ni.Valid {
		*i = Int16(ni.Int64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (i Int16) Value() (driver.Value, error) {
	return int64(i), nil
}

// Int32 ...
type Int32 int32

// Scan ...
func (i *Int32) Scan(src interface{}) error {
	var ni sql.NullInt64
	err := ni.Scan(src)
	if err == nil && ni.Valid {
		*i = Int32(ni.Int64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (i Int32) Value() (driver.Value, error) {
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

// Uint handles null as 0 and stores 0 as is.
type Uint uint

// Scan implements the Scanner interface.
func (i *Uint) Scan(src interface{}) error {
	var ni sql.NullInt64
	err := ni.Scan(src)
	if err == nil && ni.Valid {
		*i = Uint(ni.Int64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (i Uint) Value() (driver.Value, error) {
	return int64(i), nil
}

// Uint8 ...
type Uint8 uint8

// Scan ...
func (i *Uint8) Scan(src interface{}) error {
	var ni sql.NullInt64
	err := ni.Scan(src)
	if err == nil && ni.Valid {
		*i = Uint8(ni.Int64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (i Uint8) Value() (driver.Value, error) {
	return int64(i), nil
}

// Uint16 ...
type Uint16 uint16

// Scan ...
func (i *Uint16) Scan(src interface{}) error {
	var ni sql.NullInt64
	err := ni.Scan(src)
	if err == nil && ni.Valid {
		*i = Uint16(ni.Int64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (i Uint16) Value() (driver.Value, error) {
	return int64(i), nil
}

// Uint32 ...
type Uint32 uint32

// Scan ...
func (i *Uint32) Scan(src interface{}) error {
	var ni sql.NullInt64
	err := ni.Scan(src)
	if err == nil && ni.Valid {
		*i = Uint32(ni.Int64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (i Uint32) Value() (driver.Value, error) {
	return int64(i), nil
}

// Uint64 handles null as 0 but stores 0 as null. It's because uint64 is usually
// used for identifier.
type Uint64 uint64

// Scan implements the Scanner interface.
func (i *Uint64) Scan(src interface{}) error {
	var ni sql.NullInt64
	err := ni.Scan(src)
	if err == nil && ni.Valid {
		*i = Uint64(ni.Int64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (i Uint64) Value() (driver.Value, error) {
	if i == 0 {
		return nil, nil
	}
	return int64(i), nil
}

// Float32 handles null as 0
type Float32 float32

// Scan implements the Scanner interface.
func (f *Float32) Scan(src interface{}) error {
	var nf sql.NullFloat64
	err := nf.Scan(src)
	if err == nil && nf.Valid {
		*f = Float32(nf.Float64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (f Float32) Value() (driver.Value, error) {
	return float64(f), nil
}

// Float64 handles null as 0
type Float64 float64

// Float is alias to Float64
type Float = Float64

// Scan implements the Scanner interface.
func (f *Float64) Scan(src interface{}) error {
	var nf sql.NullFloat64
	err := nf.Scan(src)
	if err == nil && nf.Valid {
		*f = Float64(nf.Float64)
	}
	return err
}

// Value implements the driver Valuer interface.
func (f Float64) Value() (driver.Value, error) {
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
		return false, nil
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

// Value implements the driver Valuer interface.
func (t Time) Value() (driver.Value, error) {
	tt := time.Time(t)
	if tt.IsZero() {
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
		return fmt.Errorf("sqlgen: unsupported json source %v", reflect.TypeOf(src))
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

func ArrayScanner(v interface{}) Array {
	return Array{V: v, Opts: Opts{UseArrayInsteadOfJSON: true}}
}

// Array ...
type Array struct {
	V interface{}
	Opts
}

// Scan implements the Scanner interface.
func (a Array) Scan(src interface{}) error {
	if !a.UseArrayInsteadOfJSON {
		return JSON{a.V}.Scan(src)
	}

	switch v := a.V.(type) {
	case *[]int64:
		return (*pq.Int64Array)(v).Scan(src)

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
		*v = ints
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
		*v = times
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
		*v = times
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
func (m Map) SQLUpdate(w SQLWriter) error {
	return m.SQLUpdateAll(w)
}

// SQLUpdateAll ...
func (m Map) SQLUpdateAll(w SQLWriter) error {
	w.WriteRawString(`UPDATE `)
	w.WriteName(m.Table)
	switch len(m.M) {
	case 0:
		return ErrNoColumn
	case 1:
		w.WriteRawString(` SET `)
		for k, v := range m.M {
			w.WriteName(k)
			w.WriteArg(v)
		}
		w.WriteRawString(` = `)
		w.WriteMarker()
	default:
		w.WriteQueryString(` SET (`)
		for k, v := range m.M {
			w.WriteName(k)
			w.WriteByte(',')
			w.WriteArg(v)
		}
		w.TrimLast(1)
		w.WriteRawString(`) = (`)
		w.WriteMarkers(len(m.M))
		w.WriteRawString(`)`)
	}
	return nil
}

func WriteCols(w SQLWriter, prefix string, cols string) {
	idx := 0
	w.WriteRawString(prefix)
	w.WriteByte('.')
	for i := 0; i < len(cols); i++ {
		ch := cols[i]
		if ch == ',' {
			w.WriteQueryString(cols[idx : i+1])
			w.WriteRawString(prefix)
			w.WriteByte('.')
			idx = i + 1
		}
	}
	if idx < len(cols) {
		w.WriteQueryString(cols[idx:])
	}
}

// Ternary ...
func Ternary(cond bool, exp1, exp2 interface{}) interface{} {
	if cond {
		return exp1
	}
	return exp2
}

//go:nosplit
func unsafeBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
