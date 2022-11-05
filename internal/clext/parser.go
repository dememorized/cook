package clext

import (
	"fmt"
	"strings"
	"text/scanner"
)

type parser struct {
	pos    int
	curr   Token
	tokens []Token
	Error  error
}

func (p *parser) Next() Token {
	p.curr = p.Peek()
	if p.curr.Type != TokenEOF {
		p.pos++
	}
	return p.curr
}

func (p *parser) Peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{
			Type: TokenEOF,
		}
	}

	return p.tokens[p.pos]
}

func Parse(filename string, recipe []Token) (*AST, error) {
	ast := &AST{
		Filename: filename,
		Steps:    make([]Step, 0),
	}

	p := &parser{
		pos:    0,
		tokens: recipe,
		Error:  nil,
	}
	p.Next()

	step := Step{
		base: base{pos: scanner.Position{
			Filename: filename,
			Offset:   0,
			Line:     1,
			Column:   1,
		}},
	}

	for p.Error == nil && p.curr.Type != TokenEOF {
		t := p.curr
		b := base{pos: t.Position}

		switch t.Type {
		case TokenNewLine:
			t := p.Next()
			if t.Type == TokenNewLine && step.HasInstructions() {
				ast.Steps = append(ast.Steps, step)
				step = Step{base: b}
			}
		case TokenDoubleDash:
			step.Components = append(step.Components, Comment{
				base:    b,
				Comment: p.eatUntil(oneOf(TokenNewLine)),
			})
		case TokenDoubleGT:
			if t.Position.Column != 1 {
				step.Components = append(step.Components, Instruction{
					base:        b,
					Instruction: ">>",
				})
				p.skip(oneOf(TokenDoubleGT))
				continue
			}

			md := Metadata{
				base: b,
			}

			p.skip(oneOf(TokenDoubleGT, TokenWhitespace))
			md.Key = strings.TrimSpace(p.eatUntil(oneOf(TokenColon, TokenNewLine)))

			if !p.skip(oneOf(TokenColon)) {
				p.Error = fmt.Errorf("expected colon to separate metadata key and value on line %d", t.Position.Line)
			}
			p.skip(oneOf(TokenWhitespace))
			md.Value = p.eatUntil(oneOf(TokenNewLine))

			step.Components = append(step.Components, md)
		case TokenText, TokenWhitespace, TokenRightBrace, TokenLeftBrace:
			txt := p.eatUntil(notIn(
				TokenText,
				TokenWhitespace,
				TokenPercent,
				TokenLeftBrace,
				TokenRightBrace,
				TokenColon,
				TokenDoubleGT,
			))

			step.Components = append(step.Components, Instruction{
				base:        b,
				Instruction: txt,
			})
		case TokenTilde:
			timer := Timer{base: b}

			p.skip(oneOf(TokenTilde))
			timer.Name = p.eatUntil(oneOf(TokenLeftBrace))
			p.skip(oneOf(TokenLeftBrace))
			timer.Magnitude = p.eatUntil(oneOf(TokenPercent))
			p.skip(oneOf(TokenPercent))
			timer.Unit = p.eatUntil(oneOf(TokenRightBrace))
			p.skip(oneOf(TokenRightBrace))

			step.Components = append(step.Components, timer)
		case TokenHash:
			p.skip(oneOf(TokenHash))
			cookware := Cookware{base: b}

			if p.seekTerminal(oneOf(TokenLeftBrace), oneOf(TokenText, TokenWhitespace)) {
				cookware.Name = strings.TrimSpace(p.eatUntil(oneOf(TokenLeftBrace)))
				p.skip(oneOf(TokenLeftBrace, TokenRightBrace))
			} else {
				cookware.Name = p.eatUntil(notIn(TokenText))
			}

			step.Components = append(step.Components, cookware)
		case TokenAt:
			p.skip(oneOf(TokenAt))
			ing := Ingredient{base: b}

			if p.seekTerminal(oneOf(TokenLeftBrace), oneOf(TokenText, TokenWhitespace)) {
				ing.Name = p.eatUntil(oneOf(TokenLeftBrace))
				p.skip(oneOf(TokenLeftBrace))
				ing.Quantity = strings.TrimSpace(p.eatUntil(oneOf(TokenPercent, TokenRightBrace)))
				if p.curr.Type == TokenPercent {
					p.skip(oneOf(TokenPercent))
					ing.Unit = strings.TrimSpace(p.eatUntil(oneOf(TokenRightBrace)))
				}
				p.skip(oneOf(TokenRightBrace))
			} else {
				ing.Name = p.eatUntil(notIn(TokenText))
			}

			step.Components = append(step.Components, ing)
		default:
			p.Error = fmt.Errorf(
				"%s: got unknown token %s with value %x",
				t.Position,
				t.Type.String(),
				t.Value,
			)
			p.Next()
		}
	}

	if len(step.Components) != 0 {
		ast.Steps = append(ast.Steps, step)
	}

	return ast, p.Error
}

type condition = func(TokenType) bool

func (p *parser) skip(cond condition) bool {
	skipped := false
	for tok := p.curr; tok.Type != TokenEOF; tok = p.Next() {
		if !cond(tok.Type) {
			break
		}
		skipped = true
	}
	return skipped
}

func (p *parser) eatUntil(end condition) string {
	tokens := p.rawEatUntil(end)
	b := strings.Builder{}

	for _, tok := range tokens {
		b.WriteString(tok.Value)
	}
	return b.String()
}

func (p *parser) rawEatUntil(end condition) []Token {
	tokens := []Token{}

	for tok := p.curr; tok.Type != TokenEOF; tok = p.Next() {
		if end(tok.Type) {
			break
		}

		if tok.Type == TokenInvalid || tok.Type == TokenUnknown {
			p.Error = fmt.Errorf("found invalid token on position: %s", tok.Position)
			return nil
		}

		tokens = append(tokens, tok)
	}

	return tokens
}

// seekTerminal allows the parser to look ahead to determine whether
// a certain type condition will be met while only another condition
// is applied.
//
// We need this because an ingredient can be defined as both
//
//	Something @ingredient something else.
//
// and
//
//	Something @multi word ingredient{} something else.
func (p *parser) seekTerminal(terminal condition, allowedInter condition) bool {
	for pos := p.pos; pos < len(p.tokens); pos++ {
		tokenType := p.tokens[pos].Type
		if terminal(tokenType) {
			return true
		}
		if !allowedInter(tokenType) {
			return false
		}
	}
	return false
}

func oneOf(typs ...TokenType) condition {
	return func(tokenType TokenType) bool {
		for _, typ := range typs {
			if typ == tokenType {
				return true
			}
		}
		return false
	}
}

func notIn(typs ...TokenType) condition {
	return func(tokenType TokenType) bool {
		return !oneOf(typs...)(tokenType)
	}
}
