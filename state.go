package gospel

type VarObj[T any] struct {
	context Context
	value   T
	id      string
	copy    bool
}

func MakeVarObj[T any](context Context) *VarObj[T] {
	return &VarObj[T]{
		context: context,
		id:      "",
	}
}

func (s *VarObj[T]) SetCopy(copy bool) {
	s.copy = copy
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

func (s *VarObj[T]) GetRaw() any {
	return s.Get()
}

func (s *VarObj[T]) Set(value any) {
	if s.copy {
		s.context.SetById(s.id, value)
	} else if sv, ok := value.(T); ok {
		s.value = sv
	}
}

type ContextVarObj interface {
	SetId(string)
	Id() string
	SetCopy(bool)
	Set(any)
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

func Var[T any](c Context, value T) *VarObj[T] {
	sv := MakeVarObj[T](c)
	sv.Set(value)
	c.AddVar(sv, "", false)
	return sv
}

func GlobalVar[T any](c Context, key string, v T) *VarObj[T] {

	variable := MakeVarObj[T](c)
	variable.Set(v)

	c.AddVar(variable, key, true)

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
