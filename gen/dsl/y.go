//line y.y:2
package dsl

import __yyfmt__ "fmt"

//line y.y:3
//line y.y:7
type yySymType struct {
	yys  int
	dc   DeclCommon
	dec  *Declaration
	decs Declarations
	jn   *Join
	jns  Joins
	opt  *Option
	opts Options
	str  string
}

const GENERATE = 57346
const FROM = 57347
const AS = 57348
const JOIN = 57349
const ON = 57350
const IDENT = 57351
const STRING = 57352
const JCOND = 57353

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"'.'",
	"';'",
	"'('",
	"')'",
	"GENERATE",
	"FROM",
	"AS",
	"JOIN",
	"ON",
	"IDENT",
	"STRING",
	"JCOND",
	"','",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line y.y:215

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 47

var yyAct = [...]int{

	21, 34, 6, 15, 20, 18, 24, 29, 36, 44,
	7, 22, 23, 22, 23, 25, 16, 26, 43, 19,
	13, 4, 41, 27, 28, 30, 10, 5, 31, 32,
	35, 1, 38, 37, 35, 39, 33, 40, 3, 42,
	14, 9, 11, 17, 8, 2, 12,
}
var yyPact = [...]int{

	13, -1000, 22, -1000, -3, 13, 20, -1000, -1000, 11,
	3, -1000, 8, 0, -1, -1000, 0, 8, -1000, 0,
	19, 24, -1000, -1000, -1000, 3, -1000, -1000, 19, -2,
	-3, 0, -1000, -2, -1000, -1000, 0, 15, -1000, 6,
	-1000, -1000, -1000, -6, -1000,
}
var yyPgo = [...]int{

	0, 46, 4, 7, 38, 45, 5, 43, 42, 3,
	41, 40, 0, 1, 2, 39, 31,
}
var yyR1 = [...]int{

	0, 16, 16, 5, 5, 4, 10, 10, 11, 11,
	11, 9, 8, 8, 8, 1, 2, 2, 13, 13,
	13, 14, 14, 12, 12, 7, 7, 6, 3, 3,
	15, 15,
}
var yyR2 = [...]int{

	0, 1, 2, 1, 3, 4, 0, 3, 0, 1,
	3, 2, 0, 1, 2, 4, 1, 3, 0, 1,
	2, 0, 1, 1, 1, 1, 2, 5, 0, 3,
	0, 2,
}
var yyChk = [...]int{

	-1000, -16, -5, -4, 8, 5, -14, 13, -4, -10,
	6, -8, -1, 9, -11, -9, 13, -7, -6, 11,
	-2, -12, 13, 14, 7, 16, -12, -6, -2, -3,
	6, 4, -9, -3, -13, -12, 10, -14, -12, -13,
	-12, 7, -15, 12, 15,
}
var yyDef = [...]int{

	0, -2, 1, 3, 21, 2, 6, 22, 4, 12,
	8, 5, 13, 0, 0, 9, 0, 14, 25, 0,
	28, 16, 23, 24, 7, 0, 11, 26, 28, 18,
	21, 0, 10, 18, 15, 19, 0, 0, 17, 30,
	20, 29, 27, 0, 31,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	6, 7, 3, 3, 16, 3, 4, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 5,
}
var yyTok2 = [...]int{

	2, 3, 8, 9, 10, 11, 12, 13, 14, 15,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:37
		{
			result = &File{
				Declarations: yyDollar[1].decs,
			}
		}
	case 2:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:43
		{
			result = &File{
				Declarations: yyDollar[1].decs,
			}
		}
	case 3:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:51
		{
			yyVAL.decs = []*Declaration{yyDollar[1].dec}
		}
	case 4:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:55
		{
			yyVAL.decs = append(yyDollar[1].decs, yyDollar[3].dec)
		}
	case 5:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line y.y:62
		{
			yyVAL.dec = &Declaration{
				Options: yyDollar[3].opts,
			}
			switch len(yyDollar[4].jns) {
			case 0:
			case 1:
				yyVAL.dec.DeclCommon = yyDollar[4].jns[0].DeclCommon
			default:
				yyVAL.dec.Joins = yyDollar[4].jns
			}
			if yyDollar[2].str != "" {
				yyVAL.dec.StructName = yyDollar[2].str
			}
		}
	case 6:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line y.y:80
		{
			yyVAL.opts = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:84
		{
			yyVAL.opts = yyDollar[2].opts
		}
	case 8:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line y.y:90
		{
			yyVAL.opts = nil
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:94
		{
			yyVAL.opts = []*Option{yyDollar[1].opt}
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:98
		{
			yyVAL.opts = append(yyDollar[1].opts, yyDollar[3].opt)
		}
	case 11:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:104
		{
			yyVAL.opt = &Option{
				Name:  yyDollar[1].str,
				Value: yyDollar[2].str,
			}
		}
	case 12:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line y.y:113
		{
			yyVAL.jns = nil
		}
	case 13:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:117
		{
			yyVAL.jns = []*Join{{DeclCommon: yyDollar[1].dc}}
		}
	case 14:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:121
		{
			yyVAL.jns = append([]*Join{{DeclCommon: yyDollar[1].dc}}, yyDollar[2].jns...)
		}
	case 15:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line y.y:127
		{
			yyVAL.dc = DeclCommon{
				SchemaName: yyDollar[2].dc.SchemaName,
				TableName:  yyDollar[2].dc.TableName,
				StructName: yyDollar[3].dc.StructName,
				Alias:      yyDollar[4].str,
			}
		}
	case 16:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:138
		{
			yyVAL.dc = DeclCommon{
				TableName: yyDollar[1].str,
			}
		}
	case 17:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:144
		{
			yyVAL.dc = DeclCommon{
				SchemaName: yyDollar[1].str,
				TableName:  yyDollar[3].str,
			}
		}
	case 18:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line y.y:153
		{
			yyVAL.str = ""
		}
	case 20:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:158
		{
			yyVAL.str = yyDollar[2].str
		}
	case 21:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line y.y:164
		{
			yyVAL.str = ""
		}
	case 25:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line y.y:173
		{
			yyVAL.jns = []*Join{yyDollar[1].jn}
		}
	case 26:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:177
		{
			yyVAL.jns = append(yyDollar[1].jns, yyDollar[2].jn)
		}
	case 27:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line y.y:183
		{
			yyVAL.jn = &Join{
				DeclCommon: DeclCommon{
					SchemaName: yyDollar[2].dc.SchemaName,
					TableName:  yyDollar[2].dc.TableName,
					StructName: yyDollar[3].dc.StructName,
					Alias:      yyDollar[4].str,
				},
				OnCond: yyDollar[5].str,
			}
		}
	case 28:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line y.y:197
		{
			yyVAL.dc = DeclCommon{}
		}
	case 29:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line y.y:201
		{
			yyVAL.dc = DeclCommon{StructName: yyDollar[2].str}
		}
	case 30:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line y.y:207
		{
			yyVAL.str = ""
		}
	case 31:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line y.y:211
		{
			yyVAL.str = yyDollar[2].str
		}
	}
	goto yystack /* stack new state and value */
}
