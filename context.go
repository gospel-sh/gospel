package gospel

import (
	"fmt"
)

type Context interface {
	Key(string) Context
	AddVar(variable ContextVarObj)
	AddFunc(callback ContextFuncObj)
}

type DefaultContext struct {
	key   string
	root  *DefaultContext
	Store *Store
}

type Store struct {
	Variables map[string][]ContextVarObj
	Funcs     map[string][]ContextFuncObj
}

func MakeDefaultContext() *DefaultContext {
	dc := &DefaultContext{
		key:   "root",
		Store: MakeStore(),
	}

	dc.root = dc

	return dc
}

func MakeStore() *Store {
	return &Store{
		Variables: make(map[string][]ContextVarObj),
		Funcs:     make(map[string][]ContextFuncObj),
	}
}

func (s *Store) AddFunc(key string, callback ContextFuncObj) int {
	s.Funcs[key] = append(s.Funcs[key], callback)
	return len(s.Funcs[key])
}

func (s *Store) AddVar(key string, variable ContextVarObj) int {
	s.Variables[key] = append(s.Variables[key], variable)
	return len(s.Variables[key])
}

func (d *DefaultContext) Key(key string) Context {

	return &DefaultContext{
		key:  fmt.Sprintf("%s.%s", d.key, key),
		root: d.root,
	}
}

func (d *DefaultContext) AddFunc(callback ContextFuncObj) {
	i := d.root.Store.AddFunc(d.key, callback)
	Log.Info("Adding callback %s.%d", d.key, i)
	callback.SetId(fmt.Sprintf("%s.%d", d.key, i))

}

func (d *DefaultContext) AddVar(variable ContextVarObj) {
	i := d.root.Store.AddVar(d.key, variable)
	Log.Info("Adding state %s.%d", d.key, i)
	variable.SetId(fmt.Sprintf("%s.%d", d.key, i))
}
