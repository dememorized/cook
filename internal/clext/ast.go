package clext

import (
	"fmt"
	"strings"
	"text/scanner"
)

type AST struct {
	Filename string
	Steps    []Step
	Errors   []ParseError
}

func (a *AST) Metadata() map[string]string {
	md := map[string]string{}
	for _, step := range a.Steps {
		for _, m := range step.Metadata() {
			md[m.Key] = m.Value
		}
	}
	return md
}

func (a *AST) ReportError(s *scanner.Scanner, msg string) {
	a.Errors = append(a.Errors, ParseError{
		Position: s.Position,
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
	return fmt.Sprintf("(step %s)\n", b.String())
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

func (s Step) Metadata() []Metadata {
	list := []Metadata{}
	for _, c := range s.Components {
		i, ok := c.(Metadata)
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
	return fmt.Sprintf(`(instruction "%s")`, i.Instruction)
}

type Comment struct {
	base
	Comment string
}

func (c Comment) String() string {
	return fmt.Sprintf(`(comment "%s")`, c.Comment)
}

type Ingredient struct {
	base
	Name     string
	Quantity string
	Unit     string
}

func (i Ingredient) String() string {
	return fmt.Sprintf(`(ingredient "%s" "%s" "%s")`, i.Name, i.Quantity, i.Unit)
}

type Cookware struct {
	base
	Name string
}

func (c Cookware) String() string {
	return fmt.Sprintf(`(cookware "%s")`, c.Name)
}

type Timer struct {
	base
	Name      string
	Magnitude string
	Unit      string
}

func (t Timer) String() string {
	return fmt.Sprintf(`(timer "%s" "%s")`, t.Magnitude, t.Unit)
}

type Metadata struct {
	base
	Key   string
	Value string
}

func (m Metadata) String() string {
	return fmt.Sprintf(`(metadata "%s" "%s")`, m.Key, m.Value)
}
