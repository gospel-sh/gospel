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

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var WHITESPACE = " \t\n"

type Program struct {
	Statements []*Statement
}

type StatementType int

const (
	HTMLStmt = iota
	CSSStmt
)

type Statement struct {
	Type StatementType
	HTML *HTMLStatement
	CSS  *CSSStatement
}

type HTMLStatementType int

const (
	TemplateStmt = iota
)

type HTMLStatement struct {
	Type     HTMLStatementType
	Template *HTMLTemplate
}

type HTMLTemplate struct {
	Name string
	// we just return a normal HTML element
	Element *HTMLElement
}

type CSSStatement struct {
	Name  string
	Rules []*Rule
}

type Parser struct {
	Pos     int
	stack   []int
	Source  string
	Comment string
}

func (p *Parser) push(success *bool) func() {
	p.stack = append(p.stack, p.Pos)

	return func() {
		if *success {
			return
		}
		p.Pos = p.stack[len(p.stack)-1]
		p.stack = p.stack[:len(p.stack)-1]
	}
}

func (p *Parser) has(prefix string, allowWhitespace bool) bool {
	str := p.Source[p.Pos:]

	if allowWhitespace {
		str = strings.TrimLeft(str, WHITESPACE)
	}

	if len(str) < len(prefix) {
		return false
	}

	return str[:len(prefix)] == prefix
}

func (p *Parser) consumeWhitespace() error {
	return p.consume("", true)
}

func (p *Parser) consume(prefix string, allowWhitespace bool) error {

	if len(prefix) > len(p.Source)-p.Pos {
		return fmt.Errorf("prefix '%s' not found", prefix)
	}

	str := p.Source[p.Pos:]

	if allowWhitespace {
		lastLen := len(str)
		str = strings.TrimLeft(str, " \t\n")
		p.Pos += lastLen - len(str)
	}

	if str[:len(prefix)] != prefix {
		return fmt.Errorf("prefix '%s' not found: '%s'(%d)", prefix, str[:len(prefix)], p.Pos)
	}

	p.Pos += len(prefix)

	return nil

}

func (p *Parser) consumeRegexp(re *regexp.Regexp) ([]string, error) {
	str := p.Source[p.Pos:]

	if match := re.FindStringSubmatch(str); match == nil {
		return nil, fmt.Errorf("did not match")
	} else {
		p.Pos += len(match[0])
		return match, nil
	}
}

func (p *Parser) consumeIdentifier() (string, error) {
	// identifiers can be made of a-zA-Z

	str := p.Source[p.Pos:]

	// we decode the first rune
	r, _ := utf8.DecodeRuneInString(str)
	if !unicode.Is(unicode.Letter, r) {
		return "", fmt.Errorf("expected an identifier to start with a letter, got '%v'.", r)
	}

	newStr := strings.TrimLeftFunc(str, func(v rune) bool {
		return unicode.In(v, unicode.Letter, unicode.Number)
	})

	identifier := str[:len(str)-len(newStr)]
	p.Pos += len(identifier)
	return identifier, nil
}

func (p *Parser) end() bool {
	// we allow whitespace
	return strings.TrimLeft(p.Source[p.Pos:], WHITESPACE) == ""
}

func (p *Parser) Parse(source string) (*Program, error) {

	p.Pos = 0
	p.Source = source

	if statements, err := p.ParseStatements(); err != nil {
		return nil, err
	} else {
		return &Program{
			Statements: statements,
		}, nil
	}
}

func (p *Parser) ParseStatements() ([]*Statement, error) {
	statements := make([]*Statement, 0, 1)

	for {
		statement, err := p.ParseStatement()

		if err != nil {
			return statements, err
		}

		if statement == nil {
			// end of statements reached
			return statements, nil
		}

		statements = append(statements, statement)
	}
}

