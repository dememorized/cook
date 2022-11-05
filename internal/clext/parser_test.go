package clext

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

//go:embed testdata/pancakes.cook
var pancakes string

// canonicalTestFile is a JSON version of the test suite provided by
// the upstream cooklang/spec repository.
// https://github.com/cooklang/spec/blob/fa9bc51515b3317da434cb2b5a4a6ac12257e60b/tests/canonical.yaml
//
//go:embed testdata/canonical.json
var canonicalTestFile []byte

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

type canonicalTests struct {
	Version int                      `json:"version"`
	Tests   map[string]canonicalTest `json:"tests"`
}

type canonicalTest struct {
	Source string `json:"source"`
	Result struct {
		Steps    [][]map[string]interface{} `json:"steps"`
		Metadata map[string]string          `json:"metadata"`
	} `json:"result"`
}

func TestCanonical(t *testing.T) {
	var tests canonicalTests
	err := json.Unmarshal(canonicalTestFile, &tests)
	if err != nil {
		t.Error(err)
		return
	}

	skippedTests := map[string]struct{}{
		"testFractions":           {},
		"testFractionsWithSpaces": {},
		"testTimerFractional":     {},
	}

	for k, v := range tests.Tests {
		k, v := k, v
		t.Run(k, func(t *testing.T) {
			if _, exists := skippedTests[k]; exists {
				t.Skipf("Skipping test %s", k)
			}

			tokens, errs := Tokenize(k, strings.NewReader(v.Source))
			if len(errs) != 0 {
				t.Errorf("got errors when tokenizing: %v", errs)
				return
			}

			recipe, err := Parse(k, tokens)
			if err != nil {
				t.Errorf("got error when parsing recipe: %v", err)
				return
			}

			if !reflect.DeepEqual(recipe.Metadata(), v.Result.Metadata) {
				t.Errorf("expected metadata for result and recipe to be equal, but they were not:\n"+
					"expected: %v\n"+
					"got: %v\n", v.Result.Metadata, recipe.Metadata())
				return
			}

			for i, step := range recipe.Steps {
				if !step.HasInstructions() {
					continue
				}

				StepsMatches(t, step, v.Result.Steps[i])
			}
		})
	}
}

func StepsMatches(t testing.TB, step Step, expected []map[string]interface{}) {
	if len(step.Components) < len(expected) {
		t.Errorf("there are more expected components than actual components in the step")
		return
	}
	i := 0

	for _, exp := range expected {
		tpe, ok := exp["type"]
		if !ok {
			t.Errorf("expected 'type' in the expected step %v", expected)
			t.FailNow()
		}

		var c Component
		i, c = nextComponent(i, step)
		switch tpe {
		case "ingredient":
			ingredient, ok := c.(Ingredient)
			if !ok {
				t.Errorf("expected next kind to be ingredient, got %s, %v %s", reflect.TypeOf(c), expected, step)
				return
			}
			equals(t, ingredient.Name, exp["name"])
			if ingredient.Quantity != "" {
				equals(t, ingredient.Quantity, exp["quantity"])
			} else {
				equals(t, "some", exp["quantity"])
			}
			equals(t, ingredient.Unit, exp["units"])
		case "cookware":
			cookware, ok := c.(Cookware)
			if !ok {
				t.Errorf("expected next kind to be cookware, got %s", reflect.TypeOf(c))
				return
			}
			equals(t, cookware.Name, exp["name"])
		case "text":
			instruction, ok := c.(Instruction)
			if !ok {
				t.Errorf("expected next kind to be instruction, got %s", reflect.TypeOf(c))
				return
			}
			equals(t, instruction.Instruction, exp["value"])
		case "timer":
			timer, ok := c.(Timer)
			if !ok {
				t.Errorf("expected next kind to be ingredient, got %s", reflect.TypeOf(c))
				return
			}

			equals(t, timer.Name, exp["name"])
			equals(t, timer.Magnitude, exp["quantity"])
			equals(t, timer.Unit, exp["units"])
		default:
			t.Errorf("unknown type %s", tpe)
			return
		}
	}
}

func equals(t testing.TB, actual any, expected any) {
	t.Helper()

	actualString, isString := actual.(string)
	_, expectNumber := expected.(float64)

	cmp := actual
	if isString && expectNumber {
		var err error
		cmp, err = strconv.ParseFloat(actualString, 64)
		if err != nil {
			t.Errorf("got error when converting value to float")
		}

		actual = fmt.Sprintf("%s (converted to %f)", actual, cmp)
	}

	if !reflect.DeepEqual(cmp, expected) {
		t.Errorf("expected %v = %v [%s = %s]", actual, expected,
			reflect.TypeOf(actual), reflect.TypeOf(expected))
	}
}

func nextComponent(i int, step Step) (int, Component) {
	for j, component := range step.Components[i:] {
		switch component.(type) {
		case Instruction, Ingredient, Cookware, Timer:
			return i + j + 1, component
		}
	}
	return 0, nil
}
