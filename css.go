package gospel

import (
	"fmt"
	"sort"
	"strings"
)

type Ruleset interface{}

// A stylesheet contains a number of rulesets
type Stylesheet struct {
	classIndex int
	Rules      []*Rule
}

func MakeStylesheet() *Stylesheet {
	return &Stylesheet{
		Rules:      make([]*Rule, 0),
		classIndex: 1,
	}
}

// Returns a link to the stylesheet as well as a route that return the styles
func (s *Stylesheet) Link() Element {
	return nil
}

// Returns the styles in a <style> tag
func (s *Stylesheet) Styles() Element {
	return StyleTag(
		L(s.String()),
	)
}

func subrules(parent Ruleset, args []any) []*Rule {
	// we filter out rules directly
	subrules := filter[Rule](args)

	// we convert HTML elements into tag rules
	elements := filter[HTMLElement](args)

	for _, element := range elements {
		subrules = append(subrules, MakeRule(parent, &TagSelector{element.Tag}, element.Args...))
	}

	attributes := filter[HTMLAttribute](args)
	for _, attribute := range attributes {
		if attribute.Name == "class" {
			strValue, ok := attribute.Value.(string)
			if !ok {
				// to do: properly handle this
				continue
			}
			subrules = append(subrules, MakeRule(parent, &ClassSelector{ClassName: strValue, Source: attribute}, attribute.Args...))
		}
	}

	return subrules
}

func MakeRule(parent Ruleset, selector Selector, args ...any) *Rule {

	subrules := subrules(parent, args)

	rule := &Rule{
		Parent:       parent,
		Selector:     selector,
		Declarations: filter[Declaration](args),
		Subrules:     subrules,
	}

	for _, subrule := range subrules {
		subrule.Parent = rule
	}

	return rule

}

func (s *Stylesheet) Fragment(args ...any) *Rule {
	className := fmt.Sprintf("gospel-%d", s.classIndex)
	s.classIndex++

	return s.NamedFragment(className, args...)
}

func (s *Stylesheet) NamedFragment(name string, args ...any) *Rule {
	rule := MakeRule(s, &ClassSelector{ClassName: name}, args...)
	return rule

}

func (s *Stylesheet) Rule(args ...any) *Rule {
	rule := s.Fragment(args...)
	s.Rules = append(s.Rules, rule)
	return rule
}

func (s *Stylesheet) NamedRule(name string, args ...any) *Rule {
	rule := s.NamedFragment(name, args...)
	s.Rules = append(s.Rules, rule)
	return rule
}

func (s *Stylesheet) String() string {

	filteredRules := make([]*Rule, 0, len(s.Rules))

	for _, rule := range s.Rules {
		if rule.Parent != s {
			continue
		}
		filteredRules = append(filteredRules, rule)
	}

	flatRules := Flatten(filteredRules, nil, nil)

	css := ""

	for _, flatRule := range flatRules {

		if len(flatRule.Declarations) == 0 {
			// this rule is empty
			continue
		}

		css += flatRule.String() + "\n"
	}

	return css
}

func Flatten(rules []*Rule, selectors []Selector, mediaQueries []*MediaQuery) []*FlatRule {
	flatRules := make([]*FlatRule, 0, len(rules))

	for _, rule := range rules {

		ruleSelectors := make([]Selector, len(selectors))
		copy(ruleSelectors, selectors)

		ruleMediaQueries := make([]*MediaQuery, len(mediaQueries))
		copy(ruleMediaQueries, mediaQueries)

		if rule.Selector != nil {
			// we append the rule selector to the existing ones
			ruleSelectors = append(ruleSelectors, rule.Selector)
		}

		if rule.MediaQuery != nil {
			// we append the rule media query (if it exists) to the existing ones
			ruleMediaQueries = append(ruleMediaQueries, rule.MediaQuery)
		}

		flatRule := &FlatRule{
			Selectors:    ruleSelectors,
			MediaQueries: ruleMediaQueries,
			Declarations: rule.Declarations,
		}

		// we append the flattened role to our list
		flatRules = append(flatRules, flatRule)

		if len(rule.Subrules) > 0 {
			// we recurse into the subrules and repeat the flattening process
			flatRules = append(flatRules, Flatten(rule.Subrules, ruleSelectors, ruleMediaQueries)...)
		}
	}

	return flatRules
}

