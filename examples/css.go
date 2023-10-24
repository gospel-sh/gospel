package examples

import (
	. "github.com/gospel-sh/gospel"
	"strings"
)

var CSS = MakeStylesheet()

var BR4 = CSS.NamedRule("br-4", BorderRadius(Px(4)))

var IsError = CSS.NamedFragment(
	"is-error",
	BackgroundColor("#f00"),
)

var IsSuccess = CSS.NamedFragment(
	"is-success",
	BackgroundColor("#0f0"),
)

var ButtonStyle = CSS.NamedRule(
	"button",
	IsSuccess.Derive(
		Span(
			TextDecoration("underline"),
		),
	),
	IsError.Derive(Span(BorderRadius("4px"))),
)

var MessageStyle = CSS.NamedRule(
	"message",
	IsSuccess.Derive(
		BackgroundColor("#afe"),
		P(
			TextDecoration("underline"),
		),
	),
)

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

var Scaling = 1.0

var Scaled = CSS.Rule(
	Width("400px"),
	Height("400px"),
	Position("relative"),
	Border("4px solid #eee"),
	Padding(Px(10)),
	Iframe(
		Width(Calc(Sub(Percent(100/Scaling), Px(14)))),
		Height(Calc(Sub(Percent(100/Scaling), Px(14)))),
		Position("absolute"),
		Transform(Scale(Scaling)),
		TransformOrigin("top left"),
	),
)

func ButtonsExample(c Context) Element {
	return Button(
		Styles(ButtonStyle, IsSuccess),
		Span("foo"),
	)
}

func NavbarExample(c Context) Element {
	return F(
		c.DeferElement("styles", func(c Context) Element { return CSS.Styles() }),
		Div(
			Styles(BR4, Rounded),
			"this is a rounded div",
			Span("this should be blue", A("and this red")),
		),
	)
}

func SidebarExample(c Context) Element {
	return F(
		c.DeferElement("styles", func(c Context) Element { return CSS.Styles() }),
		Div(
			Styles(BR4, Rounded),
			"this is a sidebar",
			Span("this should be blue", A("and this red")),
		),
	)
}

type Example struct {
	Name string
	View func(c Context) Element
}

var Examples = []Example{
	{
		"Buttons",
		ButtonsExample,
	},
	{
		"Navbar",
		NavbarExample,
	},
	{
		"Sidebar",
		SidebarExample,
	},
}

func iframe(url string) Element {
	return F(
		H1(url),
		Div(
			Styles(Scaled),
			Iframe(
				Src(url),
			),
		),
	)
}

func CSSExample(c Context) Element {

	router := UseRouter(c)

	MobileWidth.Value = 640

	routes := []*RouteConfig{}
	iframes := []any{}

	for _, exampleConfig := range Examples {
		url := Fmt("/%s", strings.ToLower(exampleConfig.Name))
		routes = append(routes, Route(url, exampleConfig.View))
		iframes = append(iframes, iframe(Fmt("/css%s", url)))
	}

	routes = append(routes, Route("",
		F(
			iframes...,
		),
	),
	)

	return F(
		c.DeferElement("styles", func(c Context) Element { return CSS.Styles() }),
		router.Match(
			c,
			routes...,
		),
	)
}
