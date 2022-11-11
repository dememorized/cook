package aromalang

import (
	"fmt"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	const filename = "testdata/pancakes.aroma"
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

	fmt.Println(ast)
}
