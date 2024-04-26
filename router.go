// Gospel - Golang Simple Extensible Web Framework
// Copyright (C) 2019-2024 - The Gospel Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the 3-Clause BSD License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// license for more details.
//
// You should have received a copy of the 3-Clause BSD License
// along with this program.  If not, see <https://opensource.org/licenses/BSD-3-Clause>.

package gospel

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

type Router struct {
	context       Context
	prefix        string
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
	Route       string         `json:"route"`
	ElementFunc any            `json:"element" graph:"include"`
	regexp      *regexp.Regexp `json:"-"`
	err         error          `json:"-"`
}

func (r *RouteConfig) Match(context Context, router *Router, generate bool) (Element, error) {

	path := context.Request().URL.Path

	var previousPath string

	currentRoute := router.CurrentRoute()

	if currentRoute != nil {
		// we remove the prefix that was already matched
		previousPath = path[len(router.Prefix()):len(router.Prefix()+currentRoute.Path)]
	}

	path = path[len(router.Prefix()+previousPath):]

	re, err := r.Regexp()

	if err != nil {
		Log.Warning("Cannot compile route '%s': %v", r.Route, err)
		return nil, nil
	}

	// we match against the current path fragment
	match := re.FindStringSubmatch(path)

	if len(match) > 0 {
		matchedRoute := &MatchedRoute{
			Config:    r,
			Path:      previousPath + match[0],
			Fragments: match[1:],
		}

		name := fmt.Sprintf("route.%s", r.Route)

		if generate {

			// to do: simplify this a bit and add a proper sub-context

			// we replace the route with the matched one
			router.PushRoute(matchedRoute)
			defer router.PopRoute()

			element, ok := matchedRoute.Config.ElementFunc.(Element)

			if !ok {
				var err error
				if element, err = callElementFunc(context, matchedRoute.Config.ElementFunc, matchedRoute.Fragments); err != nil {
					Log.Error("error in matched route '%s': %v", matchedRoute.Path, err)
					return nil, err
				}
			}

			if generator, ok := element.(Generator); ok {
				if generatedValue, err := generator.Generate(context); err != nil {
					return nil, err
				} else if generatedElement, ok := generatedValue.(Element); !ok {
					return nil, fmt.Errorf("expected an element")
				} else {
					return generatedElement, nil
				}

			}

			return element, nil
		}

		return context.Element(name, routeElementFunc(router, matchedRoute)), nil
	}
	return nil, nil
}

// generates an element if the route config matches the current route
func (r *RouteConfig) Generate(c Context) (any, error) {
	element, err := r.Match(c, UseRouter(c), true)

	if err != nil {
		return nil, err
	}

	if generator, ok := element.(Generator); ok {
		return generator.Generate(c)
	} else {
		return element, nil
	}

}

func (r *RouteConfig) RenderCode() string {
	return ""
}

func (r *RouteConfig) Regexp() (*regexp.Regexp, error) {
	if r.regexp == nil {
		// we compile the regular expression
		routeRegexp := r.Route

		// we always enforce matching from the beginning
		if !strings.HasPrefix("^", routeRegexp) && routeRegexp != "" {
			routeRegexp = "^" + routeRegexp
		}

		if re, err := regexp.Compile(routeRegexp); err != nil {
			return nil, fmt.Errorf("cannot compile regex '%s': %w", routeRegexp, err)
		} else {
			r.regexp = re
		}
	}
	return r.regexp, nil
}

func (r *Router) Prefix() string {
	return r.prefix
}

func (r *Router) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *Router) URL(path string) string {
	return fmt.Sprintf("%s%s", r.prefix, path)
}

func (r *Router) Request() *http.Request {
	return r.context.Request()
}

func (r *Router) Context() Context {
	return r.context
}

func (r *Router) RedirectTo(path string) {
	r.redirectedTo = r.URL(path)
}

func (r *Router) Query() url.Values {
	return r.context.Request().URL.Query()
}

func (r *Router) RedirectedTo() string {
	return r.redirectedTo
}

