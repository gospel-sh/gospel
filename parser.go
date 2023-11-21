package gospel

import (
	"fmt"
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

	if statements, err := p.parseStatements(); err != nil {
		return nil, err
	} else {
		return &Program{
			Statements: statements,
		}, nil
	}
}

func (p *Parser) parseStatements() ([]*Statement, error) {
	statements := make([]*Statement, 0, 1)

	for {
		statement, err := p.parseStatement()

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

func (p *Parser) parseStatement() (*Statement, error) {
	if p.has("html ", true) {
		// this is a HTML statement
		if htmlStatement, err := p.parseHTMLStatement(); err != nil {
			return nil, err
		} else {
			return &Statement{
				Type: HTMLStmt,
				HTML: htmlStatement,
			}, nil
		}
	} else if p.has("css ", true) {
		// this is a CSS statement
		if cssStatement, err := p.parseCSSStatement(); err != nil {
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

func (p *Parser) parseHTMLStatement() (*HTMLStatement, error) {

	// we consume the HTML keyword
	p.consume("html", true)

	if err := p.consume(" ", false); err != nil {
		// we expect at least one whitespace character
		return nil, fmt.Errorf("expected a whitespace after the 'html' keyword: %w", err)
	}

	if p.has("template", false) {
		if template, err := p.parseTemplate(); err != nil {
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

func (p *Parser) parseTemplate() (*HTMLTemplate, error) {

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

	if htmlElement, err := p.parseHTMLElement(); err != nil {
		return nil, fmt.Errorf("error parsing HTML element: %w", err)
	} else if htmlElement == nil {
		return nil, fmt.Errorf("expected a HTML element")
	} else {
		t.Element = htmlElement
	}

	return t, nil
}

var tagRegexp = regexp.MustCompile(`[a-z][a-z\-]*`)

func (p *Parser) parseHTMLChildren() ([]any, error) {

	success := false
	defer p.push(&success)()

	elements := make([]any, 0, 10)

	for {

		if element, err := p.parseHTMLElement(); err != nil {
			return nil, err
		} else if element != nil {
			elements = append(elements, element)
			continue
		}

		if element, err := p.parseHTMLTextNode(); err != nil {
			return nil, err
		} else if element != nil {
			elements = append(elements, element)
			continue
		}

		// no more elements
		break
	}

	success = true
	return elements, nil
}

var htmlTextNodeRegexp = regexp.MustCompile(`(?ms)[^<]*`)

func (p *Parser) parseHTMLTextNode() (*HTMLElement, error) {
	if text, err := p.consumeRegexp(htmlTextNodeRegexp); err != nil {
		return nil, fmt.Errorf("not a text node: %w", err)
	} else {
		// we can trim whitespace per HTML spec
		textValue := strings.TrimSpace(text[0])
		if len(textValue) == 0 {
			return nil, nil
		}
		return Literal(textValue), nil
	}

}

var HTMLElementOpenTagError = fmt.Errorf("error parsing HTML element opening tag")
var HTMLElementCloseTagError = fmt.Errorf("error parsing HTML element closing tag")
var HTMLElementChildrenError = fmt.Errorf("error parsing HTML element children")

func (p *Parser) parseHTMLElement() (*HTMLElement, error) {

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

	if attributes, err := p.parseHTMLAttributes(); err != nil {
		return nil, fmt.Errorf("%w: %w", HTMLElementOpenTagError, err)
	} else {
		e.Attributes = attributes
	}

	// to do: handle self-closing tags
	if err := p.consume(">", true); err != nil {
		return nil, fmt.Errorf("%w - expected '>': %w", HTMLElementOpenTagError, err)
	}

	if children, err := p.parseHTMLChildren(); err != nil {
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

func (p *Parser) parseHTMLAttributes() ([]*HTMLAttribute, error) {

	success := false
	defer p.push(&success)()

	attributes := make([]*HTMLAttribute, 0, 10)

	for {

		a := &HTMLAttribute{}

		if p.has(">", true) {
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
			if value, err := p.parseExpression("}"); err != nil {
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

func (p *Parser) parseCSSStatement() (*CSSStatement, error) {
	return nil, nil
}
