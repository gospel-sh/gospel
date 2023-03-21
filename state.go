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

func (s *VarObj[T]) Set(value any) *Setter[T] {
	if sv, ok := value.(T); ok {
		return &Setter[T]{s, sv}
	} else {
		// to do: coerce
		return nil
	}
}

type ContextVarObj interface {
	SetId(string)
	Id() string
	GetRaw() any
}

type Setter[T any] struct {
	variable *VarObj[T]
	value    T
}

type FuncObj[T any] struct {
	context Context
	value   T
	id      string
}

func (c *FuncObj[T]) SetId(id string) {
	c.id = id
}

func (c *FuncObj[T]) Id() string {
	return c.id
}

type ContextFuncObj interface {
	SetId(string)
	Id() string
}

func Func[T any](c Context, value T) *FuncObj[T] {
	cf := &FuncObj[T]{c, value, ""}
	c.AddFunc(cf)
	return cf
}

func Var[T any](c Context, value T) *VarObj[T] {
	sv := &VarObj[T]{c, value, ""}
	c.AddVar(sv)
	return sv
}