// Selectors

type Selector interface {
	String() string
}

type ClassSelector struct {
	ClassName string
	Source    *HTMLAttribute
}

func (s *ClassSelector) String() string {
	return fmt.Sprintf(".%s", s.ClassName)
}

type TagSelector struct {
	TagName string
}

func (s *TagSelector) String() string {
	return fmt.Sprintf(" %s", s.TagName)
}

type MediaQuery struct {
	Name  string
	Value any
}

func (m *MediaQuery) String() string {
	return fmt.Sprintf("%s: %v", m.Name, m.Value)
}

// A FlatRule is a "real" CSS rule that can easily be rendered
type FlatRule struct {
	Selectors    []Selector
	MediaQueries []*MediaQuery
	Declarations []*Declaration
}

func indent(s string, n int) string {

	var indented string

	lines := strings.Split(s, "\n")
	prefix := strings.Repeat(" ", n)

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		indented += prefix + line + "\n"
	}

	return indented
}

func deduplicateDeclarations(declarations []*Declaration) []*Declaration {
	dm := map[string]*Declaration{}
	keys := make([]string, 0, len(declarations))

	for _, declaration := range declarations {

		if _, ok := dm[declaration.Property]; !ok {
			keys = append(keys, declaration.Property)
		}

		// we store the declaration in the map
		// the last declaration in the list wins
		dm[declaration.Property] = declaration
	}

	// we sort the declarations by property name and return them
	dd := make([]*Declaration, 0, len(declarations))
	sort.Strings(keys)

	for _, v := range keys {
		dd = append(dd, dm[v])
	}

	return dd
}

func (f *FlatRule) String() string {
	var selectors, mediaQueries, declarations []string

	for _, mediaQuery := range f.MediaQueries {
		mediaQueries = append(mediaQueries, fmt.Sprintf("(%s)", mediaQuery.String()))
	}

	for _, selector := range f.Selectors {
		selectors = append(selectors, selector.String())
	}

	for _, declaration := range deduplicateDeclarations(f.Declarations) {
		declarations = append(declarations, declaration.String())
	}

	rule := fmt.Sprintf("%s {\n%s}\n", strings.Trim(strings.Join(selectors, ""), " "), indent(strings.Join(declarations, "\n"), 2))

	if len(mediaQueries) > 0 {
		return fmt.Sprintf("@media %s {\n%s}\n", strings.Join(mediaQueries, " and "), indent(rule, 2))
	}

	return rule
}

// A Rule contains a number of declarations and potentially subrules
type Rule struct {
	Parent       Ruleset
	Selector     Selector
	MediaQuery   *MediaQuery
	Declarations []*Declaration
	Subrules     []*Rule
}

func (r *Rule) Class() string {

	classSelector, ok := r.Selector.(*ClassSelector)

	if !ok {
		// to do: better error handling
		Log.Warning("Not a class-based rule")
		return ""
	}

	return classSelector.ClassName
}

func (r *Rule) Derive(args ...any) *Rule {

	declarations := make([]*Declaration, len(r.Declarations))
	copy(declarations, r.Declarations)

	subrules := make([]*Rule, len(r.Subrules))
	copy(subrules, r.Subrules)

	// we make a copy of the current rule
	nr := &Rule{
		Parent:       r.Parent,
		Selector:     r.Selector,
		MediaQuery:   r.MediaQuery,
		Declarations: declarations,
		Subrules:     subrules,
	}

	// we extend the rule with the args
	nr.Extend(args...)
	return nr

}

func (r *Rule) Extend(args ...any) {

	subrules := subrules(r, args)

	for _, subrule := range subrules {
		subrule.Parent = r
	}

	// we extend the rule with the new subrules and declarations
	r.Declarations = append(r.Declarations, filter[Declaration](args)...)
	r.Subrules = append(r.Subrules, subrules...)

}

