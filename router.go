package gadget

import (
	"fmt"
	"strings"
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
	Children  []Route
}

type Router []Route

type RouteMatch struct {
	Route    Route
	SubPaths []string
	Params   map[string]string
}

func (route Route) Parse(parts []string) ([]RouteMatch, []string) {
	fmt.Printf("Route %s parts %#v\n", route.Path, parts)

	routePath := strings.Trim(route.Path, "/")
	if len(parts) == 0 {
		fmt.Println("Nothing!")
		return []RouteMatch{}, []string{}
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

	match := RouteMatch{Route: route, Params: make(map[string]string)}
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

	myMatch := []RouteMatch{match}
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

func (router Router) ParseRoute(path string) []RouteMatch {
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
			return result
		}
	}
	return nil
}
