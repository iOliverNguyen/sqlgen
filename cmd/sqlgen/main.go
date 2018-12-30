// Experiment
//
// go run ./cmd/sqlgen/*.go ./examples/sample

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/ng-vu/sqlgen/gen/dsl"
	"github.com/ng-vu/sqlgen/gen/gocmt"
	gen "github.com/ng-vu/sqlgen/gen/sqlgen"
	"github.com/ng-vu/sqlgen/gen/strs"
)

var (
	flFile       = flag.String("f", "", "Parse from file definition (file)")
	flSkipSource = flag.Bool("s", false, "Do not parse from source comment (skip-source)")
	flPrint      = flag.Bool("p", false, "Print parsed declarations to stdout and exit")
	flOutFile    = flag.String("o", "", "Write to file instead of stdout")

	packages []string
)

func init() {
	const usage = `Usage:
  sqlgen [-f file] [-p] [-s] [packages]

Example:
  sqlgen github.com/ng-vu/sqlgen/examples/sample
  sqlgen -f definition.sqlgen package1 package2

Or use with "go:generate"
  //go:generate sqlgen

Flags:
`
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprint(w, usage)
		flag.PrintDefaults()
	}
}

func main() {
	parseFlags()
	for _, pkg := range packages {
		err := parsePackage(pkg)
		must(err)
	}
}

func parseFlags() {
	flag.Parse()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	wdir, err := os.Getwd()
	must(err)

	switch {
	case len(flag.Args()) > 0:
		for _, arg := range flag.Args() {
			arg = strings.TrimSpace(arg)
			if arg == "" {
				continue
			}
			switch arg[0] {
			case '/':
			case '.':
				arg = filepath.Join(wdir, arg)
			default:
				arg = filepath.Join(gopath, "src", arg)
			}
			packages = append(packages, arg)
		}
		if len(packages) == 0 {
			fmt.Fprint(os.Stderr, "No package provided")
			os.Exit(1)
		}
	case os.Getenv("GOPACKAGE") != "":
		packages = []string{wdir}
	default:
		flag.Usage()
		os.Exit(255)
	}
}

type Decl struct {
	Decl *dsl.Declaration
	Type types.Type
}

func parsePackage(pkgpath string) error {
	decl, err := gocmt.ParseDir(pkgpath)
	must(err)

	var b strings.Builder
	for _, group := range decl.Block {
		for _, line := range group {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	src := b.String()
	fileDecl, err := dsl.ParseString("unknown", src)
	if err != nil {
		return err
	}

	for _, decl := range fileDecl.Declarations {
		if err = linkDeclaration(decl, nil); err != nil {
			return err
		}
	}

	typeDecls := make([]*dsl.Declaration, len(decl.Types))
	for i, t := range decl.Types {
		b.Reset()
		for _, line := range t.Comment {
			b.WriteString(line)
		}
		s := b.String()
		name := t.Type.Name.Name
		typeDecl, err := dsl.ParseString("type "+name, s)
		if err != nil {
			return fmt.Errorf("Parse error on type %v: %v", name, err)
		}

		switch len(typeDecl.Declarations) {
		case 0:
			return fmt.Errorf("Empty declarations on type %v", name)
		case 1:
			typeDecls[i] = typeDecl.Declarations[0]
			if err = linkDeclaration(typeDecls[i], t.Type); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Multiple declarations on type %v", name)
		}
	}

	if *flPrint {
		fmt.Println(fileDecl)
		for _, decl := range typeDecls {
			fmt.Println(decl)
		}
		return nil
	}

	var allDecls []*Decl
	mapDecl := make(map[string]*Decl)
	for _, decl := range fileDecl.Declarations {
		d, err := addToMap(mapDecl, decl)
		if err != nil {
			return err
		}
		allDecls = append(allDecls, d)
	}
	for _, decl := range typeDecls {
		d, err := addToMap(mapDecl, decl)
		if err != nil {
			return err
		}
		allDecls = append(allDecls, d)
	}

	info := types.Info{
		// TODO
		Defs: make(map[*ast.Ident]types.Object),
	}
	files := make([]*ast.File, 0, len(decl.Package.Files))
	for _, file := range decl.Package.Files {
		files = append(files, file)
	}

	conf := types.Config{
		Importer: importer.Default(),

		IgnoreFuncBodies:         true,
		DisableUnusedImportCheck: true,
	}
	tpkg, err := conf.Check(pkgpath, decl.FileSet, files, &info)
	if err != nil {
		return err
	}

	for ident, d := range info.Defs {
		decl := mapDecl[ident.Name]
		if decl != nil {
			decl.Type = d.Type()
		}
	}
	for name, decl := range mapDecl {
		if decl.Type == nil {
			return fmt.Errorf("Error: type %v not found", name)
		}
	}

	adapter := NewAdapter()
	adapter.pkg = tpkg
	g := gen.New(adapter, nil)
	for _, decl := range allDecls {
		err = g.Add(decl.Decl.StructName, []types.Type{decl.Type})
		if err != nil {
			return err
		}
	}
	g.GenerateCommon()
	for _, decl := range allDecls {
		err = g.GenQueryFor(decl.Type)
		if err != nil {
			return err
		}
	}

	if *flOutFile != "" {
		file, err := os.OpenFile(*flOutFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		adapter.WriteTo(file)
		return nil
	}
	adapter.WriteTo(os.Stdout)
	return nil
}

func linkDeclaration(decl *dsl.Declaration, typ *ast.TypeSpec) error {
	if err := decl.ParseOptions(); err != nil {
		return fmt.Errorf("Error: %v on declaration:\n\n%v", err, decl)
	}
	if decl.StructName == "" {
		if typ == nil {
			return fmt.Errorf("Error: no struct name on declaration:\n\n%v", decl)
		}
		decl.StructName = typ.Name.Name
	}
	if decl.TableName == "" {
		decl.TableName = strs.ToSnake(decl.StructName)
	}
	if decl.OptPlural == "" {
		decl.OptPlural = strs.ToPlural(decl.StructName)
	}
	return nil
}

func addToMap(m map[string]*Decl, decl *dsl.Declaration) (*Decl, error) {
	if decl.StructName == "" {
		return nil, fmt.Errorf("No struct name on declaration\n\n%v", decl)
	}
	if _, ok := m[decl.StructName]; ok {
		return nil, fmt.Errorf("Duplicated declaration for type %v", decl.StructName)
	}
	d := &Decl{
		Decl: decl,
	}
	m[decl.StructName] = d
	return d, nil
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
