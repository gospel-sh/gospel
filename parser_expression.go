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

package gospel

type ExpressionType int

const (
	HTMLExpression = iota
	CSSExpression
)

/*
Program              := { Statement }
Statement            := Assignment | FunctionDef | IfStatement | Loop | ReturnStatement | ExpressionStatement
Assignment           := Identifier "=" Expression
FunctionDef          := "def" Identifier "(" [ Parameters ] ")" Block
Parameters           := Identifier { "," Identifier }
IfStatement          := "if" Expression Block [ "else" Block ]
Loop                 := "while" Expression Block | "for" Identifier "in" Expression Block
Block                := "{" { Statement } "}"
ReturnStatement      := "return" [ Expression ]
ExpressionStatement  := Expression
Expression           := ListComprehension | Term { ("+" | "-") Term }
ListComprehension    := "[" Expression "for" Identifier "in" Expression [ "if" Expression ] "]"
Term                 := Factor { ("*" | "/") Factor }
Factor               := UnaryOp Factor | Number | Identifier | FunctionCall | "(" Expression ")"
UnaryOp              := "-" | "+"
FunctionCall         := Identifier "(" [Expression { "," Expression }] ")"
Identifier           := Char { Char | Digit }
Number               := Digit { Digit }
*/

type Expression struct {
}

func (p *Parser) ParseExpression(delimiter string) (*Expression, error) {

	success := false
	defer p.push(&success)()

	return nil, nil
}
