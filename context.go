package gospel

import (
	"fmt"
	"net/http"
)

type ElementFunction func(c Context) Element

type Context interface {
	Request() *http.Request
	Execute(ElementFunction) Element
	Modified(variable ContextVarObj)
	ElementFunction(string, ElementFunction) ElementFunction
	Element(string, ElementFunction) Element
	GetVar(key string) ContextVarObj
	SetById(id string, value any)
	GetById(id string) ContextVarObj
	AddVar(variable ContextVarObj, key string, global bool)
	AddFunc(callback ContextFuncObj, key string)
	Interactive() bool
}

type DefaultContext struct {
	key             string
	interactive     bool
	request         *http.Request
	root            *DefaultContext
	Store           *Store
	PersistentStore PersistentStore
}

type PersistentStore interface {
	Get(id string, key string, value interface{}) error
	Set(id string, key string, value interface{}) error
}

type Store struct {
	VariableIndices map[string]int
	Variables       map[string]ContextVarObj
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
		Variables:       make(map[string]ContextVarObj),
		VariableIndices: make(map[string]int),
		Funcs:           make(map[string][]ContextFuncObj),
	}
}

func (s *Store) SetById(id string, value any) {
	if variable, ok := s.Variables[id]; ok {
		variable.Set(value)
	}
}

func (s *Store) GetById(id string) ContextVarObj {
	v, _ := s.Variables[id]
	return v
}

func (s *Store) Flush() {
	s.Funcs = make(map[string][]ContextFuncObj)
	s.VariableIndices = make(map[string]int)
}

func (s *Store) GetVar(key string) ContextVarObj {
	if variable, ok := s.Variables[key]; ok {
		return variable
	}
	return nil
}

func (s *Store) AddFunc(key string, callback ContextFuncObj) int {
	s.Funcs[key] = append(s.Funcs[key], callback)
	return len(s.Funcs[key])
}

func (s *Store) AddVar(variable ContextVarObj, key string, global bool) (string, bool) {

	var i int
	var fullKey string

	// for global variables, the index will always be 0
	if !global {
		// by default, we'll use the 0 index
		i, _ = s.VariableIndices[key]
		s.VariableIndices[key] = i + 1
		fullKey = fmt.Sprintf("%s.%d", key, i)
	} else {
		fullKey = key
	}

	if v, ok := s.Variables[fullKey]; ok {
		Log.Info("Found previous value: %v", v.GetRaw())
		return fullKey, true
	}

	variable.SetCopy(false)
	s.Variables[fullKey] = variable

	return fullKey, false
}

func (d *DefaultContext) Request() *http.Request {
	return d.root.request
}

func (d *DefaultContext) SetById(id string, variable any) {
	d.root.Store.SetById(id, variable)
}

func (d *DefaultContext) GetById(id string) ContextVarObj {
	return d.root.Store.GetById(id)
}

func (d *DefaultContext) Interactive() bool {
	return d.root.interactive
}

func (d *DefaultContext) ElementFunction(key string, elementFunction ElementFunction) ElementFunction {

	return func(c Context) Element {
		return c.Element(key, elementFunction)
	}

}

func (d *DefaultContext) Element(key string, elementFunction ElementFunction) Element {

	Log.Info("Memorizing key %s.%s", d.key, key)

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
	d.root.Store.Flush()
	Log.Info("Flushing...")
	// non-interactive tree generation (i.e. do not modify variables)
	d.root.interactive = false
	return elementFunction(d)
}

func (d *DefaultContext) GetVar(key string) ContextVarObj {
	Log.Info("Variable '%s' requested from '%s'...", key, d.key)
	return d.root.Store.GetVar(key)
}

func (d *DefaultContext) AddFunc(function ContextFuncObj, key string) {
	i := d.root.Store.AddFunc(d.key, function)
	Log.Info("Adding function %s.%d", d.key, i)
	function.SetId(fmt.Sprintf("%s.%d", d.key, i))

}

func (d *DefaultContext) Modified(variable ContextVarObj) {
	Log.Info("Variable '%s' modified from '%s'", variable.Id(), d.key)
}

func (d *DefaultContext) AddVar(variable ContextVarObj, key string, global bool) {

	if key == "" {
		key = d.key
	}

	id, exists := d.root.Store.AddVar(variable, key, global)

	variable.SetId(id)
	variable.SetCopy(exists)
}
