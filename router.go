package gadget

import (
	"fmt"
	"strings"

	"github.com/go-gadget/gadget/j"
	"github.com/go-gadget/gadget/vtree"
)

/*
What does the router look like?

THe initialization is a nested structure with sub paths, holding components

router := []Route{
	Route{
		Path: "/",
		Component: HomeComponent,
	},
	Route{
		Path: "/user/",
		Component: UserListComponent,
	},
	Route{
		Path: "/user/:id",
		Component: UserComponent,
		Children: []Route{
			Route{
				Path: "",  // not suppported yet
				Component: UserIndex
			},
			Route{
				Path: "profile",
				Component: UserProfile,
			},
			Route{
				Path: "posts",
				Component: UserPosts,
			}
		}
	}
}

These routes are then set on Gadget, which will
- on startup check the current path and map it to a component or a set of nested components
- will detecct route changes (transitions) and rerender appropriaately

TODO:
- support index route (named "")
*/

type Route struct {
	Path      string
	Name      string
	Component Builder
	Children  Router
}

type Router []Route

type RouteMatch struct {
	Route    Route
	SubPaths []string
	Params   map[string]string
}

type CurrentRoute struct {
	Matches []*RouteMatch
	Params  map[string]string
}

func (cr *CurrentRoute) Get(level int) *RouteMatch {
	if level >= len(cr.Matches) {
		// default to "index" subroute?
		return nil
	}
	return cr.Matches[level]
}

func (cr *CurrentRoute) PathID(level int) string {
	if level == -1 {
		level = len(cr.Matches) - 1
	}

	if level >= len(cr.Matches) {
		level = len(cr.Matches) - 1
	}

	parts := make([]string, level+1)
	for i := 0; i <= level; i++ {
		parts[i] = cr.Matches[i].Route.Name
	}

	return strings.Join(parts, ".")
}

func (route Route) Parse(parts []string) ([]*RouteMatch, []string) {
	routePath := strings.Trim(route.Path, "/")
	if len(parts) == 0 {
		return []*RouteMatch{}, []string{}
	}
	p := parts[0]
	if p == "" {
		p = "/"
	}
	// special case: starts with :

	routeParts := strings.Split(routePath, "/")

	// route expects more parts than we have left
	numOfParts := len(routeParts)
	if numOfParts > len(parts) {
		return nil, nil
	}

	match := &RouteMatch{Route: route, Params: make(map[string]string)}
	for i, rp := range routeParts {
		if strings.HasPrefix(rp, ":") {
			match.Params[rp[1:]] = parts[i]
		} else if rp != parts[i] {
			return nil, nil
		}
		match.SubPaths = append(match.SubPaths, rp)
	}

	myMatch := []*RouteMatch{match}
	remaining := parts[numOfParts:] // can be empty!

	if len(remaining) > 0 {

		// at this point the route matches, but does the rest match?
		for _, childRoute := range route.Children {
			childRes, childRemain := childRoute.Parse(parts[numOfParts:])
			if childRes != nil && len(childRemain) == 0 {
				return append(myMatch, childRes...), []string{}
			}
		}
	}
	return myMatch, remaining
}

func (router Router) Find(name string) []Route {
	for _, r := range router {
		if r.Name == name {
			return []Route{r}
		}
	}
	// Start recursing
	for _, r := range router {
		if len(r.Children) > 0 {
			if sub := r.Children.Find(name); sub != nil {
				return append([]Route{r}, sub...)
			}
		}
	}
	return nil

}
func (router Router) BuildPath(name string, params map[string]string) string {
	// How to deal with '/' when constructing paths? Always end in /?

	route := router.Find(name)
	path := ""
	if route != nil {
		for _, r := range route {
			routeParts := strings.Split(r.Path, "/")
			for _, rp := range routeParts {
				if strings.HasPrefix(rp, ":") {
					if val, ok := params[rp[1:]]; ok {
						path += val + "/"
					} else {
						fmt.Printf("Could not map %s to value\n", rp)
						path += rp + "/"
					}
				} else {
					path += rp + "/"
				}
			}
		}
	}
	return path
}

