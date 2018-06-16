%{

package dsl

%}

%union {
    dec  *Declaration
    decs Declarations
    opt  *Option
    opts Options
    str  string
}

%type  <dec>  decl _from
%type  <decs> decls
%type  <opt>  opt
%type  <opts> list_opts opts
%type  <str>  name _as as _ident

%token ';' '(' ')'

%token GENERATE FROM AS

%token <str> IDENT STRING

%%

top:
    decls
    {
        result = &File{
            Declarations: $1,
        }
    }
|   decls ';'
    {
        result = &File{
            Declarations: $1,
        }
    }

decls:
    decl
    {
        $$ = []*Declaration{$1}
    }
|   decls ';' decl
    {
        $$ = append($1, $3)
    }


decl:
    GENERATE _ident list_opts _from
    {
        $$ = &Declaration{
            StructName: $2,
            Options: $3,
            TableName: $4.TableName,
            Alias: $4.Alias,
        }
    }

list_opts:
    /* empty */
    {
        $$ = nil
    }
|   '(' opts ')'
    {
        $$ = $2
    }

opts:
    /* empty */
    {
        $$ = nil
    }
|   opt
    {
        $$ = []*Option{$1}
    }
|   opts ',' opt
    {
        $$ = append($1, $3)
    }

opt:
    IDENT name
    {
        $$ = &Option{
            Name: $1,
            Value: $2,
        }
    }

_from:
    /* empty */
    {
        $$ = &Declaration{}
    }
|   FROM name _as
    {
        $$ = &Declaration{
            TableName: $2,
            Alias: $3,
        }
    }

_as:
    /* empty */
    {
        $$ = ""
    }
|   as

as:
    AS name
    {
        $$ = $2
    }

_ident:
    /* empty */
    {
        $$ = ""
    }
|   IDENT

name: IDENT | STRING

%%
