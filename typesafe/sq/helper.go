package sq

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ng-vu/sqlgen/core"
)

type (
	JOIN_TYPE     = core.JoinType
	ABSTRACT_TYPE string

	AS    string
	ON    string
	TABLE string
)

const (
	JOIN         JOIN_TYPE = ""
	FULL_JOIN    JOIN_TYPE = "FULL"
	LEFT_JOIN    JOIN_TYPE = "LEFT"
	RIGHT_JOIN   JOIN_TYPE = "RIGHT"
	NATURAL_JOIN JOIN_TYPE = "NATURAL"
	CROSS_JOIN   JOIN_TYPE = "CROSS"
	SELF_JOIN    JOIN_TYPE = "SELF"
)

type (
	SEL int
	INS int
	UPD int
	DEL int
)

func NewExpr(preds ...interface{}) WriterTo {
	if len(preds) == 0 {
		panic("sqlgen: no param")
	}
	args := preds[1:]
	switch pred := preds[0].(type) {
	case string:
		return ExprString{pred, args}
	case []byte:
		return Expr{pred, args}
	case WriterTo:
		if len(preds) == 1 {
			return pred
		}
		msg := fmt.Sprintf("sqlgen: NewExpr only accepts single WriterTo")
		panic(msg)
	default:
		msg := fmt.Sprintf("sqlgen: unsupported type (got %T)", pred)
		panic(msg)
	}
}

type Expr struct {
	SQL  []byte
	Args []interface{}
}

var _ WriterTo = Expr{}

func (e Expr) WriteSQLTo(w core.SQLWriter) error {
	w.WriteQuery(e.SQL)
	w.WriteArgs(e.Args)
	return nil
}

type ExprString struct {
	SQL  string
	Args []interface{}
}

var _ WriterTo = ExprString{}

func (e ExprString) WriteSQLTo(w core.SQLWriter) error {
	w.WriteQueryString(e.SQL)
	w.WriteArgs(e.Args)
	return nil
}

type Parts []WriterTo

func (ps Parts) WriteSQLTo(w core.SQLWriter, sep string) error {
	for i, p := range ps {
		if i > 0 {
			w.WriteRawString(sep)
		}
		if err := p.WriteSQLTo(w); err != nil {
			return err
		}
	}
	return nil
}

func (ps *Parts) Append(preds ...interface{}) Parts {

	if len(preds) == 0 || preds[0] == nil {
		return *ps
	}
	args := preds[1:]
	switch pred := preds[0].(type) {
	case string:
		*ps = append(*ps, ExprString{pred, args})
	case []byte:
		*ps = append(*ps, Expr{pred, args})
	case WriterTo:
		*ps = append(*ps, pred)
		for _, arg := range args {
			if p, ok := arg.(WriterTo); ok {
				*ps = append(*ps, p)
			} else {
				msg := fmt.Sprintf("sqlgen: all args must implement WriterTo interface (got %T)", arg)
				panic(msg)
			}
		}
	default:
		msg := fmt.Sprintf("sqlgen: unsupported type (got %T)", pred)
		panic(msg)
	}
	return *ps
}

type InPart struct {
	in     bool
	column string
	args   []interface{} // len(args) must be greater than 0
}

func In(column string, args ...interface{}) WriterTo {
	return NewInPart(true, column, args...)
}

func NotIn(column string, args ...interface{}) WriterTo {
	return NewInPart(false, column, args...)
}

func NewInPart(in bool, column string, args ...interface{}) WriterTo {
	switch len(args) {
	case 0:
		return NewExpr("FALSE")
	case 1:
		if t := reflect.TypeOf(args[0]); t.Kind() == reflect.Slice {
			vArgs := reflect.ValueOf(args[0])
			if vArgs.Len() == 0 {
				return NewExpr("FALSE")
			}
			args = make([]interface{}, vArgs.Len())
			for i, n := 0, vArgs.Len(); i < n; i++ {
				args[i] = vArgs.Index(i).Interface()
			}
		}
	default:
		// no-op
	}
	return InPart{
		in:     in,
		column: column,
		args:   args,
	}
}

func (p InPart) WriteSQLTo(w core.SQLWriter) error {
	if len(p.args) == 0 {
		return fmt.Errorf("sqlgen: unexpected len(args)")
	}

	w.WriteQueryName(p.column)
	if !p.in {
		w.WriteRawString(" NOT")
	}
	w.WriteRawString(" IN (")
	w.WriteMarkers(len(p.args))
	w.WriteByte(')')
	w.WriteArgs(p.args)
	return nil
}

// In with multple columns
type InsPart struct {
	in      bool
	columns []string

	// len(args) must be greater than 0 and multiple of len(columns)
	args []interface{}
}

func Ins(columns []string, args ...interface{}) WriterTo {
	return NewInsPart(true, columns, args...)
}

func NotIns(columns []string, args ...interface{}) WriterTo {
	return NewInsPart(false, columns, args...)
}

