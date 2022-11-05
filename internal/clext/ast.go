package clext

import (
	"strings"
	"text/scanner"
)

type AST struct {
	Filename string
	Steps    []Step
	Errors   []ParseError
}

func (a *AST) ReportError(s *scanner.Scanner, msg string) {
	a.Errors = append(a.Errors, ParseError{
		Position: s.Pos(),
		Message:  msg,
	})
}

type ParseError struct {
	Position scanner.Position
	Message  string
}

func (e ParseError) Error() string {
	return e.Message
}

type Component interface {
	isComponent()
	Position() scanner.Position
	String() string
}

type base struct {
	pos scanner.Position
}

func (base) isComponent() {}

func (b base) Position() scanner.Position {
	return b.pos
}

type Step struct {
	base
	Components []Component
}

func (s Step) String() string {
	b := strings.Builder{}
	for _, step := range s.Components {
		switch step.(type) {
		case Comment:
			continue
		default:
			b.WriteString(step.String())
		}
	}
	return b.String() + "\n\n"
}

func (s Step) HasInstructions() bool {
	for _, c := range s.Components {
		switch c.(type) {
		case Instruction, Ingredient:
			return true
		}
	}

	return false
}

func (s Step) Ingredients() []Ingredient {
	list := []Ingredient{}
	for _, c := range s.Components {
		i, ok := c.(Ingredient)
		if ok {
			list = append(list, i)
		}
	}
	return list
}

type Instruction struct {
	base
	Instruction string
}

func (i Instruction) String() string {
	return i.Instruction
}

type Comment struct {
	base
	Comment string
}

func (c Comment) String() string {
	return c.Comment
}

type Ingredient struct {
	base
	Ingredient string
	Quantity   string
	Unit       string
}

func (i Ingredient) String() string {
	b := strings.Builder{}

	if i.Quantity != "" {
		b.WriteString(i.Quantity + " ")
	}
	if i.Unit != "" {
		b.WriteString(i.Unit + " ")
	}
	b.WriteString(i.Ingredient)

	return b.String()
}

type Cookware struct {
	base
	Name string
}

func (c Cookware) String() string {
	return c.Name
}

type Timer struct {
	base
	Name      string
	Magnitude string
	Unit      string
}

func (t Timer) String() string {
	return t.Magnitude + " " + t.Unit
}
