package gospel

func If(condition any, args ...any) []any {

	var c bool

	if cv, ok := condition.(*StateVariable[bool]); ok {
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

type StateVariable[T any] struct {
	context Context
	value   T
	id      string
}

type ContextStateVariable interface {
	SetId(string)
}

func (s *StateVariable[T]) SetId(id string) {
	s.id = id
}

func (s *StateVariable[T]) Get() T {
	return s.value
}

type Setter[T any] struct {
	variable *StateVariable[T]
	value    T
}

type CallbackFunction[T any] struct {
	context Context
	value   T
	id      string
}

type ContextCallbackFunction interface {
	SetId(string)
}

func (c *CallbackFunction[T]) SetId(id string) {
	c.id = id
}

func Callback[T any](c Context, value T) *CallbackFunction[T] {
	cf := &CallbackFunction[T]{c, value, ""}
	c.AddCallback(cf)
	return cf
}

func (s *StateVariable[T]) Set(value any) *Setter[T] {
	if sv, ok := value.(T); ok {
		return &Setter[T]{s, sv}
	} else {
		// to do: coerce
		return nil
	}
}

func State[T any](c Context, value T) *StateVariable[T] {
	sv := &StateVariable[T]{c, value, ""}
	c.AddState(sv)
	return sv
}
