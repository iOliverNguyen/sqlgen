//go:generate goyacc -v=__y.output -o=y.go y.y
//go:generate goimports -w y.go

package dsl

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/scanner"
)

// Variable to store the parsed result
var result *RootDeclaration

type RootDeclaration struct {
	Declarations Declarations
}

func (f *RootDeclaration) String() string {
	return f.Declarations.String()
}

type Declarations []*Declaration

func (ds Declarations) String() string {
	var buf strings.Builder
	for _, d := range ds {
		fmt.Fprintf(&buf, "%v;\n", d)
	}
	return buf.String()
}

type DeclCommon struct {
	StructName string
	SchemaName string
	TableName  string
	Alias      string

	OptPlural string
}

func (d DeclCommon) TableFullName() string {
	if d.SchemaName != "" {
		return d.SchemaName + ".`" + d.TableName + "`"
	}
	return `"` + d.TableName + `"`
}

type Declaration struct {
	DeclCommon
	Options Options
	Joins   Joins
}

func (d *Declaration) String() string {
	var buf strings.Builder
	buf.WriteString("generate ")
	writeIdent(&buf, d.StructName)
	if len(d.Options) > 0 {
		buf.WriteString(" (")
		buf.WriteString(d.Options.String())
		buf.WriteString(")")
	}
	if len(d.Joins) == 0 {
		buf.WriteString(" from ")
		writeTableName(&buf, d.SchemaName, d.TableName)
	} else {
		buf.WriteString("\n    from ")
		buf.WriteString(d.Joins.String())
	}
	if d.Alias != "" {
		buf.WriteString(` as "`)
		buf.WriteString(d.Alias)
		buf.WriteString(`"`)
	}
	return buf.String()
}

func (d *Declaration) ParseOptions() error {
	for _, opt := range d.Options {
		switch opt.Name {
		case "plural":
			if d.OptPlural != "" {
				return errors.New("Option `plural` already defined")
			}
			d.OptPlural = opt.Value
		default:
			return fmt.Errorf("Unknown option `%v`", opt.Name)
		}
	}
	return nil
}

type Joins []*Join

func (jns Joins) String() string {
	var buf strings.Builder
	for i, jn := range jns {
		if i > 0 {
			buf.WriteString("\n    ")
			if jn.JoinType != "" {
				buf.WriteString(strings.ToLower(jn.JoinType))
				buf.WriteString(" ")
			}
			buf.WriteString("join ")
		}
		buf.WriteString(jn.String())
	}
	return buf.String()
}

type Join struct {
	DeclCommon
	JoinType string
	OnCond   string
}

func (jn *Join) String() string {
	var buf strings.Builder
	writeTableName(&buf, jn.SchemaName, jn.TableName)
	if jn.StructName != "" {
		buf.WriteString(" (")
		buf.WriteString(jn.StructName)
		buf.WriteString(")")
	}
	if jn.Alias != "" {
		buf.WriteString(" as ")
		buf.WriteString(jn.Alias)
	}
	if jn.OnCond != "" {
		buf.WriteString(" on ")
		buf.WriteString(jn.OnCond)
	}
	return buf.String()
}

type Options []*Option

func (opts Options) String() string {
	if len(opts) == 0 {
		return ""
	}

	var buf strings.Builder
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

func writeTableName(w *strings.Builder, schema, table string) {
	if schema != "" {
		b, _ := json.Marshal(schema)
		w.Write(b)
		w.WriteString(".")
	}
	if table == "" {
		w.WriteString(`"{}"`)
	} else {
		b, _ := json.Marshal(table)
		w.Write(b)
	}
}

func writeIdent(w *strings.Builder, s string) {
	if s == "" {
		w.WriteString(`{}`)
	} else {
		w.WriteString(s)
	}
}

func quoteName(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

type lexer struct {
	scanner.Scanner
	src  string
	last int
	next string
	on   bool
	err  error
}

func isKeyword(s string) bool {
	switch s {
	case "generate", "join", "left", "right", "full", "from", "inner", "as", "on":
		return true
	default:
		return false
	}
}

func (l *lexer) Lex(yylval *yySymType) (tok int) {
	defer func() { l.last = tok }()

	var text string
	if l.next != "" {
		text = l.next
		l.next = ""
	} else {
		if tok := l.Scan(); tok == scanner.EOF {
			return 0
		}
		text = l.TokenText()
	}

	// lex JCOND
	if l.on {
		l.on = false
		if text[0] == '`' {
			yylval.str = text[1 : len(text)-1]
			return JCOND
		}

		start := l.Position.Offset
		for text != ";" && !isKeyword(text) {
			if tok := l.Scan(); tok == scanner.EOF {
				text = ""
				break
			}
			text = l.TokenText()
		}
		end := l.Position.Offset
		cond := strings.TrimSpace(l.src[start:end])
		if cond != "" {
			l.next = text
			yylval.str = cond
			return JCOND
		}
		// else continue
	}

	switch text {
	case "":
		return 0
	case ".", ";", "(", ")":
		return int(text[0])
	case "generate":
		if l.last != 0 && l.last != ';' {
			l.next = text
			return ';'
		}
		return GENERATE
	case "from":
		return FROM
	case "as":
		return AS
	case "full":
		return FULL
	case "left":
		return LEFT
	case "right":
		return RIGHT
	case "inner":
		return INNER
	case "join":
		return JOIN
	case "on":
		l.on = true
		return ON
	default:
		if text[0] == '`' {
			yylval.str = text[1 : len(text)-1]
			return STRING
		}
		if text[0] == '"' {
			var v string
			if err := json.Unmarshal([]byte(text), &v); err != nil {
				return 0
			}
			yylval.str = v
			return STRING
		}
		yylval.str = text
		return IDENT
	}
}

func (l *lexer) Error(s string) {
	l.err = fmt.Errorf("Error at %v: %v", l.Position, s)
}

func ParseString(filename, src string) (*RootDeclaration, error) {
	l := &lexer{src: src}
	l.Init(strings.NewReader(src))
	l.Filename = filename
	if yyParse(l) != 0 {
		return nil, l.err
	}
	return result, nil
}
