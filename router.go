package gospel

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
)

type Router struct {
	context      Context
	currentRoute *MatchedRoute
	variable     ContextVarObj
	redirectedTo string
}

func MakeRouter(context Context) *Router {
	router := &Router{
		context: context,
	}

	router.variable = GlobalVar(context, "router", router)

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
	r.redirectedTo = url
}

func (r *Router) RedirectedTo() string {
	return r.redirectedTo
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

type MatchedRoute struct {
	Fragments []string
	Config    *RouteConfig
}

func routeElementFunc(r *Router, matchedRoute *MatchedRoute) ElementFunction {
	// we ensure the full routing context is always present when the function
	// is being called, as the context might call the element function
	// repeatedly e.g. due to variable changes...
	return func(c Context) Element {
		// we replace the route with the matched one
		previousRoute := r.ReplaceCurrentRoute(matchedRoute)
		// we call the element function with the given context
		element := callElementFunc(c, matchedRoute.Config.ElementFunc, matchedRoute.Fragments)
		// we restore the previous route
		r.ReplaceCurrentRoute(previousRoute)

		return element
	}
}

func (r *Router) ReplaceCurrentRoute(matchedRoute *MatchedRoute) *MatchedRoute {
	currentRoute := r.currentRoute
	r.currentRoute = matchedRoute
	return currentRoute
}

func (r *Router) Match(c Context, routeConfigs ...*RouteConfig) Element {

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

			matchedRoute := &MatchedRoute{
				Config:    routeConfig,
				Fragments: match[1:],
			}

			return c.Element(fmt.Sprintf("route.%d", i), routeElementFunc(r, matchedRoute))
		}
	}

	return nil
}

func UseRouter(c Context) *Router {

	// check if router is defined in context already
	// if so, return it

	routerVar := GlobalVar[*Router](c, "router", nil)

	if routerVar != nil {
		return routerVar.Get()
	}

	return nil

}
