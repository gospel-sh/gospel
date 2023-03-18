package gospel

import (
	"fmt"
)

type Element interface {
	Render(Context) string
}

type HTMLElement struct {
	Tag      string
	Children []Element
}

func (h *HTMLElement) Render(c Context) string {

	renderedChildren := ""

	for _, child := range h.Children {
		renderedChildren += child.Render(c)
	}

	return fmt.Sprintf("<%[1]s>%[2]s</%[1]s>", h.Tag, renderedChildren)
}

type Literal struct {
	Value string
}

func (l *Literal) Render(c Context) string {
	return l.Value
}

func children(args ...any) (children []Element) {

	children = make([]Element, 0, len(args))

	for _, arg := range args {
		if elem, ok := arg.(Element); ok {
			children = append(children, elem)
		} else if str, ok := arg.(string); ok {
			children = append(children, &Literal{str})
		}
	}

	return
}

func Html(args ...any) Element {

	return &HTMLElement{"html", children(args...)}
}

func H1(args ...any) Element {
	return &HTMLElement{"h1", children(args...)}
}

func Div(args ...any) Element {
	return &HTMLElement{"div", children(args...)}
}
