package aromalang

import (
	"fmt"
	"strconv"
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

func (a *AST) String() string {
	md := a.Metadata()
	mdString := strings.Builder{}

	if len(md) != 0 {
		mdString.WriteByte('\n')
	}
	for k, v := range md {
		mdString.WriteString(fmt.Sprintf("\t%s %s\n", strconv.Quote(k), strconv.Quote(v)))
	}

	stepString := strings.Builder{}
	for _, step := range a.Steps {
		stepString.WriteString("\n" + step.String())
	}

	return fmt.Sprintf("(recipe {%s}\n[%s])", mdString.String(), stepString.String())
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

type Base struct {
	Pos scanner.Position
}

func (Base) isComponent() {}

func (b Base) Position() scanner.Position {
	return b.Pos
}

type Step struct {
	Base
	Components []Component
}

func (s Step) String() string {
	comps := make([]string, 0, len(s.Components))
	for _, step := range s.Components {
		switch step.(type) {
		// Comments are not printed, and Metadata is top level only in Cooklang
		case Comment, Metadata:
			continue
		default:
			comps = append(comps, step.String())
		}
	}
	return fmt.Sprintf("(step {}\n\t[%s])\n", strings.Join(comps, "\n\t"))
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
	Base
	Instruction string
}

func (i Instruction) String() string {
	return fmt.Sprintf(`(instruction "%s")`, i.Instruction)
}

type Comment struct {
	Base
	Comment string
}

func (c Comment) String() string {
	return fmt.Sprintf(`(comment "%s")`, c.Comment)
}

type Ingredient struct {
	Base
	Name     string
	Quantity string
	Unit     string
}

func (i Ingredient) String() string {
	args := []string{}
	if i.Quantity != "" {
		args = append(args, fmt.Sprintf(":quantity %s", strconv.Quote(i.Quantity)))
	}
	if i.Unit != "" {
		args = append(args, fmt.Sprintf(":unit %s", strconv.Quote(i.Unit)))
	}

	return fmt.Sprintf(`(ingredient "%s" {%s})`, i.Name, strings.Join(args, " "))
}

type Cookware struct {
	Base
	Name string
}

func (c Cookware) String() string {
	return fmt.Sprintf(`(cookware "%s")`, c.Name)
}

type Timer struct {
	Base
	Name      string
	Magnitude string
	Unit      string
}

func (t Timer) String() string {
	return fmt.Sprintf(`(timer %s {:magnitude %s :unit %s})`, strconv.Quote(t.Name), strconv.Quote(t.Magnitude), strconv.Quote(t.Unit))
}

type Metadata struct {
	Base
	Key   string
	Value string
}

func (m Metadata) String() string {
	return fmt.Sprintf(`(metadata "%s" "%s")`, m.Key, m.Value)
}