func Route(route string, elementFunc ...any) *RouteConfig {

	var element any

	if len(elementFunc) == 1 {
		element = elementFunc[0]
	} else {
		// we create a fragment
		element = F(elementFunc...)
	}

	return &RouteConfig{
		Route:       route,
		ElementFunc: element,
	}
}

// calls the handler with the context
func callElementFunc(context Context, handler any, params []string) (Element, error) {

	handlerValue := reflect.ValueOf(handler)
	handlerType := reflect.TypeOf(handler)

	if handlerValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	paramsValues := make([]reflect.Value, 0, len(params))

	for _, param := range params {
		paramsValues = append(paramsValues, reflect.ValueOf(param))
	}

	contextValue := reflect.ValueOf(context)

	var responseValue []reflect.Value

	// we check that the handler has more than one parameter
	if handlerType.NumIn() == 0 {
		return nil, fmt.Errorf("handler does not accept any arguments")
	}

	// we check that the handler accepts a context as its first parameter
	if !handlerType.In(0).Implements(reflect.TypeOf((*Context)(nil)).Elem()) {
		return nil, fmt.Errorf("handler does not accept a context")
	}

	if handlerType.NumIn() == 1 {
		// the handler only accepts a context
		responseValue = handlerValue.Call(append([]reflect.Value{contextValue}))
	} else if handlerType.NumIn() == 1+len(paramsValues) {
		// the handler accepts context and URL parameters (which we check below)
		for i := 1; i < handlerType.NumIn(); i++ {
			if handlerType.In(i).Kind() != reflect.String {
				return nil, fmt.Errorf("handler function does not accept a string")
			}
		}
		responseValue = handlerValue.Call(append([]reflect.Value{contextValue}, paramsValues...))
	} else {
		// the handler has an unexpected number of parameters
		return nil, fmt.Errorf("invalid number of parameters in handler (expected 1 or %d, got %d)", 1+len(paramsValues), handlerType.NumIn())
	}

	v := responseValue[0].Interface()

	if v != nil {
		return v.(Element), nil
	}

	return nil, nil

}

type MatchedRoute struct {
	Path      string
	Fragments []string
	Config    *RouteConfig
	Generate  bool
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
			var err error
			if element, err = callElementFunc(c, matchedRoute.Config.ElementFunc, matchedRoute.Fragments); err != nil {
				Log.Error("error in matched route '%s': %v", matchedRoute.Path, err)
				return nil
			}
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

func (r *Router) LastPath() string {
	if len(r.matchedRoutes) < 2 {
		return ""
	}
	return r.matchedRoutes[len(r.matchedRoutes)-2].Path
}

func (r *Router) RedirectUp() {
	if lastPath := r.LastPath(); lastPath != "" {
		r.RedirectTo(lastPath)
	}
}

func (r *Router) RedirectUpBy(i int) {

	if i >= len(r.matchedRoutes) {
		return
	}

	r.RedirectTo(r.matchedRoutes[len(r.matchedRoutes)-1-i].Path)
}

func (r *Router) CurrentPathWithQuery() string {
	if len(r.matchedRoutes) == 0 {
		return ""
	}
	return PathWithQuery(r.matchedRoutes[len(r.matchedRoutes)-1].Path, r.Query())
}

func (r *Router) UpdateQuery(updatedQuery map[string][]string) string {
	if len(r.matchedRoutes) == 0 {
		return ""
	}

	query := r.Query()

	for k, v := range updatedQuery {
		if v == nil {
			// we remove the key
			delete(query, k)
		} else {
			// we update the key
			query[k] = v
		}
	}

	return PathWithQuery(r.matchedRoutes[len(r.matchedRoutes)-1].Path, query)
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

func (r *Router) FullPath() string {
	return r.context.Request().URL.Path
}

func (r *Router) Match(c Context, routeConfigs ...*RouteConfig) Element {

	for _, routeConfig := range routeConfigs {

		if routeConfig == nil {
			continue
		}

		element, _ := routeConfig.Match(c, r, false)

		// if the route didn't return anything we try the next one...
		if element == nil && c.RespondWith() == nil && r.RedirectedTo() == "" {
			continue
		}
		return element
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

func init() {
	MustRegisterMacro("route", Route)
}
