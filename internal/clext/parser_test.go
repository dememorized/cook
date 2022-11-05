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
	tokens, parseErrs := Tokenize(filename, strings.NewReader(pancakes))
	if len(parseErrs) != 0 {
		t.Error(parseErrs)
		t.FailNow()
	}

	ast, err := Parse(filename, tokens)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	fmt.Printf("%s\n", ast)

	for _, step := range ast.Steps {
		fmt.Println(step.Ingredients())
	}
}
