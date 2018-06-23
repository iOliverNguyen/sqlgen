package gosrc

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type PackageDeclaration struct {
	Block [][]string
	Types []TypeDeclaration
}

type TypeDeclaration struct {
	Type    *ast.TypeSpec
	Comment []string
}

func ParseDir(path string) (*PackageDeclaration, error) {
	fset := token.NewFileSet()
	pkg, err := parseDir(fset, path)
	if err != nil {
		return nil, err
	}

	types, mapCmts, err := extractTypeSpecs(pkg)
	if err != nil {
		return nil, err
	}

	// Extract types and comments
	var res PackageDeclaration
	for _, typ := range types {
		groups, err := parseCommandGroup(typ.Doc)
		if err != nil {
			return nil, err
		}
		switch len(groups) {
		case 0:
			// continue
		case 1:
			res.Types = append(res.Types, TypeDeclaration{
				Type:    typ,
				Comment: groups[0],
			})
		default:
			return nil, errors.New("Multiple declarations on type " + typ.Name.Name)
		}
	}

	// Extract floating comments
	var g [][]string
	for _, file := range pkg.Files {
		for _, cmt := range file.Comments {
			if mapCmts[cmt] {
				continue
			}
			groups, err := parseCommandGroup(cmt)
			if err != nil {
				return nil, err
			}
			for _, group := range groups {
				g = append(g, group)
			}
		}
	}

	res.Block = g
	return &res, nil
}

func parseDir(fset *token.FileSet, path string) (*ast.Package, error) {
	pkgs, err := parser.ParseDir(fset, path, fileFilter, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("sqlgen: Error while parsing package %v: %v", path, err)
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("sqlgen: Unexpected error while parsing package %v", path)
	}
	for _, p := range pkgs {
		return p, nil
	}
	panic("unreachable")
}

func extractTypeSpecs(pkg *ast.Package) ([]*ast.TypeSpec, map[*ast.CommentGroup]bool, error) {
	cmts := make(map[*ast.CommentGroup]bool)
	var types []*ast.TypeSpec
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok != token.TYPE {
				continue
			}

			count := 0
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				count++
				types = append(types, typeSpec)
				if typeSpec.Doc == nil && genDecl.Doc != nil {
					typeSpec.Doc = genDecl.Doc
				}
				if typeSpec.Doc != nil {
					cmts[typeSpec.Doc] = true
				}
			}
			if count > 1 {
				parsed, err := parseCommandGroup(genDecl.Doc)
				if err != nil || len(parsed) > 0 {
					var names []string
					for _, spec := range types {
						names = append(names, spec.Name.Name)
					}
					l := len(names) - 1
					return nil, nil, errors.New("Must not mix declaration on type " + strings.Join(names[:l], ",") + " and " + names[l])
				}
			}
		}
	}
	return types, cmts, nil
}

func parseCommandGroup(cg *ast.CommentGroup) ([][]string, error) {
	if cg == nil {
		return nil, nil
	}
	var lines []string
	for _, c := range cg.List {
		lines = append(lines, trimComment(c)...)
	}
	return ParseComment(lines)
}

func fileFilter(file os.FileInfo) bool {
	name := file.Name()
	return !strings.HasPrefix(name, "_") && !strings.HasSuffix(name, "_test.go")
}

func trimComment(c *ast.Comment) []string {
	txt := c.Text
	switch txt[1] {
	case '/':
		return []string{txt[2:]}
	case '*':
		return strings.Split(txt[2:len(txt)-2], "\n")
	}
	panic("unexpected")
}