func (p *Parser) ParseStatement() (*Statement, error) {
	if p.has("html ", true) {
		// this is a HTML statement
		if htmlStatement, err := p.ParseHTMLStatement(); err != nil {
			return nil, err
		} else {
			return &Statement{
				Type: HTMLStmt,
				HTML: htmlStatement,
			}, nil
		}
	} else if p.has("css ", true) {
		// this is a CSS statement
		if cssStatement, err := p.ParseCSSStatement(); err != nil {
			return nil, err
		} else {
			return &Statement{
				Type: CSSStmt,
				CSS:  cssStatement,
			}, nil
		}
	} else if p.end() {
		// we're at the end, nothing more to parse
		return nil, nil
	}
	return nil, fmt.Errorf("expected a statement, but didn't find one (%d)", p.Pos)
}

func (p *Parser) ParseHTMLStatement() (*HTMLStatement, error) {

	// we consume the HTML keyword
	p.consume("html", true)

	if err := p.consume(" ", false); err != nil {
		// we expect at least one whitespace character
		return nil, fmt.Errorf("expected a whitespace after the 'html' keyword: %w", err)
	}

	if p.has("template", false) {
		if template, err := p.ParseTemplate(); err != nil {
			return nil, fmt.Errorf("invalid HTML template statement: %w", err)
		} else {
			return &HTMLStatement{
				Type:     TemplateStmt,
				Template: template,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid HTML statement")
}

func (p *Parser) ParseTemplate() (*HTMLTemplate, error) {

	if err := p.consume("template", false); err != nil {
		return nil, fmt.Errorf("expected keyword 'template': %w", err)
	}

	if err := p.consume(" ", false); err != nil {
		return nil, fmt.Errorf("expected a whitespace after 'template': %w", err)
	}

	t := &HTMLTemplate{}

	if identifier, err := p.consumeIdentifier(); err != nil {
		return nil, fmt.Errorf("expected an identifier after 'template': %w", err)
	} else {
		t.Name = identifier
	}

	if htmlElement, err := p.ParseHTMLElement(); err != nil {
		return nil, fmt.Errorf("error parsing HTML element: %w", err)
	} else if htmlElement == nil {
		return nil, fmt.Errorf("expected a HTML element")
	} else {
		t.Element = htmlElement
	}

	return t, nil
}

var tagRegexp = regexp.MustCompile(`[a-z][a-z\-0-9]*`)

func (p *Parser) ParseHTMLChildren() ([]any, error) {

	success := false
	defer p.push(&success)()

	elements := make([]any, 0, 10)

	for {

		pos := p.Pos

		if element, err := p.ParseFunctionBlock(); err != nil {
			return nil, fmt.Errorf("error parsing function block: %w", err)
		} else if element != nil {
			elements = append(elements, element)
			continue
		}

		if element, err := p.ParseHTMLElement(); err != nil {
			return nil, err
		} else if element != nil {
			elements = append(elements, element)
			continue
		}

		if element, err := p.ParseHTMLTextNode(); err != nil {
			return nil, err
		} else if element != nil {
			elements = append(elements, element)
			continue
		}

		if p.Pos == pos {
			// we haven't made any progress, no more elements
			break
		}
	}

	success = true
	return elements, nil
}

var htmlTextNodeRegexp = regexp.MustCompile(`(?ms)[^<\{]*`)

func (p *Parser) ParseHTMLTextNode() (*HTMLElement, error) {
	textContent := ""
loop:
	for {
		if text, err := p.consumeRegexp(htmlTextNodeRegexp); err != nil {
			return nil, fmt.Errorf("not a text node: %w", err)
		} else {
			textContent += text[0]
			if p.Pos < len(p.Source) {
			escapes:
				switch p.Source[p.Pos] {
				case '<': // this is a HTML element
					break
				case '{': // this might be an expression
					if p.Pos+1 < len(p.Source) {
						switch p.Source[p.Pos+1] {
						case ':': // this is a macro
							break escapes
						}
					}
					// this isn't a special block, we consume the '{' and continue
					if err := p.consume("{", false); err != nil {
						return nil, fmt.Errorf("expected '{': %v", err)
					}
					textContent += "{"
					continue loop
				}
			}
			// we can trim whitespace per HTML spec
			// textValue := strings.TrimSpace(text[0])
			if len(textContent) == 0 {
				return nil, nil
			}
			return Literal(textContent), nil
		}
	}

}

var HTMLElementOpenTagError = fmt.Errorf("error parsing HTML element opening tag")
var HTMLElementCloseTagError = fmt.Errorf("error parsing HTML element closing tag")
var HTMLElementChildrenError = fmt.Errorf("error parsing HTML element children")

type Macro struct {
	Value reflect.Value
}

func (m *Macro) Call(args []any) ([]any, error) {

	ft := m.Value.Type()

	var resultValues []reflect.Value

	if ft.IsVariadic() {
		// this is a variadic function
		// we check if we have enough arguments for the regular function arguments
		if len(args) < ft.NumIn() {
			return nil, fmt.Errorf("invalid number of regular arguments (expected at least %d, got %d)", ft.NumIn(), len(args))
		}
		argValues := make([]reflect.Value, ft.NumIn())
		// we check that all arguments are assignable
		for i := 0; i < ft.NumIn()-1; i++ {
			// we check the types for the non-variadic arguments
			argType := reflect.TypeOf(args[i])

			if !argType.AssignableTo(ft.In(i)) {
				return nil, fmt.Errorf("argument %d has the wrong type", i)
			}
			// we build the arguments list
			argValues[i] = reflect.ValueOf(args[i])
		}

		numVariadicArgs := len(args) - ft.NumIn() + 1

		variadicType := ft.In(ft.NumIn() - 1)
		variadicArgs := reflect.MakeSlice(variadicType, numVariadicArgs, numVariadicArgs)

		for i := 0; i < numVariadicArgs; i++ {
			variadicArgs.Index(i).Set(reflect.ValueOf(args[i+ft.NumIn()-1]))
		}
		argValues[ft.NumIn()-1] = variadicArgs
		resultValues = m.Value.CallSlice(argValues)

	} else {
		// this is a non-variadic function, number of arguments must match exactly
		if len(args) != ft.NumIn() {
			return nil, fmt.Errorf("invalid number of arguments (expected %d, got %d)", ft.NumIn(), len(args))
		}

		argValues := make([]reflect.Value, len(args))
		// we check that all arguments are assignable
		for i := 0; i < len(args); i++ {
			// we check the types
			argType := reflect.TypeOf(args[i])

			if !argType.AssignableTo(ft.In(i)) {
				return nil, fmt.Errorf("argument %d has the wrong type", i)
			}

			// we build the arguments list
			argValues[i] = reflect.ValueOf(args[i])
		}

		resultValues = m.Value.Call(argValues)
	}

	result := make([]any, len(resultValues))

	for i, resultValue := range resultValues {
		if !resultValue.CanInterface() {
			return nil, fmt.Errorf("invalid result value")
		}
		result[i] = resultValue.Interface()
	}

	return result, nil
}

var Macros = map[string]Macro{}

func RegisterMacro(name string, macro any) error {
	macroType := reflect.TypeOf(macro)

	if macroType.Kind() != reflect.Func {
		return fmt.Errorf("not a function")
	}

	Macros[name] = Macro{
		Value: reflect.ValueOf(macro),
	}

	return nil
}

func MustRegisterMacro(name string, macro any) {
	if err := RegisterMacro(name, macro); err != nil {
		panic(err)
	}
}

type Function struct {
	Name      string              `json:"name"`
	Children  []any               `json:"children" graph:"include"`
	Arguments []*FunctionArgument `json:"arguments" graph:"include"`
	Result    []any               `json:"result" graph:"include"`
}

func (f *Function) Generate(c Context) (any, error) {
	results := []any{}

	for _, result := range f.Result {
		if result == nil {
			continue
		}
		if generator, ok := result.(Generator); ok {
			if gr, err := generator.Generate(c); err != nil {
				return nil, err
			} else {
				results = append(results, gr)
			}
		} else {
			// we directly append the result
			results = append(results, result)
		}
	}

	return results, nil
}

func (f *Function) RenderCode() string {
	arguments := []string{}
	for _, argument := range f.Arguments {
		arguments = append(arguments, argument.RenderCode())
	}
	children := []string{}
	for _, child := range f.Children {
		if generator, ok := child.(Generator); !ok {
			// to do: handle this case
			continue
		} else {
			children = append(children, generator.RenderCode())
		}
	}
	return fmt.Sprintf("{:%[1]s %[2]s :}%[3]s{:/%[1]s:}", f.Name, strings.Join(arguments, " "), strings.Join(children, ""))
}

func (p *Parser) ParseFunctionBlock() (*Function, error) {

	if !p.has("{:", true) || p.has("{:/", true) {
		return nil, nil
	}

	success := false
	defer p.push(&success)()

	if err := p.consume("{:", true); err != nil {
		return nil, fmt.Errorf("%w - expected '{:': %w", HTMLElementOpenTagError, err)
	}

	f := &Function{}

	if name, err := p.consumeRegexp(tagRegexp); err != nil {
		return nil, fmt.Errorf("%w - expected a name: %w", HTMLElementOpenTagError, err)
	} else {
		f.Name = name[0]
	}

	if arguments, err := p.ParseFunctionArguments(); err != nil {
		return nil, fmt.Errorf("error parsing attributes: %w: %w", HTMLElementOpenTagError, err)
	} else {
		f.Arguments = arguments
	}

	// to do: handle self-closing tags
	if err := p.consume(":}", true); err != nil {
		return nil, fmt.Errorf("%w - expected ':}': %w", HTMLElementOpenTagError, err)
	}

	if children, err := p.ParseHTMLChildren(); err != nil {
		return nil, fmt.Errorf("%w - %w", HTMLElementChildrenError, err)
	} else {
		f.Children = children
	}

	// we consume the closing name
	if err := p.consume(fmt.Sprintf("{:/%s:}", f.Name), true); err != nil {
		return nil, fmt.Errorf("%w - %w", HTMLElementCloseTagError, err)
	}

	macro, ok := Macros[f.Name]

	if !ok {
		return nil, fmt.Errorf("unknown macro '%s'", f.Name)
	}

	argValues := []any{}

	for _, argument := range f.Arguments {
		argValues = append(argValues, argument.Value)
	}

	if result, err := macro.Call(append(argValues, f.Children...)); err != nil {
		return nil, fmt.Errorf("error executing macro '%s': %v", f.Name, err)
	} else {
		f.Result = result
	}

	success = true
	return f, nil

}

var simpleStringArgumentRegexp = regexp.MustCompile(`[^\s\>]+`)
var stringArgumentRegex = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)

type FunctionArgument struct {
	Value any
}

func (f *FunctionArgument) RenderCode() string {
	if strValue, ok := f.Value.(string); ok {
		return strValue
	} else if generator, ok := f.Value.(Generator); ok {
		return generator.RenderCode()
	}
	// to do: handle this case
	return ""
}

func (p *Parser) ParseFunctionArguments() ([]*FunctionArgument, error) {

	success := false
	defer p.push(&success)()

	arguments := make([]*FunctionArgument, 0, 10)

	for {

		a := &FunctionArgument{}

		if p.has(":}", true) || p.has("}", true) {
			if err := p.consumeWhitespace(); err != nil {
				return nil, err
			}
			// we have reached the end
			success = true
			return arguments, nil
		}

		if err := p.consumeWhitespace(); err != nil {
			return nil, err
		}

		if p.has("\"", false) {
			// this is a string
			if string, err := p.consumeRegexp(stringArgumentRegex); err != nil {
				return nil, err
			} else {
				a.Value = string[1]
			}
		} else {
			// this is a simple string
			if value, err := p.consumeRegexp(simpleStringArgumentRegexp); err != nil {
				return nil, fmt.Errorf("expected a simple argument value: %w", err)
			} else {
				a.Value = value[0]
			}
		}

		arguments = append(arguments, a)
	}

	return nil, nil
}

func (p *Parser) ParseHTMLElement() (*HTMLElement, error) {

	if !p.has("<", true) || p.has("</", true) {
		return nil, nil
	}

	success := false
	defer p.push(&success)()

	if err := p.consume("<", true); err != nil {
		return nil, fmt.Errorf("%w - expected '<': %w", HTMLElementOpenTagError, err)
	}

	e := &HTMLElement{}

	if tag, err := p.consumeRegexp(tagRegexp); err != nil {
		return nil, fmt.Errorf("%w - expected a tag: %w", HTMLElementOpenTagError, err)
	} else {
		e.Tag = tag[0]
	}

	if attributes, err := p.ParseHTMLAttributes(); err != nil {
		return nil, fmt.Errorf("%w: %w", HTMLElementOpenTagError, err)
	} else {
		e.Attributes = attributes
	}

	if err := p.consume("/>", true); err == nil {
		// this is an explicitly closed tag, we don't expect any children
		return e, nil
	}

	// to do: handle self-closing tags
	if err := p.consume(">", true); err != nil {
		return nil, fmt.Errorf("%w - expected '>': %w", HTMLElementOpenTagError, err)
	}

	if e.Tag == "meta" || e.Tag == "img" {
		// this is a self-closing tag...
		return e, nil
	}

	if children, err := p.ParseHTMLChildren(); err != nil {
		return nil, fmt.Errorf("%w - %w", HTMLElementChildrenError, err)
	} else {
		e.Children = children
	}

	// we consume the closing tag
	if err := p.consume(fmt.Sprintf("</%s>", e.Tag), true); err != nil {
		return nil, fmt.Errorf("%w - %w", HTMLElementCloseTagError, err)
	}

	success = true
	return e, nil
}

var attributeNameRegexp = regexp.MustCompile(`[a-zA-Z][a-zA-Z\-]*`)
var simpleStringAttributeRegexp = regexp.MustCompile(`[^\s\>]+`)
var stringAttributeRegex = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)

func (p *Parser) ParseHTMLAttributes() ([]*HTMLAttribute, error) {

	success := false
	defer p.push(&success)()

	attributes := make([]*HTMLAttribute, 0, 10)

	for {

		a := &HTMLAttribute{}

		if p.has(">", true) || p.has("/>", true) {
			if err := p.consumeWhitespace(); err != nil {
				return nil, err
			}
			// we have reached the end
			success = true
			return attributes, nil
		}

		if err := p.consumeWhitespace(); err != nil {
			return nil, err
		}

		if name, err := p.consumeRegexp(attributeNameRegexp); err != nil {
			return nil, fmt.Errorf("expected an attribute name: %w", err)
		} else {
			a.Name = name[0]
		}

		if err := p.consume("=", false); err != nil {
			return nil, fmt.Errorf("expected '=': %w", err)
		}

		if p.has("\"", false) {
			// this is a string
			if string, err := p.consumeRegexp(stringAttributeRegex); err != nil {
				return nil, err
			} else {
				a.Value = string[1]
			}
		} else if p.has("{", false) {
			if value, err := p.ParseExpression("}"); err != nil {
				return nil, fmt.Errorf("cannot parse expression: %w", err)
			} else {
				a.Value = value
			}
			// this is an expression
		} else {
			// this is a simple string
			if value, err := p.consumeRegexp(simpleStringAttributeRegexp); err != nil {
				return nil, fmt.Errorf("expected a simple attribute value: %w", err)
			} else {
				a.Value = value[0]
			}
		}

		attributes = append(attributes, a)
	}

	return nil, nil
}

func (p *Parser) ParseCSSStatement() (*CSSStatement, error) {
	return nil, nil
}
