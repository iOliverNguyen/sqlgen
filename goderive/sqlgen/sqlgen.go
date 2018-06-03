package sqlgen

import (
	"go/types"

	"github.com/awalterschulze/goderive/derive"
	ggen "github.com/ng-vu/sqlgen/gen"
	gen "github.com/ng-vu/sqlgen/gen/sqlgen"
	"github.com/ng-vu/sqlgen/goderive/substruct"
)

// NewPlugin creates a new sqlgen plugin.
// This function returns the plugin name, default prefix and a constructor for the clone code generator.
func NewPlugin() derive.Plugin {
	return derive.NewPlugin("sqlgen", "sqlgen", New)
}

// New is a constructor for the clone code generator.
// This generator should be reconstructed for each package.
func New(typesMap derive.TypesMap, p derive.Printer, deps map[string]derive.Dependency) derive.Generator {
	adapter := ggen.NewGoderiveAdapter(typesMap, p)
	g := &gn{
		TypesMap: typesMap,
		Printer:  p,
	}
	g.g = gen.New(adapter, g)
	return g
}

type gn struct {
	derive.TypesMap
	derive.Printer
	g *gen.Gen
}

func (g *gn) Add(name string, typs []types.Type) (string, error) {
	g.g.Add(name, typs)
	return g.SetFuncName(name, typs[0])
}

func (g *gn) Generate(typs []types.Type) error {
	if err := g.g.ValidateTypes(); err != nil {
		return err
	}

	g.g.GenerateCommon()
	g.Generating(typs[0])
	return g.g.GenQueryFor(typs[0])
}

func (g *gn) GenSubstruct(typ, base types.Type) error {
	sgen := substruct.New(g.TypesMap, g.Printer, nil)
	if _, err := sgen.Add(g.GetFuncName(typ), []types.Type{typ, base}); err != nil {
		return err
	}
	return sgen.Generate([]types.Type{typ})
}