// A declaration maps a value to a property. A value can be either a
// literal like a string, or a variable
type Declaration struct {
	Property string `json:"property"`
	Value    any    `json:"value"`
}

func (d *Declaration) String() string {
	return fmt.Sprintf("%s: %v;", d.Property, d.Value)
}

// A variable resolves to a string
type Variable interface {
	Value() string
}

type Size struct {
	Unit  string
	Value float64
}

// helpers

func filter[T any](args []any) []*T {
	ts := make([]*T, 0, len(args))

	for _, arg := range args {
		if va, ok := arg.(*T); ok {
			ts = append(ts, va)
		}
	}

	return ts
}

// Declarations

func dec(property string) func(value any) *Declaration {
	return func(value any) *Declaration {
		return &Declaration{
			Property: property,
			Value:    value,
		}
	}
}

// Tags

func TagRule(tagName string) func(args ...any) *Rule {
	return func(args ...any) *Rule {
		return MakeRule(nil, &TagSelector{tagName}, args...)
	}
}

// Text Decoration

var TextDecoration = dec("text-decoration")

// Colors
var Color = dec("color")

// Background
var Background = dec("background")
var BackgroundColor = dec("background-color")

// Borders
var BorderRadius = dec("border-radius")
var BorderWidth = dec("border-width")
var BorderStyle = dec("border-style")
var BorderColor = dec("border-color")
var Border = dec("border")

// Padding
var Padding = dec("padding")

// Margin
var Margin = dec("margin")

// Transform
var Transform = dec("transform")
var TransformOrigin = dec("transform-origin")

// Dimensions

var Width = dec("width")
var Height = dec("height")

// Positioning

var Position = dec("position")
var Left = dec("left")
var Top = dec("top")

// Functions

type CSSFunc struct {
	Name string
	Args []any
}

func (c *CSSFunc) String() string {

	strArgs := make([]string, len(c.Args))

	for i, arg := range c.Args {
		strArgs[i] = fmt.Sprintf("%v", arg)
	}

	return fmt.Sprintf("%s(%s)", c.Name, strings.Join(strArgs, ", "))
}

func fnc(name string) func(args ...any) *CSSFunc {
	return func(args ...any) *CSSFunc {
		return &CSSFunc{
			Name: name,
			Args: args,
		}
	}
}

type Op struct {
	Op   string
	Args []any
}

func (o *Op) String() string {

	strArgs := make([]string, len(o.Args))

	for i, arg := range o.Args {
		strArgs[i] = fmt.Sprintf("%v", arg)
	}

	return strings.Join(strArgs, " "+o.Op+" ")

}

func op(name string) func(args ...any) *Op {
	return func(args ...any) *Op {
		return &Op{
			Op:   name,
			Args: args,
		}
	}
}

// Operators

var Add = op("+")
var Sub = op("-")
var Mul = op("*")
var Dir = op("/")

// Calc

var Calc = fnc("calc")
var Scale = fnc("scale")

// Sizes

func size(name string) func(value float64) *Size {
	return func(value float64) *Size {
		return &Size{
			name,
			value,
		}
	}
}

var Px = size("px")
var Percent = size("%")
var Em = size("em")
var Rem = size("rem")
var Vh = size("vh")
var Vw = size("vw")

func (p *Size) String() string {
	return fmt.Sprintf("%v%s", p.Value, p.Unit)
}

// Media Queries

var MobileWidth = Px(600)

// Returns a new rule with a media query for the mobile width
func Mobile(args ...any) *Rule {
	// we create a rule with the given arguments
	rule := MakeRule(nil, nil, args...)
	// we add the media query to the rule
	rule.MediaQuery = &MediaQuery{
		"max-width",
		MobileWidth,
	}

	return rule
}

// Helper functions

func Styles(args ...any) []any {
	classes := make([]string, 0)

	for _, arg := range args {

		switch vt := arg.(type) {
		case *Rule:
			if className := vt.Class(); className == "" {
				// to do: handle this
				continue
			} else {
				classes = append(classes, className)
			}

		}

	}

	return []any{Class(strings.Join(classes, " "))}
}

// Any selector

var Any = Tag("*")