func (router Router) Parse(path string) *CurrentRoute {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	for _, route := range router {
		result, remainder := route.Parse(parts)
		if result != nil && len(remainder) == 0 {
			cr := &CurrentRoute{Matches: result, Params: make(map[string]string)}
			for _, m := range result {
				for k, v := range m.Params {
					cr.Params[k] = v
				}
			}
			return cr
		}
	}
	return nil
}

func SetRouter(router *Router) {
	registry := GetRegistry()
	registry.Register("router", router)
}

func GetRouter() *Router {
	registry := GetRegistry()
	return registry.Get("router").(*Router)
}

type TransitionAction struct {
	oldPath string
	newPath string
}

func (t *TransitionAction) Run() {
}

type RouterState struct {
	Router       Router
	CurrentRoute *CurrentRoute
	Route404     Route
	oldPath      string
	newPath      string
	Update       chan Action
}

func NewRouterState() *RouterState {
	rs := &RouterState{}
	rs.Route404 = Route{
		Name:      "404",
		Path:      "",
		Component: GenerateComponent("<div>404 - not found</div>", nil, nil),
	}
	return rs
}

func SetRouterState(state *RouterState) {
	registry := GetRegistry()
	registry.Register("router-state", state)
}

func GetRouterState() *RouterState {
	registry := GetRegistry()
	return registry.Get("router-state").(*RouterState)
}

func (rs *RouterState) TransitionToPath(path string) {
	j.J("Transition", path)
	oldPath := rs.oldPath
	rs.oldPath = path
	bridge := GetRegistry().Get("bridge").(vtree.Subject)
	bridge.SetLocation(path)

	rs.CurrentRoute = rs.Router.Parse(path)
	if rs.CurrentRoute == nil {
		// We could inject the actual path into a copy of the 404 route?
		rs.CurrentRoute = &CurrentRoute{Matches: []*RouteMatch{&RouteMatch{Route: rs.Route404}}}
	}
	rs.Update <- &TransitionAction{oldPath, path}
}

func (rs *RouterState) TransitionToName(name string, params map[string]string) {
	newPath := rs.Router.BuildPath(name, params)
	if newPath != rs.oldPath {
		rs.oldPath = newPath

		rs.TransitionToPath(newPath)
	}
}

type RouteTraverser struct {
	level int
	cr    *CurrentRoute
}

func NewRouteTraverser(cr *CurrentRoute) *RouteTraverser {
	return &RouteTraverser{0, cr}
}

func (rt *RouteTraverser) PathID() string {
	return rt.cr.PathID(rt.level)
}

func (rt *RouteTraverser) Component(ElementType string) Builder {
	if ElementType == "router-view" {
		if rm := rt.cr.Get(rt.level); rm != nil {
			return rm.Route.Component
		}
	} else if ElementType == "router-link" {
		return RouterLinkBuilder
	}
	return nil
}

func (rt *RouteTraverser) Up() {
	rt.level++
}

type RouterLinkComponent struct {
	BaseComponent

	Id string
	To string
}

func (r *RouterLinkComponent) Props() []string {
	// support "*" - just send me all you got?
	return []string{"Id", "To"}
}

func (r *RouterLinkComponent) Template() string {
	return `<button g-click="transition">XXXfillin</button>`
}

func (r *RouterLinkComponent) Handlers() map[string]Handler {
	return map[string]Handler{
		"transition": func(Updates chan Action) {
			GetRouterState().TransitionToName(r.To, map[string]string{"id": r.Id})
		},
	}
}

func RouterLinkBuilder() Component {
	c := &RouterLinkComponent{}
	c.SetupStorage(NewStructStorage(c))
	return c
}
