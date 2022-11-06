package aromalang

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"
)

//go:embed testdata/pancakes.aroma
var pancakes string

func TestTokenize(t *testing.T) {
	const filename = "testdata/pancakes.aroma"
	tokens, parseErrs := Tokenize(filename, strings.NewReader(pancakes))
	if len(parseErrs) != 0 {
		t.Error(parseErrs)
		t.FailNow()
	}

	fmt.Println(tokens)
}
