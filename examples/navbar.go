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
