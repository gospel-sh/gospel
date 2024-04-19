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
	"encoding/json"
	"fmt"
	"strings"
)

type VarObj[T any] struct {
	context     Context
	value       T
	generator   func() T
	id          string
	copy        bool
	persistent  bool
	initialized bool
	clear       bool
	onUpdate    func()
}

func MakeVarObj[T any](context Context, generator func() T) *VarObj[T] {
	return &VarObj[T]{
		context:   context,
		generator: generator,
		id:        "",
	}
}

func (s *VarObj[T]) OnUpdate(onUpdate func()) {
	s.onUpdate = onUpdate
}

func (s *VarObj[T]) SetCopy(copy bool) {
	s.copy = copy
}

func (s *VarObj[T]) IsCopy() bool {
	return s.copy
}

func (s *VarObj[T]) SetId(id string) {
	s.id = id
}

func (s *VarObj[T]) Id() string {
	return s.id
}

func (s *VarObj[T]) Get() T {
	if s.copy {
		if vt, ok := s.context.GetById(s.id).GetRaw().(T); ok {
			return vt
		}
		return *new(T)
	}
	return s.value
}

func (s *VarObj[T]) Reset() {
	s.Set(s.generator())
}

func (s *VarObj[T]) Serialize() ([]byte, error) {
	return json.Marshal(s.value)
}

func (s *VarObj[T]) Deserialize(data []byte) error {

	nv := *new(T)

	if err := json.Unmarshal(data, &nv); err != nil {
		return err
	}

	s.value = nv
	s.initialized = true

	return nil

}

func (s *VarObj[T]) New() any {
	return new(T)
}

func (s *VarObj[T]) GetRaw() any {
	return s.Get()
}

func (s *VarObj[T]) Persistent() bool {
	return s.persistent
}

func (s *VarObj[T]) Clear() {
	s.clear = true
}

func (s *VarObj[T]) SetPersistent(value bool) {
	s.persistent = value
}

func (s *VarObj[T]) Initialized() bool {
	return s.initialized
}

func (s *VarObj[T]) Context() Context {
	return s.context
}

func (s *VarObj[T]) ScopedId() string {
	// Returns the variable ID without the context key.
	// If the variable is global, it will return the full ID.
	return strings.TrimPrefix(s.id, fmt.Sprintf("%s.", s.context.Key()))
}

func (s *VarObj[T]) Set(value any) error {
	if s.copy {
		s.context.SetById(s.id, value)
	} else if sv, ok := value.(T); ok {
		s.value = sv
		s.initialized = true
	} else if value != nil {
		Log.Error("type error: %T vs. %T", value, *new(T))
		return fmt.Errorf("type error")
	}

	if s.onUpdate != nil {
		s.onUpdate()
	}

	return nil
}

type ContextVarObj interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
	SetPersistent(bool)
	Persistent() bool
	Initialized() bool
	Context() Context
	ScopedId() string
	OnUpdate(func())
	SetId(string)
	Id() string
	SetCopy(bool)
	IsCopy() bool
	Set(any) error
	GetRaw() any
	New() any
}

type FuncObj[T any] struct {
	context Context
	value   func()
	id      string
}

func (c *FuncObj[T]) SetId(id string) {
	c.id = id
}

func (c *FuncObj[T]) Call() {
	if c.context.Interactive() {
		c.value()
	}
}

func (c *FuncObj[T]) Id() string {
	return c.id
}

func (c *FuncObj[T]) Context() Context {
	return c.context
}

func Convert[T any](v any) T {
	if vv, ok := v.(T); ok {
		return vv
	}
	return *new(T)
}

type ContextFuncObj[T any] interface {
	Context() Context
	SetId(string)
	Id() string
	Call()
}

type Opts struct {
	Foo string
}

func Func[T any](c Context, value func()) *FuncObj[T] {
	cf := &FuncObj[T]{c, value, ""}
	c.AddFunc(cf, "")
	return cf
}

func PersistentVar[T any](c Context, value T) *VarObj[T] {
	return persistentVar(c, value, "")
}

func PersistentGlobalVar[T any](c Context, key string, value T) *VarObj[T] {
	return persistentVar(c, value, key)
}

func persistentVar[T any](c Context, value T, key string) *VarObj[T] {
	sv := MakeVarObj[T](c, func() T { return value })
	sv.SetPersistent(true)
	c.AddVar(sv, key)

	// we only set the persistent variable if it hasn't been initialized yet
	if !sv.IsCopy() && !sv.Initialized() {
		sv.Set(value)
	}

	return sv
}

func CachedVar[T any](c Context, value func() T) *VarObj[T] {
	sv := MakeVarObj[T](c, value)
	c.AddVar(sv, "")

	if !sv.IsCopy() {
		sv.Set(value())
	}

	return sv
}

func Var[T any](c Context, value T) *VarObj[T] {
	return CachedVar(c, func() T { return value })
}

func NamedVar[T any](c Context, key string, v T) *VarObj[T] {

	if key == "" {
		panic("empty key")
	}

	key = fmt.Sprintf("%s.%s", c.Key(), key)

	return GlobalVar[T](c, key, v)
}

func GlobalVar[T any](c Context, key string, v T) *VarObj[T] {

	if key == "" {
		panic("empty key")
	}

	variable := MakeVarObj[T](c, func() T { return v })
	variable.Set(v)

	c.AddVar(variable, key)

	return variable
}

func UseGlobal[T any](c Context, key string) T {

	variable := c.GetVar(key)

	if variable != nil {
		if vt, ok := variable.GetRaw().(T); ok {
			return vt
		}
	}
	return *new(T)
}
