package gospel

import (
	"fmt"
)

type Element interface {
	RenderElement(Context) string
}

type Attribute interface {
	RenderAttribute(Context) string
}

type HTMLElement struct {
	Tag        string
	Void       bool
	Children   []Element
	Attributes []Attribute
}

type HTMLAttribute struct {
	Name  string
	Value any
	Args  []any
}

func Attrib(tag string) func(value any, args ...any) *HTMLAttribute {
	return func(value any, args ...any) *HTMLAttribute {
		return &HTMLAttribute{
			Name:  tag,
			Value: value,
			Args:  args,
		}
	}
}

func (a *HTMLAttribute) RenderAttribute(c Context) string {
	return fmt.Sprintf("%s=\"%s\"", a.Name, a.Value)
}

func (h *HTMLElement) RenderElement(c Context) string {

	renderedAttributes := ""

	for _, attribute := range h.Attributes {
		renderedAttributes += " " + attribute.RenderAttribute(c)
	}

	if h.Void {
		return fmt.Sprintf("<%[1]s%[2]s/>", h.Tag, renderedAttributes)
	} else {

		renderedChildren := ""

		for _, child := range h.Children {
			renderedChildren += child.RenderElement(c)
		}

		return fmt.Sprintf("<%[1]s%[3]s>%[2]s</%[1]s>", h.Tag, renderedChildren, renderedAttributes)
	}

}

type Literal struct {
	Value string
}

func (l *Literal) RenderElement(c Context) string {
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

func attributes(args ...any) (attributes []Attribute) {

	attributes = make([]Attribute, 0, len(args))

	for _, arg := range args {
		if elem, ok := arg.(Attribute); ok {
			attributes = append(attributes, elem)
		}
	}

	return
}

func Tag(tag string) func(args ...any) Element {
	return func(args ...any) Element {
		return &HTMLElement{tag, false, children(args...), attributes(args...)}
	}
}

func VoidTag(tag string) func(args ...any) Element {
	return func(args ...any) Element {
		return &HTMLElement{tag, true, children(args...), attributes(args...)}
	}
}

// HTML Attributes

var Lang = Attrib("lang")
var Charset = Attrib("charset")
var Rel = Attrib("rel")
var Sizes = Attrib("sizes")
var Href = Attrib("href")
var Type = Attrib("type")
var Class = Attrib("class")
var Id = Attrib("id")

// HTML Tags

var Html = Tag("html")
var Div = Tag("div")
var H1 = Tag("h1")
var Title = Tag("title")
var Link = VoidTag("link")
var Meta = VoidTag("meta")
var Head = Tag("head")
var Body = Tag("body")
