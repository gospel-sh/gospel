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

func (p *Parser) parseExpression(delimiter string) (*Expression, error) {

	success := false
	defer p.push(&success)()

	return nil, nil
}
