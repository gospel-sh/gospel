package gospel

import (
	"net/http"
	"strings"
)

type Router struct {
	context  Context
	variable *VarObj[*Router]
}

func MakeRouter(context Context) *Router {
	router := &Router{
		context: context,
	}

	router.variable = &VarObj[*Router]{context, router, ""}
	context.AddVar(router.variable, "router")

	return router
}

type Route struct {
}

func (r *Router) Request() *http.Request {
	return r.context.Request()
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

func (r *Router) Matches(route string) bool {
	if strings.HasPrefix(r.context.Request().URL.Path, route) {
		return true
	}
	return false
}

func (r *Router) Match(route string, elementFunc ElementFunction) Element {

	if strings.HasPrefix(r.context.Request().URL.Path, route) {
		return elementFunc(r.context)
	}

	return nil
}

func UseRouter(c Context) *Router {

	// check if router is defined in context already
	// if so, return it

	routerVar := GetVar[*Router](c, "router")

	if routerVar != nil {
		return routerVar.Get()
	}

	return nil

}
