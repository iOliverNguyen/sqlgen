package substruct

import (
	"fmt"
	"go/types"

	"github.com/awalterschulze/goderive/derive"
	ggen "github.com/ng-vu/sqlgen/gen"
	sstruct "github.com/ng-vu/sqlgen/gen/substruct"
)

// NewPlugin creates a new clone plugin.
// This function returns the plugin name, default prefix and a constructor for the clone code generator.
func NewPlugin() derive.Plugin {
	return derive.NewPlugin("substruct", "substruct", New)
}

// New is a constructor for the clone code generator.
// This generator should be reconstructed for each package.
func New(typesMap derive.TypesMap, p derive.Printer, deps map[string]derive.Dependency) derive.Generator {
	adapter := ggen.NewGoderiveAdapter(typesMap, p)
	return &gen{
		TypesMap: typesMap,
		Printer:  p,
		g:        sstruct.New(adapter),
	}
}

type gen struct {
	derive.TypesMap
	derive.Printer
	in types.Type
	g  *sstruct.Gen
}

func (g *gen) Add(name string, typs []types.Type) (string, error) {
	if len(typs) != 2 {
		return "", fmt.Errorf("%s must have at least two arguments", name)
	}
	g.in = typs[1]
	return g.SetFuncName(name, typs[0])
}

func (g *gen) Generate(typs []types.Type) error {
	return g.g.GenFuncFor(g.in, typs[0])
}

func (g *gen) GenSubstruct(typ, base types.Type) error {
	panic("Unexpected")
}
