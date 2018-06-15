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

%type  <dec>  decl
%type  <decs> decls
%type  <opt>  opt
%type  <opts> list_opts opts
%type  <str>  name _as as

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
    GENERATE IDENT list_opts FROM name _as
    {
        $$ = &Declaration{
            StructName: $2,
            Options: $3,
            TableName: $5,
            Alias: $6,
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
    opt
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

_as:
    /* empty */
    {
        $$ = ""
    }
|   as
    {
        $$ = $1
    }

as:
    AS name
    {
        $$ = $2
    }

name: IDENT | STRING

%%
