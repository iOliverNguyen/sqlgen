%{

package dsl

%}

%union {
    dc   DeclCommon
    dec  *Declaration
    decs Declarations
    jn   *Join
    jns  Joins
    opt  *Option
    opts Options
    str  string
}

%type  <dc>   from join_opts
%type  <dec>  decl
%type  <decs> decls
%type  <jn>   join
%type  <jns>  joins from_joins
%type  <opt>  opt
%type  <opts> list_opts opts
%type  <str>  name _as _ident _on

%token ';' '(' ')'

%token GENERATE FROM AS JOIN ON

%token <str> IDENT STRING JCOND

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
    GENERATE _ident list_opts from_joins
    {
        $$ = &Declaration{
            Options: $3,
        }
        switch len($4) {
        case 0:
        case 1:
            $$.DeclCommon = $4[0].DeclCommon
        default:
            $$.Joins = $4
        }
        if $2 != "" {
            $$.StructName = $2
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

from_joins:
    /* empty */
    {
        $$ = nil
    }
|   from
    {
        $$ = []*Join{{DeclCommon: $1}}
    }
|   from joins
    {
        $$ = append([]*Join{{DeclCommon: $1}}, $2...)
    }

from:
    FROM name join_opts _as
    {
        $$ = DeclCommon{
            TableName: $2,
            StructName: $3.StructName,
            Alias: $4,
        }
    }

_as:
    /* empty */
    {
        $$ = ""
    }
|   name
|   AS name
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

joins:
    join
    {
        $$ = []*Join{$1}
    }
|   joins join
    {
        $$ = append($1, $2)
    }

join:
    JOIN name join_opts _as _on
    {
        $$ = &Join{
            DeclCommon: DeclCommon{
                TableName: $2,
                StructName: $3.StructName,
                Alias: $4,
            },
            OnCond: $5,
        }
    }

join_opts:
    /* empty */
    {
        $$ = DeclCommon{}
    }
|   '(' _ident ')'
    {
        $$ = DeclCommon{ StructName: $2 }
    }

_on:
    /* empty */
    {
        $$ = ""
    }
|   ON JCOND
    {
        $$ = $2
    }

%%
