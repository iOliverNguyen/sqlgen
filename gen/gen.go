package gface

import (
	"go/types"

	"github.com/awalterschulze/goderive/derive"
)

type Goderive interface {
	GetFuncName(typs ...types.Type) string
}

type Interface interface {
	P(format string, a ...interface{})
	In()
	Out()
	NewImport(name, path string) func() string
	TypeString(types.Type) string
}

type Definition struct {
	StType string
	Table  string
	Plural string
	Alias  string
	Joins  []*JoinDef

	All, Select, Insert, Update, Delete bool
}

type JoinDef struct {
	JnType string
	StType string
	Table  string
	Alias  string
	OnCond string
}

type GoderiveAdapter struct {
	derive.TypesMap
	derive.Printer
}

func NewGoderiveAdapter(tm derive.TypesMap, p derive.Printer) Interface {
	return &GoderiveAdapter{tm, p}
}

func (g *GoderiveAdapter) NewImport(name, path string) func() string {
	return g.Printer.NewImport(name, path)
}
