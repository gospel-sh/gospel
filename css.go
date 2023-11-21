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
	Name       string
	Rules      []*Rule
}

func MakeStylesheet(name string, rules ...*Rule) *Stylesheet {
	return &Stylesheet{
		Name:       name,
		Rules:      rules,
		classIndex: 1,
	}
}

func Styled(name string, element Element) Element {

	css := MakeStylesheet(name)

	addStyles := func(styles *StylesStruct, element *HTMLElement) {
		classNames := make([]string, 0)
		for _, rule := range styles.Rules {
			if rule.Selector == nil {
				// we add a selector to the rule
				rule.Selector = css.ClassSelector()
			}
			css.AddRule(rule)
			// we add the rule class to the class names
			classNames = append(classNames, rule.Class())
		}
		element.Attributes = append(element.Attributes, Class(strings.Join(classNames, " ")))
	}

	// we add all rules to the stylesheet
	Walk(element, addStyles)

	return F(
		css.Styles(),
		element,
	)
}

// Returns a link to the stylesheet as well as a route that return the styles
func (s *Stylesheet) Link() Element {
	return nil
}

// Returns the styles in a <style> tag
func (s *Stylesheet) Styles() Element {

	ss := s.String()

	if ss == "" {
		return nil
	}

	return StyleTag(
		L(ss),
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

func (s *Stylesheet) ClassSelector() *ClassSelector {
	className := fmt.Sprintf("%s-%d", s.Name, s.classIndex)
	s.classIndex++
	return &ClassSelector{ClassName: className}
}

func (s *Stylesheet) Fragment(args ...any) *Rule {
	className := fmt.Sprintf("%s-%d", s.Name, s.classIndex)
	s.classIndex++
	return s.NamedFragment(className, args...)
}

func (s *Stylesheet) NamedFragment(name string, args ...any) *Rule {
	rule := MakeRule(s, &ClassSelector{ClassName: name}, args...)
	return rule

}

func (s *Stylesheet) AddRule(rule *Rule) {
	// we check if the rule already is in the stylesheet
	for _, existingRule := range s.Rules {
		if existingRule == rule {
			return
		}
	}

	rule.Parent = s
	s.Rules = append(s.Rules, rule)
}

func (s *Stylesheet) AddStylesheet(stylesheet *Stylesheet) {
	for _, rule := range stylesheet.Rules {
		rule.Parent = s
		s.Rules = append(s.Rules, rule)
	}
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

	flatRules := Flatten(s.Rules, nil, nil)

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

		if rule.MediaQueries != nil {
			// we append the rule media query (if it exists) to the existing ones
			ruleMediaQueries = append(ruleMediaQueries, rule.MediaQueries...)
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

type PseudoClassSelector struct {
	Type          string
	Args          []any
	ArgsFormatter func(args []any) string
}

func (s *PseudoClassSelector) String() string {

	if s.ArgsFormatter != nil {
		return fmt.Sprintf(":%s(%s)", s.Type, s.ArgsFormatter(s.Args))
	}

	return fmt.Sprintf(":%s", s.Type)
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
	MediaQueries []*MediaQuery
	Declarations []*Declaration
	Subrules     []*Rule
}

func (r *Rule) Class() string {

	classSelector, ok := r.Selector.(*ClassSelector)

	if !ok {
		return ""
	}

	return classSelector.ClassName
}

func (r *Rule) Derive(args ...any) *Rule {

	declarations := make([]*Declaration, len(r.Declarations))
	copy(declarations, r.Declarations)

	subrules := make([]*Rule, len(r.Subrules))
	copy(subrules, r.Subrules)

	mediaQueries := make([]*MediaQuery, len(r.MediaQueries))
	copy(mediaQueries, r.MediaQueries)

	// we make a copy of the current rule
	nr := &Rule{
		Parent:       r.Parent,
		Selector:     r.Selector,
		MediaQueries: mediaQueries,
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

func (s *Size) Sub(value float64) *Size {
	return &Size{
		Unit:  s.Unit,
		Value: s.Value - value,
	}
}

func (s *Size) Add(value float64) *Size {
	return &Size{
		Unit:  s.Unit,
		Value: s.Value + value,
	}
}

func (s *Size) Mul(value float64) *Size {
	return &Size{
		Unit:  s.Unit,
		Value: s.Value * value,
	}
}

func (s *Size) Div(value float64) *Size {
	return &Size{
		Unit:  s.Unit,
		Value: s.Value / value,
	}
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

// Class Rule

func ClassRule(className string, args ...any) *Rule {
	return MakeRule(nil, &ClassSelector{ClassName: className}, args...)
}

// Tags

func TagRule(tagName string) func(args ...any) *Rule {
	return func(args ...any) *Rule {
		return MakeRule(nil, &TagSelector{tagName}, args...)
	}
}

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
var Division = op("/")

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
var TabletWidth = Px(1024)
var DesktopWidth = Px(1280)

func MakeMediaQuery(name string, arg any) *MediaQuery {
	return &MediaQuery{
		name, arg,
	}
}

func mediaQueryRule(queries []*MediaQuery) func(args ...any) *Rule {
	return func(args ...any) *Rule {
		// we create a rule with the given arguments
		rule := MakeRule(nil, nil, args...)
		rule.MediaQueries = append(rule.MediaQueries, queries...)
		return rule
	}
}

var Mobile = mediaQueryRule([]*MediaQuery{MakeMediaQuery("max-width", MobileWidth)})
var Tablet = mediaQueryRule([]*MediaQuery{MakeMediaQuery("min-width", TabletWidth), MakeMediaQuery("max-width", DesktopWidth.Sub(1))})
var Desktop = mediaQueryRule([]*MediaQuery{MakeMediaQuery("min-width", DesktopWidth)})

// Helper functions

type StylesStruct struct {
	Rules []*Rule
}

// Convert a list of styles into a list of classes
func Styles(args ...any) *StylesStruct {
	rules := make([]*Rule, 0)
	ruleArgs := make([]any, 0)
	for _, arg := range args {

		switch vt := arg.(type) {
		case *Rule:
			if className := vt.Class(); className == "" {
				// we append the rule to the rule args
				ruleArgs = append(ruleArgs, vt)
			} else {
				// this is a class rule, we can directly inclcude
				rules = append(rules, vt)
			}
		case *HTMLElement:
			ruleArgs = append(ruleArgs, vt)
		case *Declaration:
			ruleArgs = append(ruleArgs, vt)
		}
	}

	if len(ruleArgs) > 0 {
		rules = append(rules, MakeRule(nil, nil, ruleArgs...))
	}

	return &StylesStruct{
		Rules: rules,
	}
}

// Any selector

var Any = Tag("*")

// Pseudo classes

func pseudoArgs(typeName string, formatter func(args []any) string) func(args ...any) func(args ...any) *Rule {
	return func(pseudoArgs ...any) func(args ...any) *Rule {
		return func(args ...any) *Rule {
			return MakeRule(nil, &PseudoClassSelector{Type: typeName, ArgsFormatter: formatter, Args: pseudoArgs}, args...)
		}
	}
}

func pseudo(typeName string) func(args ...any) *Rule {
	return func(args ...any) *Rule {
		return MakeRule(nil, &PseudoClassSelector{Type: typeName}, args...)
	}
}

var Active = pseudo("active")
var AnyLink = pseudo("any-link")
var Autofill = pseudo("autofill")
var Blank = pseudo("blank")
var Buffering = pseudo("buffering")
var Checked = pseudo("checked")
var Current = pseudoArgs("current", nil)
var Default = pseudo("default")
var Defined = pseudo("defined")
var Dir = pseudoArgs("dir", nil)
var Disabled = pseudo("disabled")
var Empty = pseudo("empty")
var Enabled = pseudo("enabled")
var FirstChild = pseudo("first-child")
var FirstOfType = pseudo("first-of-type")
var Focus = pseudo("focus")
var FocusVisible = pseudo("focus-visible")
var FocusWithin = pseudo("focus-within")
var Fullscreen = pseudo("fullscreen")
var Future = pseudo("future")
var Has = pseudoArgs("has", nil)
var Hover = pseudo("hover")
var Indeterminate = pseudo("indeterminate")

// ...
var LastChild = pseudo("last-child")

// Declarations

// Text

var TextDecoration = dec("text-decoration")
var TextTransform = dec("text-transform")

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
var BorderBottom = dec("border-bottom")
var BorderTop = dec("border-top")
var BorderLeft = dec("border-left")
var BorderRight = dec("border-right")

// Padding
var Padding = dec("padding")
var PaddingTop = dec("padding-top")
var PaddingBottom = dec("padding-bottom")
var PaddingLeft = dec("padding-left")
var PaddingRight = dec("padding-right")

// Margin
var Margin = dec("margin")
var MarginTop = dec("margin-top")
var MarginBottom = dec("margin-bottom")
var MarginLeft = dec("margin-left")
var MarginRight = dec("margin-right")

// Transform
var Transform = dec("transform")
var TransformOrigin = dec("transform-origin")

// Dimensions

var Width = dec("width")
var MinWidth = dec("min-width")
var MaxWidth = dec("max-width")
var Height = dec("height")
var MinHeight = dec("min-height")
var MaxHeight = dec("max-height")

// Positioning

var Position = dec("position")
var Left = dec("left")
var Top = dec("top")

// Display

var Display = dec("display")

// Anything

var Anything = dec("*")

// Flexbox

var FlexShrink = dec("flex-shrink")
var FlexGrow = dec("flex-grow")
var FlexBasis = dec("flex-basis")
var FlexDirection = dec("flex-direction")
var AlignItems = dec("align-items")
var JustifyContent = dec("justify-content")

// Fonts

var FontSize = dec("font-size")
var FontFamily = dec("font-family")
var FontWeight = dec("font-weight")
var FontStretch = dec("font-stretch")
var LetterSpacing = dec("letter-spacing")
var LineHeight = dec("line-height")

// Lists

var ListStyle = dec("list-style")

// Opacity etc.

var Opacity = dec("opacity")
