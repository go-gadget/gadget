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
			Component: HomeComponent,
		},
		Route{
			Path:      "/user/:id",
			Component: UserComponent,
			Children: []Route{
				Route{
					Path:      "profile",
					Component: UserProfile,
				},
				Route{
					Path:      "posts",
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
	})

	t.Run("Test /user/123x", func(t *testing.T) {
		res := router.ParseRoute("/user/123")

		if len(res) != 1 {
			t.Error("Expected 1 component")
		}
		// expect UserComponent, id=123 somewhere
	})
	t.Run("Test /user/123/profile", func(t *testing.T) {
		res := router.ParseRoute("/user/123/profile")

		if len(res) != 2 {
			t.Error("Expected 2 components")
		}
		// expect UserComponent, UserProfile, id=123 somewhere
	})
}
