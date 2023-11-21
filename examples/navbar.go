package examples

import (
	. "github.com/gospel-sh/gospel"
)

var NavbarCSS = MakeStylesheet("navbar")

func Navbar(items ...any) Element {
	return Nav(
		Styles(
			Display("block"),
			Padding(DefaultPadding),
			Background("rgb(84, 35, 231)"),
			Height(Px(50)),
		),
		Div(
			Styles(
				Padding(Px(10)),
				MaxWidth(Px(1200)),
				Margin("0 auto"),
			),
			F(
				items...,
			),
		),
	)
}

func NavbarExample(c Context) Element {
	return F(
		Div(
			Styles(),
			Navbar(
				H1(
					Styles(
						TextTransform("uppercase"),
						FontWeight("bolder"),
					),
					"lemonsqueezy",
				),
			),
		),
	)
}

func init() {
	Examples = append(Examples, Example{
		"Navbar",
		NavbarExample,
		NavbarCSS,
	})
}
