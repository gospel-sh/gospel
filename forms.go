package gospel

import (
	"net/url"
)

type FormData struct {
	context Context
	id      string
	method  string
	data    url.Values
}

const (
	POST = "POST"
	GET  = "GET"
)

func (f *FormData) Var(name string, value string) *VarObj[string] {
	v := NamedVar(f.context, name, value)

	// we check if the variable exists in the form
	if f.data != nil {

		if f.data.Has(name) {
			// this value exists, we set it
			v.Set(f.data.Get(name))
		}

	}

	return v
}

func (f *FormData) Data() url.Values {
	return f.data
}

func (f *FormData) Context() Context {
	return f.context
}

func (f *FormData) Form(args ...any) Element {
	return Form(
		append(args, Input(Type("hidden"), Name("_gspl"), Value(f.id)), Method(f.method)),
	)
}

func (f *FormData) Set(data url.Values) {
	f.data = data
}

func (f *FormData) OnSubmit(onSubmit func()) {
	req := f.context.Request()
	if req.Method == f.method && f.data.Get("_gspl") == f.id {
		onSubmit()
	}
}

func MakeFormData(c Context, id, method string) *FormData {
	req := c.Request()

	if req.Method == method {
		if HasContentType(req, "multipart/form-data") {
			// to do: make memory limit configurable
			if err := req.ParseMultipartForm(1024 * 1024 * 10); err != nil {
				// return nil, fmt.Errorf("cannot parse multipart form: %w", err)
				Log.Error("Cannot parse multipart form data: %v", err)
			}
		} else if err := req.ParseForm(); err != nil {
			Log.Error("Cannot parse form: %v", err)
			// return nil, fmt.Errorf("cannot parse form: %w", err)
		}
	}

	var data url.Values = url.Values{}

	// if the form ID matches we populate it with the submitted data
	if req.Form.Get("_gspl") == id {
		data = req.Form
	}

	return &FormData{
		context: c,
		id:      id,
		method:  "POST",
		data:    data,
	}
}
