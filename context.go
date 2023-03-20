package gospel

import (
	"fmt"
)

type Context interface {
	Key(string) Context
	AddState(variable ContextStateVariable)
	AddCallback(callback ContextCallbackFunction)
}

type DefaultContext struct {
	key   string
	root  *DefaultContext
	Store *Store
}

type Store struct {
	Variables map[string][]ContextStateVariable
	Callbacks map[string][]ContextCallbackFunction
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
		Variables: make(map[string][]ContextStateVariable),
		Callbacks: make(map[string][]ContextCallbackFunction),
	}
}

func (s *Store) AddCallback(key string, callback ContextCallbackFunction) int {
	s.Callbacks[key] = append(s.Callbacks[key], callback)
	return len(s.Callbacks[key])
}

func (s *Store) AddState(key string, variable ContextStateVariable) int {
	s.Variables[key] = append(s.Variables[key], variable)
	return len(s.Variables[key])
}

func (d *DefaultContext) Key(key string) Context {

	return &DefaultContext{
		key:  fmt.Sprintf("%s.%s", d.key, key),
		root: d.root,
	}
}

func (d *DefaultContext) AddCallback(callback ContextCallbackFunction) {
	i := d.root.Store.AddCallback(d.key, callback)
	Log.Info("Adding callback %s:%d", d.key, i)
	callback.SetId(fmt.Sprintf("%s:%d", i))

}

func (d *DefaultContext) AddState(variable ContextStateVariable) {
	i := d.root.Store.AddState(d.key, variable)
	Log.Info("Adding state %s:%d", d.key, i)
	variable.SetId(fmt.Sprintf("%s:%d", i))
}
