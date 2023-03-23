package gospel

import (
	"net/http"
	"strings"
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

func (r *Router) Context() Context {
	return r.context
}

func (r *Router) Route() *Route {
	return &Route{}
}

func (r *Router) RedirectTo(url string) {
	// we notify the controller that the route was modified
	r.context.Modified(r.variable)
}

func (r *Router) Match(route string, elementFunc ElementFunction) Element {

	if strings.HasPrefix(r.request.URL.Path, route) {
		return elementFunc(r.context)
	}

	return nil
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
