package examples

import (
	. "github.com/gospel-sh/gospel"
)

var UtilityCSS = MakeStylesheet("utility")

func init() {
	for i := 0; i < 48; i++ {
		UtilityCSS.NamedRule(Fmt("bw-%d", i), BorderWidth(Px(float64(i))))
	}
}

func UtilityCSSExample(c Context) Element {
	return F(
		Div(
			Class("bw-4"),
			Style("border-style: solid; border-color: black;"),
			"test",
		),
	)
}

func init() {
	Examples = append(Examples, Example{
		"Utility CSS",
		UtilityCSSExample,
		UtilityCSS,
	})
}
