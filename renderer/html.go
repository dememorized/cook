package renderer

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/dememorized/cook/aromalang"
	"html/template"
	"reflect"
	"strings"
)

type HTML struct {
	AST *aromalang.AST
}

//go:embed html-template.html
var htmlTemplateRaw string
var htmlTemplate = template.Must(template.New("html-template").Parse(htmlTemplateRaw))

type htmlData struct {
	Steps []htmlStep
}

type htmlStep struct {
	Components []any
}

func (h HTML) Render() ([]byte, error) {
	if h.AST == nil {
		return nil, fmt.Errorf("no recipe provided")
	}

	data := htmlData{}

	for _, step := range h.AST.Recipe.Steps {
		comps := []any{}
		for _, c := range step.Components {
			r, err := htmlRenderComponent(c)
			if err != nil {
				return nil, err
			}

			comps = append(comps, r)
		}
		data.Steps = append(
			data.Steps,
			htmlStep{Components: comps},
		)
	}

	buf := bytes.Buffer{}
	err := htmlTemplate.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func htmlRenderComponent(component aromalang.Component) (any, error) {
	switch c := component.(type) {
	case aromalang.Instruction:
		return c.Instruction, nil
	case aromalang.Ingredient:
		return htmlRenderIngredient(c)
	case aromalang.Cookware:
		return htmlRenderCookware(c)
	case aromalang.Timer:
		return htmlRenderTimer(c)
	default:
		return nil, fmt.Errorf("cannot render component of type %s", reflect.TypeOf(component))
	}
}

var (
	htmlTemplateIngredient = template.Must(template.New("html-ingredients").Parse(
		`<span class="cook-ingredient">{{ if .Quantity }}{{ .Quantity }} {{ end }}{{ if .Unit }}{{ .Unit }} {{ end}}{{ .Name }}</span>`,
	))
	htmlTemplateTimer = template.Must(template.New("html-timer").Parse(
		`<span class="cook-timer" alt="{{ .Name }}">{{ .Magnitude }} {{ .Unit }}</span>`,
	))
	htmlTemplateCookware = template.Must(template.New("html-timer").Parse(
		`<span class="cook-cookware">{{ .Name }}</span>`,
	))
)

func htmlRenderIngredient(ingredient aromalang.Ingredient) (template.HTML, error) {
	buf := &strings.Builder{}
	err := htmlTemplateIngredient.Execute(buf, ingredient)
	if err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

func htmlRenderTimer(timer aromalang.Timer) (template.HTML, error) {
	buf := &strings.Builder{}
	err := htmlTemplateTimer.Execute(buf, timer)
	if err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

func htmlRenderCookware(cookware aromalang.Cookware) (template.HTML, error) {
	buf := &strings.Builder{}
	err := htmlTemplateCookware.Execute(buf, cookware)
	if err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