func NewInsPart(in bool, columns []string, args ...interface{}) WriterTo {
	if len(columns) == 0 {
		panic("columns can not be empty")
	}

	switch len(args) {
	case 0:
		return NewExpr("FALSE")
	case 1:
		if t := reflect.TypeOf(args[0]); t.Kind() == reflect.Slice {
			vArgs := reflect.ValueOf(args[0])
			if vArgs.Len() == 0 {
				return NewExpr("FALSE")
			}
			if vArgs.Len()%len(columns) != 0 {
				panic("args.Len() must be multiple of len(columns)")
			}
			args = make([]interface{}, vArgs.Len())
			for i, n := 0, vArgs.Len(); i < n; i++ {
				args[i] = vArgs.Index(i).Interface()
			}
		}
	default:
		if len(args)%len(columns) != 0 {
			panic("len(args) must be multiple of len(columns)")
		}
	}
	return InsPart{
		in:      in,
		columns: columns,
		args:    args,
	}
}

func (p InsPart) WriteSQLTo(w core.SQLWriter) error {
	if len(p.args) == 0 {
		return fmt.Errorf("sqlgen: unexpected len(args)")
	}
	c := len(p.columns)
	n := len(p.args) / c

	w.WriteByte('(')
	for _, col := range p.columns {
		w.WriteQueryName(col)
		w.WriteByte(',')
	}
	w.TrimLast(1)
	w.WriteByte(')')
	if !p.in {
		w.WriteRawString(" NOT")
	}
	w.WriteRawString(" IN (")
	for i := 0; i < n; i++ {
		w.WriteByte('(')
		w.WriteMarkers(c)
		w.WriteRawString("),")
	}
	w.TrimLast(1)
	w.WriteByte(')')
	w.WriteArgs(p.args)
	return nil
}

type IsNullPart struct {
	column string
	null   bool
}

func NewIsNullPart(column string, null bool) WriterTo {
	return IsNullPart{column, null}
}

func (p IsNullPart) WriteSQLTo(w core.SQLWriter) error {
	w.WriteQueryName(p.column)
	w.WriteRawString(" IS")
	if !p.null {
		w.WriteRawString(" NOT")
	}
	w.WriteRawString(" NULL")
	return nil
}

type And []WriterTo

func (ps *And) Append(preds ...interface{}) And {
	return And((*Parts)(ps).Append(preds...))
}

func (ps And) WriteSQLTo(w core.SQLWriter) error {
	w.WriteRawString("(")
	err := Parts(ps).WriteSQLTo(w, ") AND (")
	w.WriteRawString(")")
	return err
}

type Or []WriterTo

func (ps *Or) Append(preds ...interface{}) Or {
	return Or((*Parts)(ps).Append(preds...))
}

func (ps Or) WriteSQLTo(w core.SQLWriter) error {
	w.WriteRawString("(")
	err := Parts(ps).WriteSQLTo(w, ") OR (")
	w.WriteRawString(")")
	return err
}

type Once []WriterTo

func (ps Once) WriteSQLTo(w core.SQLWriter) error {
	count := 0
	length := w.Len()
	for _, p := range ps {
		if err := p.WriteSQLTo(w); err != nil {
			return err
		}
		if w.Len() != length {
			count++
		}
	}
	if count != 1 {
		return errors.New("must provide exactly one argument")
	}
	return nil
}

type preloadPart struct {
	table string
	preds []interface{}
}

/*
ColumnFilter and ColumnFilterPtr are shorthand for quickly writting predication.
Sample code:

    db.Where(
        FilterByID(1234),
        FilterByState(state).Optional(),
        FilterByPartnerID(5678).Optional(),
    )

	func FilterByID(id int64) *ColumnFilter {
		return &ColumnFilter{
			Column: "id",
			Value:  id,
			IsZero: id == 0,
		}
	}

	func FilterByID(id *int64) *ColumnFilterPtr {
		return &ColumnFilterPtr{
			Column: "id",
			Value:  id,
			IsNil:  id == nil,
			IsZero: id == nil && *id == 0,
		}
	}

Explaination of different modes:

	func FilterByID(id int64)

	| mode      | id == 0           | id != 0 | notes                                  |
    |-----------|-------------------|---------|----------------------------------------|
	| (default) | (error)           | id = ?  | if id == 0, returns error              |
	| Optional  | (skip)            | id = ?  | if if == 0, skips                      |
	| Nullable  | IS NULL OR id = 0 | id = ?  | if id == 0, translates to "id IS NULL" |

	func FilterByPtrID(id *int64)

	| mode         | id == nil | *id == 0          | *id != 0 | explain                                                                             |
    |--------------|-----------|-------------------|----------|-------------------------------------------------------------------------------------|
	| (default)    | (error)   | IS NULL OR id = 0 | id = ?   | if id == nil, returns error; else if *id == 0, translates to "id IS NULL OR id = 0" |
	| Optional     | (skip)    | IS NULL OR id = 0 | id = ?   | if id == nil, skips; else if *id == 0, translates to "id IS NULL OR if = 0"         |
	| Nullable     | IS NULL   | id = 0            | id = ?   | if id == nil, translates to "id IS NULL"; else if *id == 0, translates to "id = 0"  |
    | RequiredZero | (error)   | id == 0           | id = ?   |                                                                                     |
    | RequiredNull | (error)   | IS NULL           | id = ?   |                                                                                     |
    | OptionalZero | (skip)    | id == 0           | id = ?   |                                                                                     |
    | OptionalNull | (skip)    | IS NULL           | id = ?   |                                                                                     |
*/
type ColumnFilter ColumnFilterPtr

