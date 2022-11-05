package clext

import (
	"io"
	"strings"
	"text/scanner"
	"unicode"
)

type Token struct {
	Type     TokenType
	Position scanner.Position
	Value    string
}

type TokenType uint16

const (
	TokenUnknown TokenType = iota
	TokenInvalid
	TokenEOF
	TokenText
	TokenWhitespace
	TokenNewLine

	TokenPercent
	TokenAt
	TokenDoubleDash
	TokenLeftBrace
	TokenRightBrace
	TokenHash
	TokenTilde
)

func (t TokenType) String() string {
	switch t {
	case TokenInvalid:
		return "Invalid"
	case TokenEOF:
		return "EOF"
	case TokenText:
		return "Text"
	case TokenWhitespace:
		return "Whitespace"
	case TokenNewLine:
		return "NewLine"
	case TokenPercent:
		return "Percent"
	case TokenAt:
		return "At"
	case TokenDoubleDash:
		return "DoubleDash"
	case TokenLeftBrace:
		return "LeftBrace"
	case TokenRightBrace:
		return "RightBrace"
	case TokenTilde:
		return "Tilde"
	default:
		return "Unknown"
	}
}

func Tokenize(filename string, recipe io.Reader) ([]Token, []ParseError) {
	errors := []ParseError{}

	scan := &scanner.Scanner{}
	scan.Init(recipe)
	scan.Filename = filename
	scan.Error = func(s *scanner.Scanner, msg string) {
		errors = append(errors, ParseError{
			Position: s.Pos(),
			Message:  msg,
		})
	}

	tokens := []Token{}

	for {
		c := scan.Next()
		if c == scanner.EOF {
			break
		}

		tok, ok := singleCharControl(c, scan)
		if ok {
			tokens = append(tokens, tok)
			continue
		}

		tok, ok = doubleCharControl(c, scan)
		if ok {
			tokens = append(tokens, tok)
			continue
		}

		tokens = append(tokens, eatWord(c, scan))
	}

	return tokens, errors
}

func singleCharControl(c rune, scan *scanner.Scanner) (Token, bool) {
	t := Token{
		Position: scan.Pos(),
		Value:    string(c),
	}

	switch c {
	case '@':
		t.Type = TokenAt
	case '#':
		t.Type = TokenHash
	case '{':
		t.Type = TokenLeftBrace
	case '}':
		t.Type = TokenRightBrace
	case '%':
		t.Type = TokenPercent
	case '\r':
		if scan.Peek() != '\n' {
			return Token{}, false
		}
		scan.Next()
		t.Type = TokenNewLine
	case '\n':
		t.Type = TokenNewLine
		if scan.Peek() == '\r' {
			scan.Next()
		}
	case '~':
		t.Type = TokenTilde
	default:
		return Token{}, false
	}

	return t, true
}

func doubleCharControl(c rune, scan *scanner.Scanner) (Token, bool) {
	val := string(c) + string(scan.Peek())

	t := Token{
		Position: scan.Pos(),
		Value:    val,
	}

	switch val {
	case "--":
		t.Type = TokenDoubleDash
	default:
		return Token{}, false
	}

	// eat up the extra character
	scan.Next()
	return t, true
}

func eatWord(c rune, scan *scanner.Scanner) Token {
	t := Token{
		Position: scan.Pos(),
	}

	t.Type = TokenInvalid
	t.Value = string(c)

	if unicode.In(c, unicode.White_Space) {
		b := strings.Builder{}
		b.WriteRune(c)

		for unicode.In(scan.Peek(), unicode.White_Space) {
			b.WriteRune(scan.Next())
		}

		t.Type = TokenWhitespace
		t.Value = b.String()
	}

	if unicode.IsOneOf(unicode.PrintRanges, c) {
		b := strings.Builder{}
		b.WriteRune(c)

		terminal := &unicode.RangeTable{
			R16: []unicode.Range16{
				{Lo: '%', Hi: '%', Stride: 1}, // 0x25
				{Lo: '{', Hi: '}', Stride: 2}, // 0x7b, 0x7d
			},
			LatinOffset: 2,
		}
		for unicode.IsOneOf(unicode.PrintRanges, scan.Peek()) && !unicode.In(scan.Peek(), terminal) {
			b.WriteRune(scan.Next())
		}

		t.Type = TokenText
		t.Value = b.String()
	}

	return t
}
