package gospel

import (
	"fmt"
	"net/http"
)

type ElementFunction func(c Context) Element

type Context interface {
	Request() *http.Request
	Redirect(string)
	RedirectedTo() string
	Execute(ElementFunction) Element
	Modified(variable ContextVarObj)
	Element(string, ElementFunction) Element
	GetVar(key string) ContextVarObj
	AddVar(variable ContextVarObj, key string)
	AddFunc(callback ContextFuncObj, key string)
	Interactive() bool
}

type DefaultContext struct {
	key         string
	interactive bool
	redirectTo  string
	request     *http.Request
	root        *DefaultContext
	Store       *Store
}

type Store struct {
	VariableIndices map[string]int
	Variables       map[string][]ContextVarObj
	Funcs           map[string][]ContextFuncObj
}

func MakeDefaultContext(request *http.Request) *DefaultContext {
	dc := &DefaultContext{
		key:     "root",
		request: request,
		Store:   MakeStore(),
	}

	dc.root = dc

	return dc
}

func MakeStore() *Store {
	return &Store{
		Variables:       make(map[string][]ContextVarObj),
		VariableIndices: make(map[string]int),
		Funcs:           make(map[string][]ContextFuncObj),
	}
}

func (s *Store) Flush() {
	s.Funcs = make(map[string][]ContextFuncObj)
	s.VariableIndices = make(map[string]int)
}

func (s *Store) GetVar(key string, index int) ContextVarObj {
	if vars, ok := s.Variables[key]; ok {
		if index > 0 && len(vars) >= index {
			return vars[index-1]
		}
	}
	return nil
}

func (s *Store) AddFunc(key string, callback ContextFuncObj) int {
	s.Funcs[key] = append(s.Funcs[key], callback)
	return len(s.Funcs[key])
}

func (s *Store) AddVar(key string, variable ContextVarObj) int {

	// by default, we'll use the 0 index
	i, _ := s.VariableIndices[key]

	s.VariableIndices[key] = i + 1

	if vars, ok := s.Variables[key]; ok {
		if i < len(vars) {
			// this variable exists already
			variable.Set(vars[i].GetRaw())
			// we replace the variable...
			vars[i] = variable
			return i + 1
		}
	}

	s.Variables[key] = append(s.Variables[key], variable)
	return len(s.Variables[key])
}

func (d *DefaultContext) Redirect(path string) {
	d.redirectTo = path
}

func (d *DefaultContext) RedirectedTo() string {
	return d.redirectTo
}

func (d *DefaultContext) Request() *http.Request {
	return d.request
}

func (d *DefaultContext) Interactive() bool {
	return d.root.interactive
}

func (d *DefaultContext) Element(key string, elementFunction ElementFunction) Element {

	Log.Info("Memorizing key %s", key)

	c := &DefaultContext{
		key:  fmt.Sprintf("%s.%s", d.key, key),
		root: d.root,
	}

	element := elementFunction(c)

	return element

}

func (d *DefaultContext) Execute(elementFunction ElementFunction) Element {
	d.root.interactive = true
	// interactive tree generation (i.e. call functions to modify variables)
	elementFunction(d)
	d.Store.Flush()
	// non-interactive tree generation (i.e. do not modify variables)
	d.root.interactive = false
	return elementFunction(d)
}

func (d *DefaultContext) GetVar(key string) ContextVarObj {
	Log.Info("Variable '%s' requested from '%s'...", key, d.key)
	return d.root.Store.GetVar(key, 1)
}

func (d *DefaultContext) AddFunc(callback ContextFuncObj, key string) {
	i := d.root.Store.AddFunc(d.key, callback)
	Log.Info("Adding callback %s.%d", d.key, i)
	callback.SetId(fmt.Sprintf("%s.%d", d.key, i))

}

func (d *DefaultContext) Modified(variable ContextVarObj) {
	Log.Info("Variable '%s' modified from '%s'", variable.Id(), d.key)
}

func (d *DefaultContext) AddVar(variable ContextVarObj, key string) {

	if key == "" {
		key = d.key
	}

	i := d.root.Store.AddVar(key, variable)
	Log.Info("Adding state %s.%d", key, i)
	variable.SetId(fmt.Sprintf("%s.%d", key, i))
}

func SetVar(c Context, v any, key string) {
	variable := MakeVarObj(c, v)
	c.AddVar(variable, key)
}

func UseVar[T any](c Context, key string) T {

	variable := c.GetVar(key)

	if variable != nil {
		if vt, ok := variable.GetRaw().(T); ok {
			return vt
		}
	}
	return *new(T)
}
