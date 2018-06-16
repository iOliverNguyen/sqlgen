package gosrc

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type PackageDeclaration struct {
	Block string
	Types []TypeDeclaration
}

type TypeDeclaration struct {
	Type    *ast.TypeSpec
	Comment string
}

func ParseDir(path string) (*PackageDeclaration, error) {
	fset := token.NewFileSet()
	pkg, err := parseDir(fset, path)
	if err != nil {
		return nil, err
	}

	types, mapCmts := extractTypeSpecs(pkg)

	// Extract types and comments
	var res PackageDeclaration
	for _, typ := range types {
		if typ.Doc == nil {
			continue
		}
		var lines []string
		for _, c := range typ.Doc.List {
			lines = append(lines, trimComment(c)...)
		}
		parsed, err := ParseComment(lines)
		if err != nil {
			return nil, err
		}
		switch len(parsed) {
		case 0:
			// skip
		case 1:
			res.Types = append(res.Types, TypeDeclaration{
				Type:    typ,
				Comment: parsed[0],
			})
		default:
			return nil, fmt.Errorf("Multiple declaration on type %v", typ.Name.Name)
		}
	}

	// Extract floating comments
	var b bytes.Buffer
	for _, file := range pkg.Files {
		for _, cmt := range file.Comments {
			if mapCmts[cmt] {
				continue
			}
			var cmts []string
			for _, c := range cmt.List {
				cmts = append(cmts, trimComment(c)...)
			}
			parsed, err := ParseComment(cmts)
			if err != nil {
				return nil, err
			}
			for _, line := range parsed {
				b.WriteString(line)
				b.WriteString("\n")
			}
		}
	}

	res.Block = b.String()
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

func extractTypeSpecs(pkg *ast.Package) ([]*ast.TypeSpec, map[*ast.CommentGroup]bool) {
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

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				types = append(types, typeSpec)
				if typeSpec.Doc == nil && genDecl.Doc != nil {
					typeSpec.Doc = genDecl.Doc
				}
				if typeSpec.Doc != nil {
					cmts[typeSpec.Doc] = true
				}
			}
		}
	}
	return types, cmts
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
