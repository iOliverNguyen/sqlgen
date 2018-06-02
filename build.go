package sq

import (
	"fmt"

	"github.com/ng-vu/sqlgen/core"
)

type (
	JOIN_TYPE = core.JoinType

	TABLE string
	AS    string
	ON    string
)

const (
	JOIN         JOIN_TYPE = ""
	FULL_JOIN    JOIN_TYPE = "FULL OUTER"
	LEFT_JOIN    JOIN_TYPE = "LEFT OUTER"
	RIGHT_JOIN   JOIN_TYPE = "RIGHT OUTER"
	NATURAL_JOIN JOIN_TYPE = "NATURAL"
	CROSS_JOIN   JOIN_TYPE = "CROSS"
	SELF_JOIN    JOIN_TYPE = "SELF"
)

type Sqlizer interface {
	ToSql() (string, []interface{}, error)
}

type expr struct {
	sql  string
	args []interface{}
}

func Expr(sql string, args ...interface{}) expr {
	return expr{sql: sql, args: args}
}

func (e expr) ToSql() (sql string, args []interface{}, err error) {
	return e.sql, e.args, nil
}

type exprs []expr

func (es exprs) Append(s core.IState, b []byte, args []interface{}, sep string) ([]byte, []interface{}, error) {
	for i, e := range es {
		if i > 0 {
			b = append(b, sep...)
		}
		b = s.AppendQueryStr(b, e.sql)
		args = append(args, e.args...)
	}
	return b, args, nil
}

type part struct {
	pred interface{}
	args []interface{}
}

type parts []Sqlizer

func (ps parts) Append(s core.IState, b []byte, args []interface{}, sep string) ([]byte, []interface{}, error) {
	for i, p := range ps {
		psql, pargs, err := p.ToSql()
		if err != nil {
			return nil, nil, err
		}
		if len(psql) == 0 {
			continue
		}

		if i > 0 {
			b = append(b, sep...)
		}
		b = s.AppendQueryStr(b, psql)
		args = append(args, pargs...)
	}
	return b, args, nil
}

type wherePart part

func newWherePart(pred interface{}, args ...interface{}) Sqlizer {
	return &wherePart{pred: pred, args: args}
}

func (p wherePart) ToSql() (sql string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case string:
		sql = pred
		args = p.args
	case []byte:
		sql = string(pred)
		args = p.args
	case Sqlizer:
		return pred.ToSql()
	case nil:
		// no-op
	default:
		err = fmt.Errorf("expected string-keyed map or string, not %T", pred)
	}
	return
}
