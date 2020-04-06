package tableParser

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Pos int

type token struct {
	typ  tokenType // the type of this token
	pos  Pos       // the starting position, in bytes, of this token in the input string
	val  string    // the value of this token
	line int       // the line number at the start of this token
}

func (i token) String() string {
	switch {
	case i.typ == tokenEOF:
		return "EOF"
	case i.typ == tokenError:
		return i.val
	case i.typ > tokenKeyword:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.15q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

type tokenType int

var runeType = map[rune]tokenType{
	'.': tokenDot,
	',': tokenComma,
	'(': tokenLeftParen,
	')': tokenRightParen,
	';': tokenSemicolon,
}

var keywords = map[string]tokenType{
	"create":  tokenCreate,
	"table":   tokenTable,
	"if":      tokenIF,
	"not":     tokenNOT,
	"exists":  tokenEXISTS,
	"null":    tokenNULL,
	"default": tokenDEFAULT,
}

const eof = -1

type TableDefine struct {
	Schema  string
	Table   string
	Columns []*TableColumn
}

type TableColumn struct {
	Name string
	Type string
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

type lexer struct {
	name      string
	input     string // the string being scanned
	pos       Pos    //current position in the input
	start     Pos    // start position of this token
	width     Pos    // width of last rune read from input
	tokens    chan token
	line      int          // line number of newlines
	startLine int          // start line of this token
	lerror    error        // last error
	lwidth    Pos          // last pos rune width before eof
	inEof     bool         // whether input is in eof status
	ast       *TableDefine // the final result ast tree
}

func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.pos = Pos(len(l.input))
		l.lwidth = l.width
		l.width = 0
		l.inEof = true
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	if l.inEof {
		l.pos = Pos(len(l.input))
		l.width = l.lwidth
		l.inEof = false
	} else {
		l.pos -= l.width
		if l.width == 1 && l.input[l.pos] == '\n' {
			l.line--
		}
	}
}

// emit passes an item back to the client.
func (l *lexer) emit(t tokenType) {
	new_t := token{t, l.start, l.input[l.start:l.pos], l.startLine}
	switch t {
	case tokenPgSymbol:
		new_t.val = strings.ReplaceAll(new_t.val, "\"\"", "\"")
	case tokenPgValue:
		new_t.val = strings.ReplaceAll(new_t.val, "''", "'")
	}
	l.tokens <- new_t
	l.start = l.pos
	l.startLine = l.line
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	//l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
	l.startLine = l.line
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

func (l *lexer) acceptLetter() {
	var r rune
	for {
		r = l.next()
		if !isLetter(r) {
			break
		}
	}
	if r != eof {
		l.backup()
	}
}

func (l *lexer) acceptDigit() {
	var r rune
	for {
		r = l.next()
		if !isDigit(r) {
			break
		}
	}
	if r != eof {
		l.backup()
	}
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.tokens <- token{tokenError, l.start, fmt.Sprintf(format, args...), l.startLine}
	return nil
}

// nextToken returns the next item from the input.
// Called by the parser, not in the lexing goroutine.
func (l *lexer) nextToken() token {
	return <-l.tokens
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for state := lexText; state != nil; {
		//fmt.Printf("%s\n", runtime.FuncForPC(reflect.ValueOf(state).Pointer()).Name())

		state = state(l)
	}
	close(l.tokens)
}

func (l *lexer) scanWord() {
	l.acceptLetter()
	for {
		r := l.peek()
		if isDigit(r) {
			l.acceptDigit()
		} else if isLetter(r) {
			l.acceptLetter()
		} else {
			break
		}
	}
}

func (l *lexer) skipBlank() {
	var r rune
	for {
		r = l.next()
		if !(isSpace(r) || isEndOfLine(r)) {
			l.backup()
			break
		}
	}
	l.ignore()
}

func (l *lexer) Lex(lval *yySymType) int {
	token := l.nextToken()
	switch token.typ {
	case tokenError:
		fallthrough
	case tokenEOF:
		return 0
	}
	lval.val = token.val
	return int(token.typ)
}

func (l *lexer) Error(s string) {
	columnPos := 1
	i := int(l.pos)
	if i >= len(l.input) {
		i = len(l.input) - 1
	}
	for i >= 0 {
		if l.input[i] == '\n' {
			break
		}
		columnPos++
		i--
	}
	l.lerror = fmt.Errorf("%s near line:%d column:%d", s, l.startLine, columnPos)
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:      name,
		input:     input,
		tokens:    make(chan token),
		line:      1,
		startLine: 1,
		ast:       new(TableDefine),
	}
	go l.run()
	return l
}

func lexText(l *lexer) stateFn {
	for {
		l.skipBlank()
		r := l.next()
		if r == eof {
			l.emit(tokenEOF)
			return nil
		}
		if t, found := runeType[r]; found {
			l.emit(t)
		} else {
			switch true {
			case r == '\'':
				l.ignore() // ignore the '
				return lexPgValue
			case r == '"': // ignore the "
				l.ignore()
				return lexPgSymbol
			case isLetter(r):
				l.backup()
				return lexIdentifier
			default:
				l.emit(tokenUnknown)
			}
		}
	}
}

func lexIdentifier(l *lexer) stateFn {
	l.scanWord()
	lowerWord := strings.ToLower(l.input[l.start:l.pos])
	if keyWordType, found := keywords[lowerWord]; found {
		l.emit(keyWordType)
	} else {
		l.emit(tokenString)
	}
	return lexText
}
func lexPgSymbol(l *lexer) stateFn {
	for {
		for {
			r := l.next()
			if r == '"' {
				break
			} else if r == eof {
				l.emit(tokenPgSymbol)
				return nil
			}
		}
		// is end or escape?
		r := l.next()
		if r == eof {
			l.backup()
			l.backup()
			l.emit(tokenPgSymbol)
			l.emit(tokenEOF)
			return nil
		} else if r != '"' {
			l.backup()
			l.backup()
			l.emit(tokenPgSymbol)
			l.next()
			l.ignore()
			break
		}
	}
	return lexText
}

func lexPgValue(l *lexer) stateFn {
	for {
		for {
			r := l.next()
			if r == '\'' {
				break
			} else if r == eof {
				l.emit(tokenEOF)
				return nil
			}
		}
		// is end or escape?
		r := l.next()
		if r == eof {
			l.backup()
			l.backup()
			l.emit(tokenPgValue)
			l.emit(tokenEOF)
			return nil
		} else if r != '\'' {
			l.backup()
			l.backup()
			l.emit(tokenPgValue)
			l.next()
			l.ignore()
			break
		}
	}
	return lexText
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isLetter(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || r == '_'
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}
