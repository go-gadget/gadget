package gadget

import (
	"fmt"
	"strings"

	"github.com/go-gadget/gadget/j"
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
		Path: "/user/:id",
		Component: UserComponent,
		Children: []Route{
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
	Level   int
}

func (cr *CurrentRoute) Next() *RouteMatch {
	if cr.Level >= len(cr.Matches) {
		// default to "index" subroute?
		return nil
	}
	cr.Level++
	return cr.Matches[cr.Level-1]
}

func (route Route) Parse(parts []string) ([]*RouteMatch, []string) {
	fmt.Printf("Route %s parts %#v\n", route.Path, parts)

	routePath := strings.Trim(route.Path, "/")
	if len(parts) == 0 {
		fmt.Println("Nothing!")
		return []*RouteMatch{}, []string{}
	}
	p := parts[0]
	if p == "" {
		p = "/"
	}
	// special case: starts with :

	routeParts := strings.Split(routePath, "/")
	fmt.Printf("routeParts %s -> %#v\n", routePath, routeParts)
	// '' match ''
	// but 'user' '123' match 'user' ':id'

	// route expects more parts than we have left
	numOfParts := len(routeParts)
	if numOfParts > len(parts) {
		fmt.Printf("%#v > %#v, so no\n", routeParts, parts)
		return nil, nil
	}

	match := &RouteMatch{Route: route, Params: make(map[string]string)}
	for i, rp := range routeParts {
		if strings.HasPrefix(rp, ":") {
			fmt.Printf("I think %s and %s match\n", rp, parts[i])
			match.Params[rp[1:]] = parts[i]
		} else if rp != parts[i] {
			fmt.Printf("%s and %s don't match, not gonna work\n", rp, parts[i])
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
	fmt.Println("--- " + path + " ---")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	for _, p := range parts {
		fmt.Printf("[%s]\n", p)
	}

	for i, route := range router {
		fmt.Printf("-> Loop %d\n", i)
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

type RouterLinkComponent struct {
	BaseComponent
	Id string
	To string
}

func (r *RouterLinkComponent) Init() {
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
			j.J("Transition", "x", r.Id, r.To, r, GlobalRouter.BuildPath(r.To, map[string]string{"id": r.Id}))
		},
	}
}

func RouterLinkBuilder() Component {
	c := &RouterLinkComponent{}
	c.SetupStorage(NewStructStorage(c))
	return c
}

var GlobalRouter Router
