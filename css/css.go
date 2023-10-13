package css

import (
	"github.com/gospel-sh/gospel"
)

type Ruleset interface {
	Rule(args ...any) *Rule
}

// A stylesheet contains a number of rulesets
type Stylesheet struct {
	Rules []*Rule
}

func MakeStylesheet() *Stylesheet {
	return &Stylesheet{
		Rules: make([]*Rule, 0),
	}
}

// Returns a link to the stylesheet as well as a route that return the styles
func (s *Stylesheet) Link() gospel.Element {
	return nil
}

// Returns the styles in a <style> tag
func (s *Stylesheet) Styles() gospel.Element {

	for _, rule := range s.Rules {
		gospel.Log.Info("Got rule: %v", rule)
	}

	return nil
}

func (s *Stylesheet) Rule(args ...any) *Rule {
	rule := &Rule{
		Parent: s,
		Declarations: filter[Declaration](args),
		Subrules: filter[Rule](args),
	}

	s.Rules = append(s.Rules, rule)

	return rule
}

// A Rule contains a number of declarations and potentially subrules
type Rule struct {
	Parent Ruleset
	Declarations []*Declaration
	Subrules []*Rule
}

func (r *Rule) Rule(args ...any) *Rule {

	rule := &Rule{
		Parent: r,
		Declarations: filter[Declaration](args),
		Subrules: filter[Rule](args),
	}

	r.Subrules = append(r.Subrules, rule)

	return rule

}

// A declaration maps a value to a property. A value can be either a
// literal like a string, or a variable
type Declaration struct {
	Property string `json:"property"`
	Value any `json:"value"`
}

// A variable resolves to a string
type Variable interface {
	Value() string
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
			Value: value,
		}
	}
}

var BorderRadius = dec("border-radius")