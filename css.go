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

func MakeStylesheet(name string, args ...any) *Stylesheet {
	ss := &Stylesheet{
		Name:       name,
		classIndex: 1,
	}

	ss.Rules = subrules(ss, args)

	return ss
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

func Direct(tagMaker func(makerArgs ...any) *HTMLElement) func(args ...any) *Rule {
	return func(args ...any) *Rule {
		element := tagMaker(args...)
		return MakeRule(nil, &TagSelector{TagName: element.Tag, Direct: true}, element.Args...)
	}
}

func subrules(parent Ruleset, args []any) []*Rule {
	// we filter out rules directly
	srs := filter[*Rule](args)

	// we convert HTML elements into tag rules
	elements := filter[*HTMLElement](args)

	for _, element := range elements {
		srs = append(srs, MakeRule(parent, &TagSelector{TagName: element.Tag, Direct: false}, element.Args...))
	}

	attributes := filter[*HTMLAttribute](args)
	for _, attribute := range attributes {
		if attribute.Name == "class" {
			strValue, ok := attribute.Value.(string)
			if !ok {
				// to do: properly handle this
				continue
			}
			srs = append(srs, MakeRule(parent, &ClassSelector{ClassName: strValue, Source: attribute}, attribute.Args...))
		}
	}

	// we filter out lists of arguments
	lists := filter[[]any](args)

	for _, list := range lists {
		srs = append(srs, subrules(parent, list)...)
	}

	return srs
}

func declarations(args []any) []*Declaration {
	decs := make([]*Declaration, 0, len(args))

	for _, arg := range args {
		switch vt := arg.(type) {
		case *Declaration:
			if vt == nil {
				continue
			}
			decs = append(decs, vt)
		case []any:
			decs = append(decs, declarations(vt)...)
		}
	}

	return decs
}

func MakeRule(parent Ruleset, selector Selector, args ...any) *Rule {

	subrules := subrules(parent, args)

	rule := &Rule{
		Parent:       parent,
		Selector:     selector,
		Declarations: declarations(args),
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
	Direct  bool
}

func (s *TagSelector) String() string {
	prefix := ""
	if s.Direct {
		prefix = " >"
	}
	return fmt.Sprintf("%s %s", prefix, s.TagName)
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
	r.Declarations = append(r.Declarations, filter[*Declaration](args)...)
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

func filter[T any](args []any) []T {
	ts := make([]T, 0, len(args))

	for _, arg := range args {
		if va, ok := arg.(T); ok {
			ts = append(ts, va)
		}
	}

	return ts
}

// Declarations

func Dec(property string) func(value any) *Declaration {
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
		return MakeRule(nil, &TagSelector{TagName: tagName, Direct: false}, args...)
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

func rules(args []any) []*Rule {
	rules := make([]*Rule, 0)
	ruleArgs := make([]any, 0)
	for _, arg := range args {

		switch vt := arg.(type) {
		case *Rule:

			// we do the nil comparison here as it's a bit complicated:
			// https://go.dev/doc/faq#nil_error
			if vt == nil {
				continue
			}

			if className := vt.Class(); className == "" {
				// we append the rule to the rule args
				ruleArgs = append(ruleArgs, vt)
			} else {
				// this is a class rule, we can directly inclcude
				rules = append(rules, vt)
			}
		case *HTMLElement:

			if vt == nil {
				continue
			}

			ruleArgs = append(ruleArgs, vt)
		case *Declaration:

			if vt == nil {
				continue
			}

			ruleArgs = append(ruleArgs, vt)
		case []any:
			stylesheet := Styles(vt...)
			rules = append(rules, stylesheet.Rules...)
		}
	}

	if len(ruleArgs) > 0 {
		rules = append(rules, MakeRule(nil, nil, ruleArgs...))
	}

	return rules
}

// Convert a list of styles into a list of classes
func Styles(args ...any) *StylesStruct {
	return &StylesStruct{
		Rules: rules(args),
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

var TextDecoration = Dec("text-decoration")
var TextTransform = Dec("text-transform")
var TextAlign = Dec("text-align")

// Box Shadow

var BoxShadow = Dec("box-shadow")

// Colors
var Color = Dec("color")

// Background
var Background = Dec("background")
var BackgroundColor = Dec("background-color")
var BackgroundImage = Dec("background-image")
var BackgroundSize = Dec("background-size")
var BackgroundPosition = Dec("background-position")

// Borders
var BorderRadius = Dec("border-radius")
var BorderWidth = Dec("border-width")
var BorderStyle = Dec("border-style")
var BorderColor = Dec("border-color")
var Border = Dec("border")
var BorderBottom = Dec("border-bottom")
var BorderTop = Dec("border-top")
var BorderLeft = Dec("border-left")
var BorderRight = Dec("border-right")

// Padding
var Padding = Dec("padding")
var PaddingTop = Dec("padding-top")
var PaddingBottom = Dec("padding-bottom")
var PaddingLeft = Dec("padding-left")
var PaddingRight = Dec("padding-right")

// Margin
var Margin = Dec("margin")
var MarginTop = Dec("margin-top")
var MarginBottom = Dec("margin-bottom")
var MarginLeft = Dec("margin-left")
var MarginRight = Dec("margin-right")

// Transform
var Transform = Dec("transform")
var TransformOrigin = Dec("transform-origin")

// Dimensions

var Width = Dec("width")
var MinWidth = Dec("min-width")
var MaxWidth = Dec("max-width")
var Height = Dec("height")
var MinHeight = Dec("min-height")
var MaxHeight = Dec("max-height")

// Positioning

var Position = Dec("position")
var Left = Dec("left")
var Top = Dec("top")
var Bottom = Dec("bottom")

// Display

var Display = Dec("display")

// Anything

var Anything = Dec("*")

// Flexbox

var FlexShrink = Dec("flex-shrink")
var FlexGrow = Dec("flex-grow")
var FlexBasis = Dec("flex-basis")
var FlexDirection = Dec("flex-direction")
var AlignItems = Dec("align-items")
var JustifyContent = Dec("justify-content")

// Fonts

var FontSize = Dec("font-size")
var FontFamily = Dec("font-family")
var FontWeight = Dec("font-weight")
var FontStretch = Dec("font-stretch")
var LetterSpacing = Dec("letter-spacing")
var LineHeight = Dec("line-height")

// Lists

var ListStyle = Dec("list-style")

// Opacity etc.

var Opacity = Dec("opacity")

// Box Sizing

var BoxSizing = Dec("box-sizing")

// SVG

var Stroke = Dec("stroke")
var Fill = Dec("fill")
var StrokeWidth = Dec("stroke-width")

// Transitions

var Transition = Dec("transition")

// Filtering

var Filter = Dec("filter")

// Grid

var GridTemplateColumns = Dec("grid-template-columns")
var GridAutoRows = Dec("grid-auto-rows")
var GridGap = Dec("grid-gap")
var JustifyItems = Dec("justify-items")
