package tableParser

import (
	"testing"
)

type lexTest struct {
	name   string
	input  string
	tokens []token
}

func mkToken(typ tokenType, text string) token {
	return token{
		typ: typ,
		val: text,
	}
}

var (
	vEOF    = mkToken(tokenEOF, "")
	vCREATE = mkToken(tokenCreate, "CREATE")
	vTABLE  = mkToken(tokenTable, "TABLE")
)

const (
	create_table       = `CREATE TABLE`
	scan_value         = `'{},;'`
	scan_symbol        = `"(,;"`
	scan_escape_value  = `'aabbc''dde'`
	scan_escape_symbol = `"aabbc""dde"`
)

var lexTests = []lexTest{
	{"empty", "", []token{vEOF}},
	{"create table", create_table, []token{
		vCREATE,
		vTABLE,
		vEOF,
	}},
	{"scan value", scan_value, []token{
		mkToken(tokenPgValue, "{},;"),
		vEOF,
	}},
	{"scan symbol", scan_symbol, []token{
		mkToken(tokenPgSymbol, `(,;`),
		vEOF,
	}},
	{"scan escape value", scan_escape_value, []token{
		mkToken(tokenPgValue, `aabbc'dde`),
		vEOF,
	}},
	{"scan escape symbol", scan_escape_symbol, []token{
		mkToken(tokenPgSymbol, `aabbc"dde`),
		vEOF,
	}},
	{"test last rune scan", ");", []token{
		mkToken(tokenRightParen, ")"),
		mkToken(tokenSemicolon, ";"),
		vEOF,
	}},
}

func collect(t *lexTest) (tokens []token) {
	l := lex(t.name, t.input)
	for {
		token := l.nextToken()
		tokens = append(tokens, token)
		if token.typ == tokenEOF || token.typ == tokenError {
			break
		}
	}
	return
}

func tokenEqual(i1, i2 []token, checkPos bool) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			return false
		}
		if i1[k].val != i2[k].val {
			return false
		}
		if checkPos && i1[k].pos != i2[k].pos {
			return false
		}
		if checkPos && i1[k].line != i2[k].line {
			return false
		}
	}
	return true
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		tokens := collect(&test)
		if !tokenEqual(tokens, test.tokens, false) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, tokens, test.tokens)
		}
	}
}
