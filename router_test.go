package gadget

import "testing"

func TestRouter(t *testing.T) {
	HomeComponent := MakeDummyFactory("<div>Home<router-view></router-view></div>", nil, nil)
	UserComponent := MakeDummyFactory("<div>User<router-view></router-view></div>", nil, nil)
	UserProfile := MakeDummyFactory("<div>Profile<router-view></router-view></div>", nil, nil)
	UserPosts := MakeDummyFactory("<div>Posts<router-view></router-view></div>", nil, nil)

	router := Router{
		Route{
			Path:      "/",
			Name:      "Home",
			Component: HomeComponent,
		},
		Route{
			Path:      "/user/:id",
			Name:      "User",
			Component: UserComponent,
			Children: []Route{
				Route{
					Path:      "profile",
					Name:      "UserProfile",
					Component: UserProfile,
				},
				Route{
					Path:      "posts",
					Name:      "UserPosts",
					Component: UserPosts,
				},
			},
		},
	}

	t.Run("Test /", func(t *testing.T) {
		res := router.ParseRoute("/")

		if len(res) != 1 {
			t.Error("Expected 1 component")
		}

		if res[0].Route.Name != "Home" {
			t.Errorf("Expected HomeComponent, got %v", res[0].Route.Name)
		}
	})

	t.Run("Test /user/123", func(t *testing.T) {
		res := router.ParseRoute("/user/123")

		if len(res) != 1 {
			t.Error("Expected 1 component")
		}
		// expect UserComponent, id=123 somewhere
		if res[0].Route.Name != "User" {
			t.Errorf("Expected UserComponent, got %v", res[0].Route.Name)
		}
		if idParam := res[0].Params["id"]; idParam != "123" {
			t.Errorf("Expected id to match 123, got %s", idParam)
		}
	})
	t.Run("Test /user/123/profile", func(t *testing.T) {
		res := router.ParseRoute("/user/123/profile")

		if len(res) != 2 {
			t.Error("Expected 2 components")
		}
		if res[0].Route.Name != "User" {
			t.Errorf("Expected UserComponent, got %v", res[0].Route.Name)
		}
		if res[1].Route.Name != "UserProfile" {
			t.Errorf("Expected UserProfile, got %v", res[0].Route.Name)
		}
		// expect UserComponent, UserProfile, id=123 somewhere

		Match(t, res[0]).Name("User").Paths("user", ":id").Params("id", "123")
		Match(t, res[1]).Name("UserProfile")
	})
}

type RouteMatcher struct {
	t     *testing.T
	match RouteMatch
}

func Match(t *testing.T, rm RouteMatch) *RouteMatcher {
	return &RouteMatcher{t, rm}
}

func (rm *RouteMatcher) Name(name string) *RouteMatcher {
	rm.t.Helper()
	if rm.match.Route.Name != name {
		rm.t.Errorf("Expected route to match %s, got %s", name, rm.match.Route.Name)
	}
	return rm
}
func (rm *RouteMatcher) Paths(parts ...string) *RouteMatcher {
	rm.t.Helper()

	if got, expected := len(rm.match.SubPaths), len(parts); got != expected {
		rm.t.Errorf("Number of params differs from expected. Expected %d, got %d", expected, got)
	}
	for i, p := range rm.match.SubPaths {
		if p != parts[i] {
			rm.t.Errorf("Paths differ at %d: Expected %s, got %s", i, p, parts[i])
		}
	}
	return rm
}

func (rm *RouteMatcher) Params(key string, value string) *RouteMatcher {
	rm.t.Helper()

	got, ok := rm.match.Params[key]

	if !ok {
		rm.t.Errorf("Didn't get expected param %s", key)
	}
	if got != value {
		rm.t.Errorf("Value for %s doesn't match expected. Got %s, expected %s", key, got, value)
	}
	return rm
}
