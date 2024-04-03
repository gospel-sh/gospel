package gospel

import (
	"net/url"
)

type FormData struct {
	context Context
	id      string
	method  string
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
	req := c.Request()

	if HasContentType(req, "multipart/form-data") {
		// to do: make memory limit configurable
		if err := req.ParseMultipartForm(1024 * 1024 * 10); err != nil {
			// return nil, fmt.Errorf("cannot parse multipart form: %w", err)
			Log.Error("Cannot parse multipart form data: $w", err)
		}
	} else if err := req.ParseForm(); err != nil {
		Log.Error("Cannot parse form: %w", err)
		// return nil, fmt.Errorf("cannot parse form: %w", err)
	}

	return &FormData{
		context: c,
		data:    Var[url.Values](c, req.Form),
	}
}
