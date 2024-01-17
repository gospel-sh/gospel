package orm

import (
	"bytes"
	"text/template"
)

type InterpolationError struct {
	msg string
}

func (self *InterpolationError) Error() string {
	return self.msg
}

func Interpolate(s string, args interface{}) (so string, er error) {
	defer func() {
		if r := recover(); r != nil {
			so = s
			er = &InterpolationError{"Interpolate crashed..."}
		}
	}()
	// we change the default delimiters because they clash with Jinja...
	t, err := template.New("test").Delims("{", "}").Parse(s)
	buf := new(bytes.Buffer)
	err = t.Execute(buf, args)
	return buf.String(), err
}
