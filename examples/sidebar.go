package examples

import (
	. "github.com/gospel-sh/gospel"
)

var SidebarCSS = MakeStylesheet("sidebar")

func SidebarExample(c Context) Element {
	return F(
		Div(
			Styles(),
			"this is a sidebar",
			Span("this should be blue", A("and this red")),
		),
	)
}

func init() {
	Examples = append(Examples, Example{
		"Sidebar",
		SidebarExample,
		SidebarCSS,
	})
}
