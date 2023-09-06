package gospel

import (
	"net/url"
)

type FormData struct {
	context Context
	data    *VarObj[url.Values]
}

func (f *FormData) Var(name string, value string) *VarObj[string] {
	v := NamedVar(f.context, name, value)

	// we check if the variable exists in the form
	if f.data != nil {

		d := f.data.Get()

		if d.Has(name) {
			// this value exists, we set it
			v.Set(d.Get(name))
		}

	}

	return v
}

func (f *FormData) Data() url.Values {
	return f.data.Get()
}

func (f *FormData) Context() Context {
	return f.context
}

func (f *FormData) Set(data url.Values) {
	f.data.Set(data)
}

func MakeFormData(c Context) *FormData {
	return &FormData{
		context: c,
		data:    Var[url.Values](c, url.Values(nil)),
	}
}
