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

func (c *Class) Render(context *Context) string {
	// we render the CSS class or style attribute
	// at
	return gospel.Fmt("%s: %v;", c.Property, c.Value)
}

type Class struct {
	Property string
	Value    any
}

func Width(value any) *Class {
	return &Class{
		Property: "width",
		Value:    value,
	}
}

func Flex() *Class {
	return &Class{
		Property: "display",
		Value:    "flex",
	}
}
