package examples

import (
	. "github.com/gospel-sh/gospel"
)

var CSS = MakeStylesheet()

var BR4 = CSS.NamedRule("br-4", BorderRadius(Px(4)))

var Rounded = CSS.NamedRule(
	"rounded",
	BorderRadius(Px(4)),
	BorderColor("grey"),
	BorderStyle("solid"),
	BorderWidth(Px(2)),
	Padding(Px(10)),
	// mobile view
	Mobile(
		BorderRadius(Px(8)),
		Padding(Px(5)),
	),
	Span(
		Padding(Px(10)),
		Color("blue"),
		// applies to any a element within
		A(
			Color("red"),
			Mobile(
				Padding(Px(10)),
			),
		),
	),
)

func CSSExample(c Context) Element {

	MobileWidth.Value = 640

	return F(
		c.DeferElement("styles", func(c Context) Element { return CSS.Styles() }),
		Div(
			Styles(BR4, Rounded),
			"this is a rounded div",
			Span("this should be blue", A("and this red")),
		),
	)
}
