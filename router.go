package gadget

import (
	"fmt"
	"strings"

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

// A Route is a single sub-path in a tree of routes
type Route struct {
	Path string
	Name string
	// rename to Factory ?
	Component *ComponentFactory
	Children  Router
}

type Traversable interface {
	BeforeTraverse()
}

// A Router definition is a collection of (nested) Routes
type Router []Route

// A RouteMatch is the match of a path against a chain of nested routes, possibly containing dynamic paths (params)
type RouteMatch struct {
	Route    Route
	SubPaths []string
	Params   map[string]string
}

// CurrentRoute is a RouteMatch with all params collected (and de-duplicated)
type CurrentRoute struct {
	Matches []*RouteMatch
	Params  map[string]string
}

// Get retrieves a route in a RouteMatch at a specific level
func (cr *CurrentRoute) Get(level int) *RouteMatch {
	if level >= len(cr.Matches) {
		// default to "index" subroute?
		return nil
	}
	return cr.Matches[level]
}

// Parse matches a split path into a sequence of (nested) Routes and dynamic routes
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

// Find recursively searches the router for the given named route
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

// BuildPath constructs a ("reverse") path out of a given route name and params
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

// Parse parses a path into a RouteMatch
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

// GetRouter gets the Router from the registry
func GetRouter(registry *Registry) *Router {
	if r := registry.Get("router"); r != nil {
		return r.(*Router)
	}
	return nil
}

type TransitionAction struct {
	oldPath string
	newPath string
}

func (t *TransitionAction) Run() {
}

type RouterState struct {
	Registry     *Registry
	CurrentRoute *CurrentRoute
	Route404     Route
	oldPath      string
	newPath      string
	Update       chan Action
}

func NewRouterState(registry *Registry) *RouterState {
	rs := &RouterState{Registry: registry}
	rs.Route404 = Route{
		Name:      "404",
		Path:      "",
		Component: GenerateComponentFactory("gadget.router.404", "<div>404 - not found</div>", nil, nil),
	}
	return rs
}

func GetRouterState(registry *Registry) *RouterState {
	return registry.Get("router-state").(*RouterState)
}

func (rs *RouterState) TransitionToPath(path string) {
	oldPath := rs.oldPath
	rs.oldPath = path
	bridge := rs.Registry.Get("bridge").(vtree.Subject)
	bridge.SetLocation(path)

	rs.CurrentRoute = GetRouter(rs.Registry).Parse(path)
	if rs.CurrentRoute == nil {
		// We could inject the actual path into a copy of the 404 route?
		rs.CurrentRoute = &CurrentRoute{Matches: []*RouteMatch{&RouteMatch{Route: rs.Route404}}}
	}
	rs.Update <- &TransitionAction{oldPath, path}
}

func (rs *RouterState) TransitionToName(name string, params map[string]string) {
	newPath := GetRouter(rs.Registry).BuildPath(name, params)
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

func RegisterRouterComponents(registry *Registry) {
	if cr := GetComponentRegistry(registry); cr != nil {
		cr.Register("router-view", RouterViewComponentFactory)
		cr.Register("router-link", RouterLinkComponentFactory)
		fmt.Println("x")
	}
}

func (rt *RouteTraverser) PopRoute() *RouteMatch {
	r := rt.cr.Get(rt.level)
	rt.level++
	return r
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
		"transition": func() {
			GetRouterState(r.State.Registry).TransitionToName(r.To, map[string]string{"id": r.Id})
		},
	}
}

var RouterLinkComponentFactory = &ComponentFactory{
	Name: "gadget.router.RouterLink",
	Builder: func() Component {
		c := &RouterLinkComponent{}
		c.SetupStorage(NewStructStorage(c))
		return c
	},
}

type RouterViewComponent struct {
	BaseComponent
	firstSlot  bool
	secondSlot bool // XXX need "not value" (or g-else) in g-value!
	level      int
	state      map[string]*ComponentFactory
	tpl        string
}

func (r *RouterViewComponent) Template() string {
	return `<div><x-component1 g-if="firstSlot"></x-component1><x-component2 g-if="secondSlot"></x-component2></div>`
}

func (r *RouterViewComponent) BeforeTraverse() {
	var m *Mount

	rt := GetGadget(r.State.Registry).Traverser
	// Have we already been visited?
	if r.level == -1 {
		// New component.
		r.level = rt.level
	} else if r.level != rt.level {
		return
	}

	r.state = map[string]*ComponentFactory{"x-component1": nil, "x-component2": nil}
	MountedName := ""

	if len(r.State.Mounts) > 0 {
		m = r.State.Mounts[0]
		MountedName = m.Name
	}

	// c is the component for the current route level
	c := rt.PopRoute()

	if c == nil {
		if m != nil {
			m.ToBeRemoved = true
		}
		return
	}

	if MountedName != c.Route.Component.Name {
		if m != nil {
			m.ToBeRemoved = true
			r.firstSlot = !r.firstSlot
			r.secondSlot = !r.firstSlot
		}
	}

	slot := "x-component1"
	if !r.firstSlot {
		slot = "x-component2"
	}
	r.state[slot] = c.Route.Component

}
func (r *RouterViewComponent) Components() map[string]*ComponentFactory {
	return r.state
}

var RouterViewComponentFactory = &ComponentFactory{
	Name: "gadget.router.RouterView",
	Builder: func() Component {
		c := &RouterViewComponent{firstSlot: true, secondSlot: false, level: -1}
		c.SetupStorage(NewStructStorage(c))
		return c
	},
}
