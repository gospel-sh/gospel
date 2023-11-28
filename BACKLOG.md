# Gospel Backlog

This is an unprioritized list with things that we want to improve, implement or fix.

## Improvements

* Currently, the `Router.Match(...)` function requires passing in a context object. We could eliminate this by returning a new type `RouterWithContext` from the `router := UseRouter(...)` call that bundles the router with the context, so that we can then call `router.Match(Route(...), Route(...)`.
* Currently, we wrap submit handlers in a `Func(...)` call that registers a function, it seems that we don't use this functionality anywhere so we can get rid off this and just use a plain old function.

## New Features

* We should implement a `OnlyOnce(...)` or `Cache(...)` function that takes a function and a list of dependencies and caches the result of the function call, so that it will not be reevaluated for multiple renders.
* We should make the rerendering smarter i.e. construct a call graph with all variables and check if we really need to rerender a given view. This might be complicated, but doable, and would increase render efficiency.
* We should introduce CSS variables that are defined in the core CSS module and can be reused across components. This would allow users to tune e.g. distances or colors even for foreign components, as long as they use these definitions from Gospel.
* We should finish implementing the parser and Gospel language to define elements and logic using a scripting language, which will open new possibilities for automated reloading and other things.
* We should implement reactive rendering based on the JS prototype we've implemented, using a functional programming approach to declaratively define component logic that can be executed both on the backend and the frontend. We need to define a common data language for this as well.

## Bugfixes

* The routers regular expression matching seems to be off, i.e. it seems that we match not only the beginning of a string but match anywhere.