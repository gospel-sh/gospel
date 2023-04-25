package gospel

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
)

type Router struct {
	context       Context
	matchedRoutes []*MatchedRoute
	variable      ContextVarObj
	redirectedTo  string
}

func PathWithQuery(path string, query map[string][]string) string {
	values := url.Values(query)
	return fmt.Sprintf("%s?%s", path, values.Encode())
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

func (r *Router) RedirectTo(path string) {
	r.redirectedTo = path
}

func (r *Router) Query() url.Values {
	return r.context.Request().URL.Query()
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
	Path      string
	Fragments []string
	Config    *RouteConfig
}

func routeElementFunc(r *Router, matchedRoute *MatchedRoute) ElementFunction {
	// we ensure the full routing context is always present when the function
	// is being called, as the context might call the element function
	// repeatedly e.g. due to variable changes...
	return func(c Context) Element {
		// we replace the route with the matched one
		r.PushRoute(matchedRoute)
		// we call the element function with the given context

		element, ok := matchedRoute.Config.ElementFunc.(Element)

		if !ok {
			element = callElementFunc(c, matchedRoute.Config.ElementFunc, matchedRoute.Fragments)
		}

		// we restore the previous route
		r.PopRoute()

		return element
	}
}

func (r *Router) CurrentRoute() *MatchedRoute {

	if len(r.matchedRoutes) == 0 {
		return nil
	}

	return r.matchedRoutes[len(r.matchedRoutes)-1]
}

func (r *Router) PushRoute(matchedRoute *MatchedRoute) {
	r.matchedRoutes = append(r.matchedRoutes, matchedRoute)
}

func (r *Router) RedirectUp() {
	if len(r.matchedRoutes) < 2 {
		return
	}

	r.RedirectTo(r.matchedRoutes[len(r.matchedRoutes)-2].Path)
}

func (r *Router) CurrentPathWithQuery() string {
	if len(r.matchedRoutes) == 0 {
		return ""
	}
	return PathWithQuery(r.matchedRoutes[len(r.matchedRoutes)-1].Path, r.Query())
}

func (r *Router) CurrentPath() string {
	if len(r.matchedRoutes) == 0 {
		return ""
	}
	return r.matchedRoutes[len(r.matchedRoutes)-1].Path
}

func (r *Router) PopRoute() {

	if len(r.matchedRoutes) == 0 {
		return
	}

	r.matchedRoutes = r.matchedRoutes[:len(r.matchedRoutes)-1]
}

func (r *Router) Match(c Context, routeConfigs ...*RouteConfig) Element {

	path := r.context.Request().URL.Path

	var previousPath string

	if r.CurrentRoute() != nil {
		// we remove the prefix that was already matched
		previousPath = path[:len(r.CurrentRoute().Path)]
		path = path[len(r.CurrentRoute().Path):]
	}

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
				Path:      previousPath + match[0],
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
