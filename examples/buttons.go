package examples

import (
	. "github.com/gospel-sh/gospel"
)

var ButtonsCSS = MakeStylesheet("buttons")

var IsError = ButtonsCSS.NamedFragment(
	"is-error",
	BackgroundColor("#f00"),
)

var IsSuccess = ButtonsCSS.NamedFragment(
	"is-success",
	BackgroundColor("#0f0"),
)

var ButtonStyle = ButtonsCSS.NamedRule(
	"button",
	Padding(DefaultPadding),
	BorderRadius(DefaultRadius),
	BorderStyle("solid"),
	BorderWidth(Px(1)),
	BorderColor("#bbb"),
	BackgroundColor("#eee"),
	Hover(
		BackgroundColor("#fff"),
	),
	IsSuccess.Derive(
		Span(
			TextDecoration("underline"),
		),
	),
	IsError.Derive(Span(BorderRadius("4px"))),
)

func ButtonsExample(c Context) Element {
	return Button(
		Styles(ButtonStyle, IsSuccess, Padding(DefaultPadding)),
		Span("button"),
	)
}

func init() {
	Examples = append(Examples, Example{
		"Buttons",
		ButtonsExample,
		ButtonsCSS,
	})
}