func (p *ColumnFilter) Optional() *ColumnFilter {
	p.mode = ModeOptional
	return p
}

func (p *ColumnFilter) Nullable() *ColumnFilter {
	p.mode = ModeNullable
	return p
}

func (p *ColumnFilter) WriteSQLTo(w core.SQLWriter) error {
	return (*ColumnFilterPtr)(p).WriteSQLTo(w)
}

type ColumnFilterPtr struct {
	Prefix string
	Column string
	Value  interface{}
	IsNil  bool
	IsZero bool

	op
	pred string
	mode FilterMode
}

func (p *ColumnFilterPtr) Optional() *ColumnFilterPtr {
	p.mode = ModeOptional
	return p
}

func (p *ColumnFilterPtr) Nullable() *ColumnFilterPtr {
	p.mode = ModeNullable
	return p
}

func (p *ColumnFilterPtr) RequiredZero() *ColumnFilterPtr {
	p.mode = ModeRequiredZero
	return p
}

func (p *ColumnFilterPtr) RequiredNull() *ColumnFilterPtr {
	p.mode = ModeRequiredNull
	return p
}

func (p *ColumnFilterPtr) OptionalZero() *ColumnFilterPtr {
	p.mode = ModeOptionalZero
	return p
}

func (p *ColumnFilterPtr) OptionalNull() *ColumnFilterPtr {
	p.mode = ModeOptionalNull
	return p
}

func (p *ColumnFilterPtr) WriteSQLTo(w core.SQLWriter) error {
	if p.IsNil {
		switch {
		case p.mode == ModeNullable:
			w.WritePrefixedName(p.Prefix, p.Column)
			if p.op == "" {
				w.WriteRawString(" IS NULL")
			} else {
				w.WriteByte(' ')
				w.WriteRawString(string(p.op))
				w.WriteByte(' ')
				w.WriteMarker()
			}
			return nil

		case p.mode&ModeOptional != 0: // optional
			return nil

		default: // required
			return core.InvalidArgumentError("missing " + p.Column)
		}
	}
	if p.IsZero {
		switch {
		case p.mode&2 != 0: // equal to null
			w.WritePrefixedName(p.Prefix, p.Column)
			if p.op == "" {
				w.WriteRawString(" IS NULL")
			} else {
				w.WriteByte(' ')
				w.WriteRawString(string(p.op))
				w.WriteByte(' ')
				w.WriteMarker()
			}
			return nil

		case p.mode&1 != 0: // equal to zero
			w.WritePrefixedName(p.Prefix, p.Column)
			if p.op == "" {
				w.WriteRawString(" = ")
			} else {
				w.WriteByte(' ')
				w.WriteRawString(string(p.op))
				w.WriteByte(' ')
			}
			w.WriteMarker()
			w.WriteArg(p.Value)
			return nil

		default: // is null or zero
			w.WritePrefixedName(p.Prefix, p.Column)
			if p.op == "" {
				w.WriteRawString(" IS NULL OR ")
				w.WritePrefixedName(p.Prefix, p.Column)
				w.WriteRawString(" = ")
				w.WriteMarker()
			} else {
				w.WriteByte(' ')
				w.WriteRawString(string(p.op))
				w.WriteByte(' ')
				w.WriteMarker()
			}
			w.WriteArg(p.Value)
			return nil
		}
	}

	w.WritePrefixedName(p.Prefix, p.Column)
	w.WriteRawString(" = ")
	w.WriteMarker()
	w.WriteArg(p.Value)
	return nil
}

// Bit layout
//
//     ______ÑO ______NZ
//
//     Z: Equal to Zero (id = 0)
//     N: Equal to NULL (id IS NULL)
//     O: Optional, won't throw error if the field is empty
//     Ñ: Nullable, the column value can be null
type FilterMode int

const (
	ModeDefault      FilterMode = iota
	ModeOptional     FilterMode = 1<<8 + 0 // _O & __
	ModeNullable     FilterMode = 2<<8 + 1 // Ñ_ & _Z
	ModeRequiredZero FilterMode = 0<<8 + 1 // __ & _Z
	ModeRequiredNull FilterMode = 0<<8 + 2 // __ & N_
	ModeOptionalZero FilterMode = 1<<8 + 1 // _O & _Z
	ModeOptionalNull FilterMode = 1<<8 + 2 // _O & N_
)

type op string

func (op *op) Gt() {
	*op = ">"
}

func (op *op) Lt() {
	*op = "<"
}

func (op *op) Gte() {
	*op = ">="
}

func (op *op) Lte() {
	*op = "<="
}

func Filter(prefix, pred string, args ...interface{}) WriterTo {
	return WriterToFunc(func(w SQLWriter) error {
		w.WriteQueryStringWithPrefix(prefix, pred)
		w.WriteArgs(args)
		return nil
	})
}
