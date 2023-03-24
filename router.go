package gospel

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
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

type RouteConfig struct {
	Route       string
	ElementFunc any
}

func (r *Router) Request() *http.Request {
	return r.context.Request()
}

func (r *Router) Context() Context {
	return r.context
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

func Route(route string, elementFunc any) *RouteConfig {
	return &RouteConfig{
		Route:       route,
		ElementFunc: elementFunc,
	}
}

// calls the handler with the context
func callElementFunc(context Context, handler any, params []string) Element {

	value := reflect.ValueOf(handler)

	if value.Kind() != reflect.Func {
		Log.Error("not a function")
		return nil
	}

	paramsValues := make([]reflect.Value, 0, len(params))

	for _, param := range params {
		paramsValues = append(paramsValues, reflect.ValueOf(param))
	}

	contextValue := reflect.ValueOf(context)

	responseValue := value.Call(append([]reflect.Value{contextValue}, paramsValues...))

	v := responseValue[0].Interface()

	if v != nil {
		return v.(Element)
	}

	return nil

}

func (r *Router) Match(routeConfigs ...*RouteConfig) Element {

	path := r.context.Request().URL.Path

	for i, routeConfig := range routeConfigs {
		re, err := regexp.Compile(routeConfig.Route)
		if err != nil {
			Log.Warning("Cannot compile route '%s': %v", routeConfig.Route, err)
			continue
		}
		Log.Info("%s - %s", path, routeConfig.Route)
		match := re.FindStringSubmatch(path)

		if len(match) > 0 {
			Log.Info("%v", match)
			return r.context.Element(fmt.Sprintf("route.%d", i), func(c Context) Element { return callElementFunc(c, routeConfig.ElementFunc, match[1:]) })
		}
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
