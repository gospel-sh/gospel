// Gospel - Golang Simple Extensible Web Framework
// Copyright (C) 2019-2024 - The Gospel Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the 3-Clause BSD License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// license for more details.
//
// You should have received a copy of the 3-Clause BSD License
// along with this program.  If not, see <https://opensource.org/licenses/BSD-3-Clause>.

package examples

import (
	. "github.com/gospel-sh/gospel"
	"strings"
)

var CSS = MakeStylesheet("site")

func init() {
	// general style reset
	CSS.AddRule(TagRule("*")(
		FontWeight("lighter"),
		Border(0),
		Padding(0),
		Margin(0),
	))
}

var Scaling = 0.7

type Example struct {
	Name string
	View func(c Context) Element
	CSS  *Stylesheet
}

var Examples = []Example{}

func iframe(url string) Element {
	return F(
		H1(url),
		Div(
			Styles(
				Height("700px"),
				Position("relative"),
				Border("4px solid #eee"),
				Iframe(
					Width(Percent(100/Scaling)),
					Height(Percent(100/Scaling)),
					Position("absolute"),
					Transform(Scale(Scaling)),
					TransformOrigin("top left"),
					Border("none"),
					Margin(0),
					Padding(0),
				),
			),
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

		func(config Example) {

			url := Fmt("/%s", strings.ToLower(config.Name))

			routes = append(routes, Route(url, func(c Context) Element {
				return F(
					config.CSS.Styles(),
					config.View(c),
				)
			}))
			iframes = append(iframes, iframe(Fmt("/css%s", url)))

		}(exampleConfig)

	}

	routes = append(routes, Route("",
		F(
			iframes...,
		),
	),
	)

	return F(
		CSS.Styles(),
		Styled(
			"root",
			router.Match(
				c,
				routes...,
			),
		),
	)
}
