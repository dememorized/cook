package aromalang

import (
	"fmt"
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

type TokenType int16

const (
	TokenUnknown TokenType = iota
	TokenInvalid
	TokenEOF
	TokenNewLine
	TokenWhitespace

	TokenLBracket
	TokenRBracket
	TokenLBrace
	TokenRBrace
	TokenLParen
	TokenRParen

	TokenIdentifier
	TokenString
	TokenNumeral
	TokenAtom
)

func (t TokenType) String() string {
	switch t {
	case TokenUnknown:
		return "Unknown"
	case TokenInvalid:
		return "Invalid"
	case TokenEOF:
		return "EOF"
	case TokenNewLine:
		return "NewLine"
	case TokenWhitespace:
		return "Whitespace"
	case TokenLBracket:
		return "LBracket"
	case TokenRBracket:
		return "RBracket"
	case TokenLBrace:
		return "LBrace"
	case TokenRBrace:
		return "RBrace"
	case TokenLParen:
		return "LParen"
	case TokenRParen:
		return "RParen"
	case TokenIdentifier:
		return "Identifier"
	case TokenString:
		return "String"
	case TokenNumeral:
		return "Numeral"
	case TokenAtom:
		return "Atom"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

func Tokenize(filename string, recipe io.Reader) ([]Token, []*ParseError) {
	errors := []*ParseError{}

	scan := &scanner.Scanner{}
	scan.Init(recipe)
	scan.Filename = filename
	scan.Error = func(s *scanner.Scanner, msg string) {
		errors = append(errors, &ParseError{
			Position: s.Pos(),
			Message:  msg,
		})
	}

	tokens := []Token{}

	for {
		t, err := nextToken(scan)
		if t.Type == TokenEOF {
			break
		}
		if err != nil {
			errors = append(errors, err)
			break
		}

		tokens = append(tokens, t)
	}

	return tokens, errors
}

func nextToken(scan *scanner.Scanner) (tok Token, err *ParseError) {
	position := scan.Pos()
	defer func() {
		tok.Position = position
	}()

	c := scan.Next()
	if c == scanner.EOF {
		return Token{
			Type:  TokenEOF,
			Value: "\u0000",
		}, nil
	}

	if tok, ok := singleCharControl(c, scan); ok {
		return tok, nil
	}

	return eatWord(c, scan)
}

func singleCharControl(c rune, scan *scanner.Scanner) (Token, bool) {
	t := Token{
		Value: string(c),
	}

	switch c {
	case '(':
		t.Type = TokenLParen
	case ')':
		t.Type = TokenRParen
	case '{':
		t.Type = TokenLBrace
	case '}':
		t.Type = TokenRBrace
	case '[':
		t.Type = TokenLBracket
	case ']':
		t.Type = TokenRBracket
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
	default:
		return Token{}, false
	}

	return t, true
}

func eatWord(c rune, scan *scanner.Scanner) (Token, *ParseError) {
	t := Token{
		Type:  TokenInvalid,
		Value: string(c),
	}

	b := strings.Builder{}
	if unicode.In(c, unicode.White_Space) {
		b.WriteRune(c)

		for unicode.In(scan.Peek(), unicode.White_Space) {
			b.WriteRune(scan.Next())
		}

		t.Type = TokenWhitespace
		t.Value = b.String()
	} else if c == '"' {
		if !unicode.In(scan.Peek(), unicode.PrintRanges...) && !unicode.In(scan.Peek(), unicode.White_Space) {
			return Token{}, &ParseError{
				Position: scan.Pos(),
				Message:  fmt.Sprintf("non-printable unicode character in string %+q (0x%x)", scan.Peek(), scan.Peek()),
			}
		}

		var escaped bool
		for c := scan.Next(); c != scanner.EOF && !(c == '"' && !escaped); c = scan.Next() {
			if escaped {
				switch c {
				case '\\', '"':
					break
				default:
					return Token{}, &ParseError{
						Position: scan.Pos(),
						Message:  fmt.Sprintf("trying to escape '%c', but only \\ and \" can be escaped. Use \\\\ for a literal \\", c),
					}
				}
				escaped = false
			} else if c == '\\' {
				escaped = true
				continue
			}

			b.WriteRune(c)
		}

		t.Type = TokenString
		t.Value = b.String()
	} else if c == ':' {
		for unicode.In(scan.Peek(), identifierStart, decimalNumbers) {
			b.WriteRune(scan.Next())
		}

		t.Type = TokenAtom
		t.Value = b.String()
	} else if unicode.In(c, identifierStart) {
		b.WriteRune(c)

		for unicode.In(scan.Peek(), identifierStart, decimalNumbers) {
			b.WriteRune(scan.Next())
		}

		t.Type = TokenIdentifier
		t.Value = b.String()
	} else if unicode.In(c, decimalNumbers, numeralControlChars) {
		b.WriteRune(c)

		for unicode.In(scan.Peek(), decimalNumbers, numeralControlChars) {
			b.WriteRune(scan.Next())
		}

		t.Type = TokenNumeral
		t.Value = b.String()
	}

	return t, nil
}

var identifierStart = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x2e, Hi: 0x2f, Stride: 1}, // . /
		{Lo: 0x41, Hi: 0x5a, Stride: 1}, // A-Z
		{Lo: 0x5f, Hi: 0x5f, Stride: 1}, // _
		{Lo: 0x61, Hi: 0x7a, Stride: 1}, // a-z
	},
	LatinOffset: 4,
}

var decimalNumbers = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x30, Hi: 0x39, Stride: 1}, // 0-9
	},
	LatinOffset: 1,
}

var numeralControlChars = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x2b, Hi: 0x2d, Stride: 1}, // + , -
		{Lo: 0x5f, Hi: 0x5f, Stride: 1}, // _
	},
	LatinOffset: 1,
}
