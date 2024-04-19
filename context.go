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
	"net/http"
)

type ElementFunction func(c Context) Element
type PureElementFunction = func() Element
type RespondWithFunction func(c Context, w http.ResponseWriter)

type Context interface {
	Request() *http.Request
	Execute(ElementFunction) Element
	SetRespondWith(RespondWithFunction)
	RespondWith() RespondWithFunction
	ElementFunction(string, ElementFunction) ElementFunction
	DeferElement(string, ElementFunction) PureElementFunction
	Element(string, ElementFunction) Element
	GetVar(key string) ContextVarObj
	SetById(id string, value any) error
	GetById(id string) ContextVarObj
	StatusCode() int
	SetStatusCode(int)
	AddVar(variable ContextVarObj, key string) error
	AddFunc(callback ContextFuncObj[any], key string)
	Interactive() bool
	ResponseWriter() http.ResponseWriter
	Scope(string) Context
	Key() string
	Clear()
}

type DefaultContext struct {
	key         string
	interactive bool
	statusCode  int
	respondWith RespondWithFunction
	request     *http.Request
	writer      http.ResponseWriter
	root        *DefaultContext
	Store       *Store
}

type PersistentStore interface {
	Get(key string, value ContextVarObj) error
	Set(key string, value ContextVarObj) error
	Clear()
}

type Store struct {
	VariableIndices map[string]int
	Variables       map[string]ContextVarObj
	Funcs           map[string][]ContextFuncObj[any]
	persistentStore PersistentStore
}

func MakeDefaultContext(request *http.Request, writer http.ResponseWriter, store *Store) *DefaultContext {
	dc := &DefaultContext{
		key:        "root",
		request:    request,
		writer:     writer,
		Store:      store,
		statusCode: 200,
	}

	dc.root = dc

	return dc
}

func MakeStore(persistentStore PersistentStore) *Store {
	return &Store{
		Variables:       make(map[string]ContextVarObj),
		VariableIndices: make(map[string]int),
		Funcs:           make(map[string][]ContextFuncObj[any]),
		persistentStore: persistentStore,
	}
}

func (s *Store) SetById(id string, value any) error {
	if variable, ok := s.Variables[id]; ok {
		return variable.Set(value)
	}
	return fmt.Errorf("not found")
}

func (s *Store) GetById(id string) ContextVarObj {
	v, _ := s.Variables[id]
	return v
}

func (s *Store) Flush() {
	s.Funcs = make(map[string][]ContextFuncObj[any])
	s.VariableIndices = make(map[string]int)
}

func (s *Store) GetVar(key string) ContextVarObj {
	if variable, ok := s.Variables[key]; ok {
		return variable
	}
	return nil
}

func (s *Store) Clear() {
	s.persistentStore.Clear()
}

func (s *Store) AddFunc(key string, callback ContextFuncObj[any]) int {
	s.Funcs[key] = append(s.Funcs[key], callback)
	return len(s.Funcs[key])
}

func (s *Store) Finalize() {
	for key, variable := range s.Variables {
		if variable.Persistent() {
			s.persistentStore.Set(key, variable)
		}
	}
}

func (s *Store) AddVar(variable ContextVarObj, key string, global bool) error {

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

	variable.SetId(fullKey)

	if _, ok := s.Variables[fullKey]; ok {
		variable.SetCopy(true)
		return nil
	}

	// this variable is new
	variable.SetCopy(false)

	s.Variables[fullKey] = variable

	// we check if the variable exists in the persistent store

	if variable.Persistent() {
		return s.persistentStore.Get(fullKey, variable)
	}

	return nil

}

func (d *DefaultContext) Clear() {
	d.root.Store.Clear()
}

func (d *DefaultContext) ResponseWriter() http.ResponseWriter {
	return d.root.writer
}

func (d *DefaultContext) SetRespondWith(f RespondWithFunction) {
	d.root.respondWith = f
}

func (d *DefaultContext) RespondWith() RespondWithFunction {
	return d.root.respondWith
}

func (d *DefaultContext) StatusCode() int {
	return d.root.statusCode
}

func (d *DefaultContext) SetStatusCode(code int) {
	d.root.statusCode = code
}

func (d *DefaultContext) Request() *http.Request {
	return d.root.request
}

func (d *DefaultContext) SetById(id string, variable any) error {
	return d.root.Store.SetById(id, variable)
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

func (d *DefaultContext) DeferElement(key string, elementFunction ElementFunction) PureElementFunction {

	c := &DefaultContext{
		key:  fmt.Sprintf("%s.%s", d.key, key),
		root: d.root,
	}

	return func() Element {
		return elementFunction(c)
	}
}

func (d *DefaultContext) Element(key string, elementFunction ElementFunction) Element {

	c := &DefaultContext{
		key:  fmt.Sprintf("%s.%s", d.key, key),
		root: d.root,
	}

	return elementFunction(c)

}

func (d *DefaultContext) Scope(key string) Context {
	return &DefaultContext{
		key:  key,
		root: d.root,
	}
}

func (d *DefaultContext) Execute(elementFunction ElementFunction) Element {
	d.root.interactive = true
	return elementFunction(d)
}

func (d *DefaultContext) GetVar(key string) ContextVarObj {
	return d.root.Store.GetVar(key)
}

func (d *DefaultContext) AddFunc(function ContextFuncObj[any], key string) {
	i := d.root.Store.AddFunc(d.key, function)
	function.SetId(fmt.Sprintf("%s.%d", d.key, i))

}

func (d *DefaultContext) AddVar(variable ContextVarObj, key string) error {

	global := true

	if key == "" {
		global = false
		// we use the key of the current context
		key = d.key
	}

	return d.root.Store.AddVar(variable, key, global)
}

func (d *DefaultContext) Key() string {
	return d.key
}
