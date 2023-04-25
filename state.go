package gospel

import (
	"fmt"
)

type VarObj[T any] struct {
	context     Context
	value       T
	generator   func() T
	id          string
	copy        bool
	persistent  bool
	initialized bool
}

func MakeVarObj[T any](context Context, generator func() T) *VarObj[T] {
	return &VarObj[T]{
		context:   context,
		generator: generator,
		id:        "",
	}
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

func (s *VarObj[T]) GetRaw() any {
	return s.Get()
}

func (s *VarObj[T]) Persistent() bool {
	return s.persistent
}

func (s *VarObj[T]) SetPersistent(value bool) {
	s.persistent = value
}

func (s *VarObj[T]) Initialized() bool {
	return s.initialized
}

func (s *VarObj[T]) Set(value any) error {
	if s.copy {
		s.context.SetById(s.id, value)
	} else if sv, ok := value.(T); ok {
		s.value = sv
		s.initialized = true
	} else {
		return fmt.Errorf("type error")
	}
	return nil
}

type ContextVarObj interface {
	SetPersistent(bool)
	Persistent() bool
	Initialized() bool
	SetId(string)
	Id() string
	SetCopy(bool)
	IsCopy() bool
	Set(any) error
	GetRaw() any
}

type FuncObj struct {
	context Context
	value   func()
	id      string
}

func (c *FuncObj) SetId(id string) {
	c.id = id
}

func (c *FuncObj) Call() {
	if c.context.Interactive() {
		c.value()
	}
}

func (c *FuncObj) Id() string {
	return c.id
}

func (c *FuncObj) Context() Context {
	return c.context
}

func Convert[T any](v any) T {
	if vv, ok := v.(T); ok {
		return vv
	}
	return *new(T)
}

type ContextFuncObj interface {
	Context() Context
	SetId(string)
	Id() string
	Call()
}

type Opts struct {
	Foo string
}

func Func(c Context, value func()) *FuncObj {

	var f Opts = Convert[Opts](Opts{"Test"})

	Log.Info("%v", f)

	cf := &FuncObj{c, value, ""}
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
