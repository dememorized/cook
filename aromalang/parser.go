package aromalang

import (
	"reflect"
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
	}

	var tokens []Token
	for _, tok := range recipe {
		switch tok.Type {
		case TokenWhitespace, TokenNewLine:
			break
		default:
			tokens = append(tokens, tok)
		}
	}

	p := &parser{
		pos:    0,
		tokens: tokens,
		Error:  nil,
	}
	p.Next()

	for p.Error == nil && p.curr.Type != TokenEOF {
		switch p.curr.Type {
		case TokenLParen:
			comp, err := p.parseElement()
			if err != nil {
				return nil, err
			}

			recipe, ok := comp.(Recipe)
			if !ok {
				return nil, NewErrorf(comp.Position(), "expected Recipe, got type '%s'", reflect.TypeOf(comp))
			}

			ast.Recipe = recipe
			err = checkType(TokenRParen, p.Next())
			if err != nil {
				return nil, err
			}
			p.Next()
		default:
			return nil, NewErrorf(p.curr.Position, "unexpected token %s", p.curr.Type)
		}
	}

	if p.Error != nil {
		return nil, p.Error
	}

	return ast, nil
}

func (p *parser) parseElement() (Component, error) {
	if p.Peek().Type != TokenIdentifier {
		return nil, NewErrorf(p.Peek().Position, "expected TokenIdentifier, got %s", p.Peek().Type)
	}

	switch tok := p.Next(); tok.Value {
	case "recipe":
		r := Recipe{
			Base: Base{Pos: tok.Position},
		}

		metadata, err := p.parseMap()
		if err != nil {
			return nil, err
		}
		r.Metadata = metadata

		steps, err := p.parseListOfSteps()
		if err != nil {
			return nil, err
		}
		r.Steps = steps

		return r, nil
	case "step":
		return Step{}, nil
	default:
		return nil, NewErrorf(tok.Position, "unknowns identifier '%s'", tok.Value)
	}
}

var mapKeyTypes = map[TokenType]struct{}{
	TokenString:     {},
	TokenAtom:       {},
	TokenIdentifier: {},
}

func (p *parser) parseMap() ([]Metadata, error) {
	startPos := p.Peek().Position
	if p.Next().Type != TokenLBrace {
		return nil, NewErrorf(p.curr.Position, "expected TokenLBrace, got %s", p.curr.Type)
	}

	var elems []Metadata

	for p.Peek().Type != TokenEOF && p.Peek().Type != TokenRBrace {
		key := p.Next()
		if _, exists := mapKeyTypes[key.Type]; !exists {
			return nil, NewErrorf(p.curr.Position, "metadata key must be a string, identifier, or atom, got %s", key.Type)
		}

		value := p.Next()
		// TODO: Change this, values should be able to be of many types, including dynamic types.
		if value.Type != TokenString {
			return nil, NewErrorf(p.curr.Position, "metadata values must be a string, got %s", key.Type)
		}

		elems = append(elems, Metadata{
			Base:  Base{Pos: key.Position},
			Key:   key.Value,
			Value: value.Value,
		})
	}
	if p.Next().Type != TokenRBrace {
		return nil, NewErrorf(startPos, "unclosed map")
	}

	return elems, nil
}

func (p *parser) parseListOfSteps() ([]Step, error) {
	var steps []Step
	startPos := p.Peek().Position
	if p.Next().Type != TokenLBracket {
		return nil, NewErrorf(p.curr.Position, "expected TokenLBracket, got %s", p.curr.Type)
	}

	for p.Peek().Type != TokenRBracket && p.Peek().Type != TokenEOF {
		if p.Next().Type != TokenLParen {
			return nil, NewErrorf(p.curr.Position, "expected TokenLParen, got %s", p.curr.Type)
		}

		s, err := p.parseElement()
		if err != nil {
			return nil, err
		}

		step, ok := s.(Step)
		if !ok {
			return nil, NewErrorf(s.Position(), "expected Step, got type '%s'", reflect.TypeOf(s))
		}

		md, err := p.parseMap()
		if err != nil {
			return nil, err
		}
		for _, m := range md {
			step.Components = append(step.Components, m)
		}

		comps, err := p.parseListOfComps()
		if err != nil {
			return nil, err
		}

		step.Components = append(step.Components, comps...)

		err = checkType(TokenRParen, p.Next())
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	if p.Next().Type != TokenRBracket {
		return nil, NewErrorf(startPos, "unclosed list")
	}

	return steps, nil
}

func (p *parser) parseListOfComps() ([]Component, error) {
	var comps []Component
	startPos := p.Peek().Position
	if p.Next().Type != TokenLBracket {
		return nil, NewErrorf(startPos, "expected TokenLBracket, got %s", p.curr.Type)
	}

	for p.Peek().Type != TokenRBracket && p.Peek().Type != TokenEOF {
		c, err := p.parseComponent()
		if err != nil {
			return nil, err
		}
		comps = append(comps, c)
	}

	if p.Next().Type != TokenRBracket {
		return nil, NewErrorf(startPos, "unclosed list")
	}

	return comps, nil
}

func (p *parser) parseComponent() (Component, error) {
	var comp Component

	if err := checkType(TokenLParen, p.Next()); err != nil {
		return nil, err
	}

	c := p.Next()
	if err := checkType(TokenIdentifier, c); err != nil {
		return nil, err
	}

	switch c.Value {
	case "instruction":
		tok := p.Next()
		if err := checkType(TokenString, tok); err != nil {
			return nil, err
		}
		comp = Instruction{
			Base:        Base{Pos: c.Position},
			Instruction: tok.Value,
		}
	case "cookware":
		tok := p.Next()
		if err := checkType(TokenString, tok); err != nil {
			return nil, err
		}
		comp = Cookware{
			Base: Base{Pos: c.Position},
			Name: tok.Value,
		}
	case "ingredient":
		ing := Ingredient{
			Base: Base{Pos: c.Position},
		}
		tokName := p.Next()
		if err := checkType(TokenString, tokName); err != nil {
			return nil, err
		}
		ing.Name = tokName.Value

		md, err := p.parseMap()
		if err != nil {
			return nil, err
		}
		for _, m := range md {
			switch m.Key {
			case "quantity":
				ing.Quantity = m.Value
			case "unit":
				ing.Unit = m.Value
			default:
				return nil, NewErrorf(c.Position, "unknown key in map. expected :quantity or :unit, got %s", m.Key)
			}
		}

		comp = ing
	case "timer":
		timer := Timer{
			Base: Base{Pos: c.Position},
		}
		tokName := p.Next()
		if err := checkType(TokenString, tokName); err == nil {
			timer.Name = tokName.Value
		}

		md, err := p.parseMap()
		if err != nil {
			return nil, err
		}
		for _, m := range md {
			switch m.Key {
			case "magnitude":
				timer.Magnitude = m.Value
			case "unit":
				timer.Unit = m.Value
			default:
				return nil, NewErrorf(c.Position, "unknown key in map. expected :magnitude or :unit, got %s", m.Key)
			}
		}

		comp = timer
	default:
		return nil, NewErrorf(c.Position, "unknown instruction '%s'", c.Value)
	}

	if err := checkType(TokenRParen, p.Next()); err != nil {
		return nil, err
	}

	return comp, nil
}

func checkType(expected TokenType, token Token) error {
	if token.Type != expected {
		return NewErrorf(token.Position, "expected token %s, got token %s", expected, token.Type)
	}
	return nil
}
