package renderer

import (
	_ "embed"
	"fmt"
	"github.com/dememorized/cook/aromalang"
	"strings"
	"testing"
)

//go:embed testdata/pancakes.aroma
var sample string

func TestGenerateHTML(t *testing.T) {
	res, err := aromalang.Parse("pancakes.aroma", strings.NewReader(sample))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	b, err := HTML{AST: res}.Render()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	fmt.Println(string(b))
}
