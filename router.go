package gospel

import (
	"net/http"
)

type Router struct {
	request  *http.Request
	context  Context
	variable *VarObj[*Router]
}

func MakeRouter(request *http.Request) *Router {
	return &Router{
		request: request,
	}
}

type Route struct {
}

func (r *Router) Request() *http.Request {
	return r.request
}

func (r *Router) SetContext(c Context) {
	r.context = c
	r.variable = &VarObj[*Router]{c, r, ""}
	c.AddVar(r.variable, "router")
}

func (r *Router) Route() *Route {
	return &Route{}
}

func (r *Router) RedirectTo(url string) {
	// we notify the controller that the route was modified
	r.context.Modified(r.variable)
}

func UseRouter(c Context) *Router {

	// check if router is defined in context already
	// if so, return it

	routerVar := c.GetVar("router", 1)

	if routerVar == nil {
		return nil
	}

	if router, ok := routerVar.GetRaw().(*Router); ok {
		return router
	}

	return nil

}
