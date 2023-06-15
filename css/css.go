package css

import (
	"strings"

	"github.com/gospel-dev/gospel"
)

type Css struct {
	gospel.HTMLElement
}

type Context struct {
}

func (c *Css) RenderElement(context gospel.Context) string {
	return c.RenderChildren(context)
}

func mapCss(elements []*gospel.HTMLElement, c *Context) {
	for _, element := range elements {
		styles := []string{}
		for _, arg := range element.Args {
			if class, ok := arg.(*Class); ok {
				styles = append(styles, class.Render(c))
			}
		}

		if len(styles) > 0 {

			stylesStr := strings.Join(styles, " ")

			element.Attributes = append(element.Attributes, gospel.Style(stylesStr))

		}
	}
}

func CSS(children ...*gospel.HTMLElement) gospel.Element {

	c := &Context{}

	mapCss(children, c)

	return &Css{
		HTMLElement: gospel.HTMLElement{
			Children: children,
		},
	}
}

type ContextValue interface {
	Get(*Context) any
}

func (c *Class) Render(context *Context) string {
	// we render the CSS class or style attribute

	v := c.Value

	if vr, ok := v.(ContextValue); ok {
		v = vr.Get(context)
	}

	return gospel.Fmt("%s: %v;", c.Property, v)
}

type Class struct {
	Property string
	Value    any
}

func Prop(property string) func(value any) *Class {
	return func(value any) *Class {
		return &Class{
			Property: property,
			Value:    value,
		}
	}
}

// Flexbox

var Flex = Prop("display")("flex")
var FlexDirection = Prop("flex-direction")
var AlignItems = Prop("align-items")
var JustifyContent = Prop("justify-content")

// Line height

var LineHeight = Prop("line-height")

// Geometry

var Width = Prop("width")
var Height = Prop("height")
var MaxWidth = Prop("max-width")
var MinWidth = Prop("min-width")
