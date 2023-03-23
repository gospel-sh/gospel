package gospel

func If(condition any, args ...any) []any {

	var c bool

	if cv, ok := condition.(*VarObj[bool]); ok {
		c = cv.Get()
	} else if c, ok = condition.(bool); !ok {
		// to do: raise a warning
		return nil
	}
	if c {
		return args
	}
	return nil
}

type VarObj[T any] struct {
	context Context
	value   T
	id      string
}

func MakeVarObj[T any](context Context, value T) *VarObj[T] {
	return &VarObj[T]{
		context: context,
		value:   value,
		id:      "",
	}
}

func (s *VarObj[T]) SetId(id string) {
	s.id = id
}

func (s *VarObj[T]) Id() string {
	return s.id
}

func (s *VarObj[T]) Get() T {
	return s.value
}

func (s *VarObj[T]) GetRaw() any {
	return s.value
}

func (s *VarObj[T]) Set(value any) {
	if tv, ok := value.(T); ok {
		s.value = tv
	}
}

type ContextVarObj interface {
	SetId(string)
	Id() string
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
	sv := &VarObj[T]{c, value, ""}
	c.AddVar(sv, "")
	return sv
}

func GetVar[T any](c Context, key string) *VarObj[T] {
	sv := c.GetVar(key)

	if sv != nil {
		if svt, ok := sv.GetRaw().(T); ok {
			return &VarObj[T]{c, svt, key}
		}
	}

	return nil
}
