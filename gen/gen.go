package gface

import (
	"go/types"
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
