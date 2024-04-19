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
