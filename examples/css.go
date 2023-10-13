package examples

import (
	. "github.com/gospel-sh/gospel"
    . "github.com/gospel-sh/gospel/css"
)

var CSS = MakeStylesheet()

var Rounded = CSS.Rule(
	BorderRadius("4px"),
)

func CSSExample(c Context) Element {
	return F(
		CSS.Styles(),
		Div(
			Style(Rounded),
			"this is a rounded div",
		),
	)
}