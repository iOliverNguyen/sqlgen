//go:generate goyacc -v= y.y

package dsl

import (
	"bytes"
	"fmt"
	"strings"
	"text/scanner"
)

// Variable to store the parsed result
var result *File

type File struct {
	Declarations Declarations
}

func (f *File) String() string {
	return f.Declarations.String()
}

type Declarations []*Declaration

func (ds Declarations) String() string {
	var buf bytes.Buffer
	for _, d := range ds {
		fmt.Fprintf(&buf, "%v;\n", d)
	}
	return buf.String()
}

type Declaration struct {
	TableName  string
	StructName string
	Alias      string
	Options    Options
}

func (d *Declaration) String() string {
	var buf bytes.Buffer
	buf.WriteString("generate ")
	buf.WriteString(d.StructName)
	if len(d.Options) > 0 {
		buf.WriteString(" (")
		buf.WriteString(d.Options.String())
		buf.WriteString(")")
	}
	buf.WriteString(` from "`)
	buf.WriteString(d.TableName)
	buf.WriteString(`"`)
	if d.Alias != "" {
		buf.WriteString(` as "`)
		buf.WriteString(d.Alias)
		buf.WriteString(`"`)
	}
	return buf.String()
}

type Options []*Option

func (opts Options) String() string {
	if len(opts) == 0 {
		return ""
	}

	var buf bytes.Buffer
	for i, opt := range opts {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(opt.String())
	}
	return buf.String()
}

type Option struct {
	Name  string
	Value string
}

func (opt Option) String() string {
	return opt.Name + " " + opt.Value
}

type lexer struct {
	scanner.Scanner
	err error
}

func (l *lexer) Lex(yylval *yySymType) int {
	if tok := l.Scan(); tok == scanner.EOF {
		return 0
	}

	text := l.TokenText()
	switch text {
	case "":
		return 0
	case ";", "(", ")":
		return int(text[0])
	case "generate":
		return GENERATE
	case "from":
		return FROM
	case "as":
		return AS
	default:
		if text[0] == '"' || text[0] == '`' {
			yylval.str = text[1 : len(text)-1]
			return STRING
		}
		yylval.str = text
		return IDENT
	}
}

func (l *lexer) Error(s string) {
	l.err = fmt.Errorf("Error at %v: %v", l.Position, s)
}

func ParseString(filename, src string) (*File, error) {
	l := &lexer{}
	l.Init(strings.NewReader(src))
	l.Filename = filename
	if yyParse(l) != 0 {
		return nil, l.err
	}
	return result, nil
}
