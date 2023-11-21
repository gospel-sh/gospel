package examples

import (
	. "github.com/gospel-sh/gospel"
)

var LandingPageCSS = MakeStylesheet("landingpage")

func init() {
	LandingPageCSS.AddRule(
		TagRule("html")(
			Height(Vh(100)),
			Color("white"),
			FontFamily("Helvetica, Arial, sans-serif"),
		),
	)
	LandingPageCSS.AddRule(TagRule("body")(MinHeight(Vh(100)), Position("relative")))
}

func LandingPageExample(c Context) Element {
	return F(
		Navbar(
			Div(
				Styles(
					Display("flex"),
					FlexDirection("row"),
				),
				H1(
					Styles(
						TextTransform("lowercase"),
						FontWeight("600"),
					),
					"worf",
				),
				Ul(
					Styles(
						MarginLeft(Px(20)),
						FontSize(Rem(1.1)),
						FlexGrow(1),
						ListStyle("none"),
						Li(
							Display("inline-block"),
							Padding(Px(6)),
							MarginRight(Px(10)),
							LastChild(
								MarginRight(0),
							),
							A(
								Color("white"),
							),
						),
					),
					Li(
						A(Href("#"), "foo"),
					),
					Li(
						A(Href("#"), "bar"),
					),
					Li(
						A(Href("#"), "baz"),
					),
				),
				Ul(
					Styles(
						MarginLeft(Px(20)),
						FontSize(Rem(1.0)),
						FlexGrow(0),
						ListStyle("none"),
						Li(
							Display("inline-block"),
							Padding(Px(6)),
							MarginRight(Px(10)),
							LastChild(
								MarginRight(0),
								A(
									Color("rgb(40, 40, 40)"),
									TextDecoration("none"),
									BorderRadius(Px(24)),
									Padding(Px(10)),
									FontWeight(600),
									Background("white"),
								),
							),
							A(
								Color("rgb(200, 200, 200)"),
								TextDecoration("none"),
								Padding(Px(10)),
							),
						),
					),
					Li(
						A(Href("#"), "sign in ", L("&#x1F465;")),
					),
					Li(
						A(Href("#"), "try now ", L("&#129106;")),
					),
				),
			),
		),
		Hero(),
	)
}

func init() {
	Examples = append(Examples, Example{
		"LandingPage",
		LandingPageExample,
		LandingPageCSS,
	})
}
