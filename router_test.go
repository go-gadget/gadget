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
		res := router.Parse("/")

		if len(res.Matches) != 1 {
			t.Error("Expected 1 component")
		}

		AssertRoute(t, res.Matches[0]).Name("Home")
	})

	t.Run("Test /user/123", func(t *testing.T) {
		res := router.Parse("/user/123")

		if len(res.Matches) != 1 {
			t.Error("Expected 1 component")
		}
		AssertRoute(t, res.Matches[0]).Name("User").Paths("user", ":id").Params("id", "123")
		if v, ok := res.Params["id"]; !ok || v != "123" {
			t.Errorf("Expected id to be in params and have value 123, got %v", v)
		}
	})
	t.Run("Test missing id: /user/", func(t *testing.T) {
		res := router.Parse("/user/")

		if res != nil {
			t.Error("Expected res to be nil")
		}
	})
	t.Run("Test /user/123/profile", func(t *testing.T) {
		res := router.Parse("/user/123/profile")

		if len(res.Matches) != 2 {
			t.Error("Expected 2 components")
		}

		AssertRoute(t, res.Matches[0]).Name("User").Paths("user", ":id").Params("id", "123")
		AssertRoute(t, res.Matches[1]).Name("UserProfile")
	})

	t.Run("Test build UserProfile route", func(t *testing.T) {
		path := router.BuildPath("UserProfile", map[string]string{"id": "123"})

		if path != "/user/123/profile/" {
			t.Errorf("Didn't get expected path, got %s", path)
		}
	})

	// Test path-id building
	t.Run("Test full nested path id", func(t *testing.T) {
		res := router.Parse("/user/123/profile")

		if pID := res.PathID(0); pID != "User" {
			t.Errorf("Didn't get expected PathID, got %s", pID)
		}

		if pID := res.PathID(1); pID != "User.UserProfile" {
			t.Errorf("Didn't get expected PathID, got %s", pID)
		}

		if pID := res.PathID(-1); pID != "User.UserProfile" {
			t.Errorf("Didn't get expected PathID, got %s", pID)
		}
	})
	t.Run("Test PathID when only param changed", func(t *testing.T) {
		a := router.Parse("/user/123/profile").PathID(-1)
		b := router.Parse("/user/234/profile").PathID(-1)
		if a != b {
			t.Errorf("Expected PathID's to be identical, but got %s <-> %s", a, b)
		}
	})
	t.Run("Test CurrentRoute full nested path id", func(t *testing.T) {
		res := router.Parse("/user/123/profile")

		rm := res.Matches

		AssertRoute(t, rm[0]).Name("User").Params("id", "123")
		AssertRoute(t, rm[1]).Name("UserProfile")
	})
}

type RouteMatcher struct {
	t     *testing.T
	match *RouteMatch
}

func AssertRoute(t *testing.T, rm *RouteMatch) *RouteMatcher {
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
