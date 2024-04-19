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

func Hero() Element {
	return Div(
		Styles(
			MinHeight(Calc(Sub(Vh(100), Px(68)))),
			Display("block"),
			Background("rgb(84, 35, 231)"),
		),
		Div(
			// Our container
			Styles(
				MaxWidth(Px(1200)),
				Margin("0 auto"),
				PaddingTop(Px(50)),
				// we add some padding
				PaddingLeft(Px(18)),
				PaddingRight(Px(18)),
				// for large screen sizes we remove the padding
				Desktop(
					PaddingLeft(0),
					PaddingRight(0),
				),
				// for mobile we reduce the top padding
				Mobile(
					PaddingTop(Px(20)),
				),
			),
			Div(
				Styles(
					Display("flex"),
					FlexDirection("row"),
					AlignItems("center"),
					JustifyContent("space-between"),
				),
				Div(
					Styles(
						FlexGrow(1),
						FlexBasis(Percent(45)),
					),
					H1(
						Styles(
							FontWeight(600),
							FontSize(Rem(4.5)),
							LetterSpacing(Px(-2)),
							LineHeight(Percent(100)),
						),
						"Payments, tax & subscriptions for software companies",
					),
					P(
						Styles(
							MarginTop(Px(30)),
							Opacity(Percent(70)),
							FontSize(Rem(1.2)),
						),
						"As your merchant of record, we handle the tax compliance burden so you can focus on more revenue and less headache.",
					),
					P(
						Styles(
							MarginTop(Px(40)),
							A(
								Background("white"),
								BorderRadius(Px(40)),
								Color("rgb(40, 40, 40)"),
								TextDecoration("none"),
								Padding(Px(14)),
								PaddingLeft(Px(30)),
								PaddingRight(Px(30)),
								FontWeight(500),
								FontSize(Rem(1.2)),
							),
						),
						A(
							Href("#"),
							"Get started for free ", L("&#129106;"),
						),
					),
				),
				Div(
					Styles(
						FlexGrow(1),
						FlexBasis(Percent(55)),
						Position("relative"),
						Height(Px(600)),
					),
					Div(
						Styles(
							Background("white"),
							Transform("rotate3d(2, 1, 1, -45deg)"),
							BorderRadius(Px(20)),
							Height(Px(500)),
							Width(Px(700)),
							Position("absolute"),
							Left(Px(100)),
							MarginTop(Px(100)),
						),
					),
				),
			),
		),
	)
}
