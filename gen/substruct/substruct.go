package substruct

import (
	"fmt"
	"go/types"
	"regexp"
	"strings"

	ggen "github.com/ng-vu/sqlgen/gen"
	"github.com/ng-vu/sqlgen/gen/strs"
)

func New(iface ggen.Interface) *Gen {
	return &Gen{Interface: iface}
}

type Gen struct {
	ggen.Interface
}

func (g *Gen) GenFuncFor(in, out types.Type) error {
	sIn, ok := pointerToStruct(in)
	if !ok {
		return fmt.Errorf("Expected pointer to struct (got %v)", in.String())
	}
	sOut, ok := pointerToStruct(out)
	if !ok {
		return fmt.Errorf("Expected pointer to struct (got %v)", out.String())
	}

	p := g
	inStr := g.TypeString(in)
	outStr := g.TypeString(out)
	sInStr := inStr[1:]
	sOutStr := outStr[1:]

	inMap := make(map[string]types.Type)
	for i, n := 0, sIn.NumFields(); i < n; i++ {
		v := sIn.Field(i)
		inMap[v.Name()] = v.Type()
	}

	outMap := make(map[string]types.Type)
	for i, n := 0, sOut.NumFields(); i < n; i++ {
		v := sOut.Field(i)
		outMap[v.Name()] = v.Type()

		name := v.Name()
		vInType := inMap[name]
		if vInType == nil {
			return fmt.Errorf("Field (%v).%v does not exist in (%v)", outStr, name, inStr)
		}

		vStr := g.TypeString(v.Type())
		vInStr := g.TypeString(vInType)
		if vStr != vInStr {
			return fmt.Errorf(
				"Field (%v).%v has different type with (%v).%v: Expect `%v`, got `%v`",
				inStr, name, outStr, name, vInStr, vStr)
		}
	}

	// If the name does not start with "substruct", we replace the prefix.
	if g, ok := g.Interface.(ggen.Goderive); ok {
		name := g.GetFuncName(out)
		if !strings.HasPrefix(name, "substruct") {
			re := regexp.MustCompile(`^[a-z]+`)
			prefix := re.FindString(name)
			name = "substruct" + name[len(prefix):]
		}

		p.P("")
		p.P("// %v is a substruct of %v", outStr, inStr)
		p.P("func %v(_ %v, _ %v) bool { return true }", name, outStr, inStr)
	}

	p.P("")
	p.P("func %vFrom%v(ps []%v) []%v {", strs.ToPlural(capitalize(sOutStr)), strs.ToPlural(capitalize(sInStr)), inStr, outStr)
	p.In()
	p.P("ss := make([]%v, len(ps))", outStr)
	p.P("for i, p := range ps {")
	p.In()
	p.P("ss[i] = New%vFrom%v(p)", capitalize(sOutStr), capitalize(sInStr))
	p.Out()
	p.P("}")
	p.P("return ss")
	p.Out()
	p.P("}")

	p.P("")
	p.P("func %vTo%v(ss []%v) []%v {", strs.ToPlural(capitalize(sOutStr)), strs.ToPlural(capitalize(sInStr)), outStr, inStr)
	p.In()
	p.P("ps := make([]%v, len(ss))", inStr)
	p.P("for i, s := range ss {")
	p.In()
	p.P("ps[i] = s.To%v()", sInStr)
	p.Out()
	p.P("}")
	p.P("return ps")
	p.Out()
	p.P("}")

	p.P("")
	p.P("func New%vFrom%v(sp %v) %v {", capitalize(sOutStr), capitalize(sInStr), inStr, outStr)
	p.In()
	p.P("if sp == nil {")
	p.In()
	p.P(`return nil`)
	p.Out()
	p.P("}")
	p.P("s := new(%v)", sOutStr)
	p.P("s.CopyFrom(sp)")
	p.P("return s")
	p.Out()
	p.P("}")

	p.P("")
	p.P("func (s %v) To%v() %v {", outStr, capitalize(sInStr), inStr)
	p.In()
	p.P("if s == nil {")
	p.In()
	p.P(`return nil`)
	p.Out()
	p.P("}")
	p.P("sp := new(%v)", sInStr)
	p.P("s.AssignTo(sp)")
	p.P("return sp")
	p.Out()
	p.P("}")

	p.P("")
	p.P("func (s %v) CopyFrom(sp %v) {", outStr, inStr)
	p.In()
	for i, n := 0, sOut.NumFields(); i < n; i++ {
		v := sOut.Field(i)
		name := v.Name()
		p.P("s.%v = sp.%v", name, name)
	}
	p.Out()
	p.P("}")

	p.P("")
	p.P("func (s %v) AssignTo(sp %v) {", outStr, inStr)
	p.In()
	for i, n := 0, sOut.NumFields(); i < n; i++ {
		v := sOut.Field(i)
		name := v.Name()
		p.P("sp.%v = s.%v", name, name)
	}
	p.Out()
	p.P("}")

	return nil
}

func pointerToStruct(typ types.Type) (*types.Struct, bool) {
	pt, ok := typ.Underlying().(*types.Pointer)
	if !ok {
		return nil, false
	}
	st, ok := pt.Elem().Underlying().(*types.Struct)
	return st, ok
}

func capitalize(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:]
}
