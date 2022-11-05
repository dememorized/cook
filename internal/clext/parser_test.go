package clext

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"
)

//go:embed testdata/pancakes.cook
var pancakes string

func TestTokenize(t *testing.T) {
	const filename = "testdata/pancakes.cook"
	tokens, _ := Tokenize(filename, strings.NewReader(pancakes))

	ast := Parse(filename, tokens)
	fmt.Printf("%s\n", ast)

	for _, step := range ast.Steps {
		fmt.Println(step.Ingredients())
	}
}
