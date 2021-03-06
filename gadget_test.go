package gadget

import (
	"strconv"
	"strings"
	"testing"

	"github.com/go-gadget/gadget/vtree"
)

func TestGadgetComponent(t *testing.T) {

	g := NewGadget(NewTestBridge())
	component := g.NewComponent(MakeDummyFactory("<div><p>Hi</p></div>", nil, nil))
	g.Mount(component)
	g.SingleLoop()

	if len(g.App.State.Mounts) != 0 {
		t.Errorf("Expected 0 mounted component, found %d", len(g.App.State.Mounts))
	}

	rendered := g.App.State.ExecutedTree.ToString()

	if rendered != "<div><p>Hi</p></div>" {
		t.Errorf("Did not get expected rendered tree, got %s", rendered)
	}
}

func TestNestedComponents(t *testing.T) {
	SetupTestGadget := func() (*Gadget, *TestBridge) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildComponentFactory := MakeDummyFactory(
			"<b>I am the child</b>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			"<div><test-child></test-child></div>",
			map[string]*ComponentFactory{"test-child": ChildComponentFactory},
			nil,
		))
		g.Mount(component)

		return g, tb
	}

	t.Run("Test single loop", func(t *testing.T) {
		g, _ := SetupTestGadget()
		g.SingleLoop()

		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := g.App.State.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}

		rendered = g.App.State.Mounts[0].Component.State.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test double loop", func(t *testing.T) {
		g, _ := SetupTestGadget()

		g.SingleLoop()
		g.SingleLoop()

		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := g.App.State.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.App.State.Mounts[0].Component.State.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})

	t.Run("Test many loops", func(t *testing.T) {
		g, tb := SetupTestGadget()

		g.SingleLoop()

		count := tb.AddCount

		if count == 0 {
			t.Error("Expected AddChanges but didn't get any")
		}

		for i := 0; i < 4; i++ {
			g.SingleLoop()
		}
		if count != tb.AddCount {
			t.Errorf("Unexpected extra Add actions. Expected %d, got %d",
				count, tb.AddCount)
		}
		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted components, found %d", len(g.App.State.Mounts))
		}
	})
}

func TestMultiNestedComponents(t *testing.T) {
	SetupTestGadget := func() (*Gadget, *TestBridge, *ComponentInstance) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildComponentFactory := MakeDummyFactory(
			"<b>I am the child</b>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			`<div><test-child g-if="BoolVal"></test-child>`+
				`<test-child></test-child></div>`,
			map[string]*ComponentFactory{"test-child": ChildComponentFactory},
			nil,
		))
		g.Mount(component)

		return g, tb, component
	}

	t.Run("Test multi loop odd", func(t *testing.T) {
		g, _, component := SetupTestGadget()

		val := false

		// odd loops hide conditional
		for i := 0; i < 5; i++ {
			component.RawSetValue("BoolVal", val)
			val = !val
			g.SingleLoop()
		}
		// Main + non-conditional child
		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted components, found %d", len(g.App.State.Mounts))
		}
	})
	t.Run("Test multi loop even", func(t *testing.T) {
		g, _, component := SetupTestGadget()

		val := false

		// even loops show conditional
		for i := 0; i < 6; i++ {
			component.RawSetValue("BoolVal", val)
			val = !val
			g.SingleLoop()
		}
		// Main + non-conditional child + conditional child
		if len(g.App.State.Mounts) != 2 {
			t.Errorf("Expected 2 mounted components, found %d", len(g.App.State.Mounts))
		}
	})
}

func TestConditionalComponent(t *testing.T) {
	SetupTestGadget := func() (*Gadget, *TestBridge, *ComponentInstance) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildComponentFactory := MakeDummyFactory(
			"<b>I am the child</b>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			`<div><test-child g-if="BoolVal"></test-child></div>`,
			map[string]*ComponentFactory{"test-child": ChildComponentFactory},
			nil,
		))
		g.Mount(component)
		return g, tb, component
	}

	t.Run("Test removed", func(t *testing.T) {
		g, _, component := SetupTestGadget()
		component.RawSetValue("BoolVal", false)
		g.SingleLoop()

		if len(g.App.State.Mounts) != 0 {
			t.Errorf("Expected 0 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := g.App.State.ExecutedTree.ToString()

		if rendered != "<div></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test present", func(t *testing.T) {
		g, _, component := SetupTestGadget()
		component.RawSetValue("BoolVal", true)
		g.SingleLoop()

		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := g.App.State.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.App.State.Mounts[0].Component.State.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test toggle true -> false", func(t *testing.T) {
		g, _, component := SetupTestGadget()
		component.RawSetValue("BoolVal", true)
		g.SingleLoop()
		component.RawSetValue("BoolVal", false)
		g.SingleLoop()

		if len(g.App.State.Mounts) != 0 {
			t.Errorf("Expected 0 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := g.App.State.ExecutedTree.ToString()

		if rendered != "<div></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})

	t.Run("Test toggle false -> true", func(t *testing.T) {
		g, _, component := SetupTestGadget()
		component.RawSetValue("BoolVal", false)
		g.SingleLoop()
		component.RawSetValue("BoolVal", true)
		g.SingleLoop()

		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := g.App.State.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.App.State.Mounts[0].Component.State.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})

	t.Run("Test repeated toggle", func(t *testing.T) {
		g, _, component := SetupTestGadget()
		val := false

		for i := 0; i < 4; i++ {
			component.RawSetValue("BoolVal", val)
			val = !val
			g.SingleLoop()
		}

		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := g.App.State.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.App.State.Mounts[0].Component.State.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
}

func TestForComponent(t *testing.T) {
	SetupTestGadget := func() (*Gadget, *TestBridge, *ComponentInstance) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildComponentFactory := MakeDummyFactory(
			"<b>I am the child</b>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			`<div><test-child g-for="IntArrayVal"></test-child></div>`,
			map[string]*ComponentFactory{"test-child": ChildComponentFactory},
			nil,
		))
		g.Mount(component)
		return g, tb, component
	}

	t.Run("Test 3 elements", func(t *testing.T) {
		g, _, component := SetupTestGadget()
		component.RawSetValue("IntArrayVal", []int{1, 2, 3})
		g.SingleLoop()

		if len(g.App.State.Mounts) != 3 {
			t.Errorf("Expected 3 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := g.App.State.ExecutedTree.ToString()
		if rendered != "<div><test-child></test-child><test-child></test-child><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}

		for _, m := range g.App.State.Mounts {
			rendered = m.Component.State.ExecutedTree.ToString()

			if rendered != "<b>I am the child</b>" {
				t.Errorf("Did not get expected rendered tree, got %s", rendered)
			}
		}
	})
}

func TestComponentArgs(t *testing.T) {
	SetupTestGadget := func(Props []string) (*Gadget, *TestBridge, *ComponentInstance) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildComponentFactory := MakeDummyFactory(
			`<b g-value="StringVal">I am the child</b>`,
			nil,
			Props,
		)
		component := g.NewComponent(MakeDummyFactory(
			`<div><test-child StringVal="Hello World"></test-child></div>`,
			map[string]*ComponentFactory{"test-child": ChildComponentFactory}, nil,
		))
		g.Mount(component)
		return g, tb, component
	}

	t.Run("Test direct attribute", func(t *testing.T) {
		g, _, _ := SetupTestGadget([]string{"StringVal"})
		g.SingleLoop()

		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := g.App.State.Mounts[0].Component.State.ExecutedTree.ToString()
		if rendered != "<b>Hello World</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})

	t.Run("Test bound attribute", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		ChildComponentFactory := MakeDummyFactory(
			`<b g-value="StringVal">I am the child</b>`,
			nil,
			[]string{"StringVal"},
		)
		// Because the parser assumes the ":" is actually a namespace separator,
		// it will get removed. Hence, in a template, you need to use a double ::
		// (or use g-bind:attr)
		component := g.NewComponent(MakeDummyFactory(
			`<div><test-child g-bind:StringVal="StringVal"></test-child></div>`,
			map[string]*ComponentFactory{"test-child": ChildComponentFactory}, nil,
		))

		component.RawSetValue("StringVal", "Hello World")
		g.Mount(component)
		g.SingleLoop()

		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := g.App.State.Mounts[0].Component.State.ExecutedTree.ToString()
		if rendered != "<b>Hello World</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
}

func TestForBindComponent(t *testing.T) {
	SetupTestGadget := func() (*Gadget, *TestBridge, *ComponentInstance) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildComponentFactory := MakeDummyFactory(
			`<b g-value="StringVal"></b>`,
			nil,
			[]string{"StringVal"},
		)
		component := g.NewComponent(MakeDummyFactory(
			`<div><p g-for="IntArrayVal"><test-child ::StringVal="_"></test-child></p></div>`,
			map[string]*ComponentFactory{"test-child": ChildComponentFactory},
			nil,
		))
		g.Mount(component)
		return g, tb, component
	}

	t.Run("Test passing _ to val to child", func(t *testing.T) {
		g, _, component := SetupTestGadget()
		component.RawSetValue("IntArrayVal", []int{1, 2, 3})
		g.SingleLoop()

		if len(g.App.State.Mounts) != 3 {
			t.Errorf("Expected 3 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := g.App.State.ExecutedTree.ToString()
		if c := strings.Count(rendered, "<test-child"); c != 3 {
			t.Errorf("Did not get expected number of components, got %d", c)
		}

		ids := make(map[vtree.ElementID]bool)
		for i, m := range g.App.State.Mounts {
			e := m.Component.State.ExecutedTree
			text := e.Children[0].(*vtree.Text).Text
			ids[e.ID] = true

			if text != strconv.Itoa(i+1) {
				t.Errorf("Did not get expected rendered tree, got %s", text)
			}
		}

		if len(ids) != 3 {
			t.Errorf("Expected three distinct id's, got %v", ids)
		}
	})
}

// nested loop using, e.g., [][]string

func AssertTemplateAtLevel(t *testing.T, g *Gadget, level int, expected string) {
	t.Helper()

	c := g.App
	for i := 0; i < level; i++ {
		c = c.State.Mounts[0].Component
	}

	if r := c.State.ExecutedTree.ToString(); r != expected {
		t.Errorf("Didn't get expected template at level %d, got %s", level, r)
	}
}

func AssertMountsAtLevel(t *testing.T, g *Gadget, level int, expected int) {
	t.Helper()

	c := g.App
	for i := 0; i < level; i++ {
		c = c.State.Mounts[0].Component
	}

	if r := len(c.State.Mounts); r != expected {
		t.Errorf("Didn't get expected mount-count at level %d, got %d", level, r)
	}
}

func TestRoutes(t *testing.T) {
	Level1Component := MakeNamedDummyFactory("Main", `<div>1<router-view></router-view></div>`, nil, nil)
	Level2aComponent := MakeNamedDummyFactory("2a", "<div>2a</div>", nil, nil)
	Level2bComponent := MakeNamedDummyFactory("2b", "<div>2b</div>", nil, nil)

	router := Router{
		Route{
			Path:      "/level1/:id",
			Name:      "Level1",
			Component: Level1Component,
			Children: []Route{
				Route{
					Path:      "level2a",
					Name:      "Level2a",
					Component: Level2aComponent,
				},
				Route{
					Path:      "level2b",
					Name:      "Level2b",
					Component: Level2bComponent,
				},
			},
		},
	}

	t.Run("Single loop", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Update
		}()
		g.RouterState.TransitionToPath("/level1/123/level2a")
		g.SingleLoop()

		AssertMountsAtLevel(t, g, 0, 1)
		AssertTemplateAtLevel(t, g, 2, "<div>1<router-view></router-view></div>")

		AssertMountsAtLevel(t, g, 2, 1)
		AssertTemplateAtLevel(t, g, 4, "<div>2a</div>")
	})

	t.Run("Transition a->b", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Update
			<-g.Update
		}()
		g.RouterState.TransitionToPath("/level1/123/level2a")
		g.SingleLoop()
		g.RouterState.TransitionToPath("/level1/123/level2b")
		g.SingleLoop()

		AssertMountsAtLevel(t, g, 0, 1)
		AssertTemplateAtLevel(t, g, 2, "<div>1<router-view></router-view></div>")
		AssertMountsAtLevel(t, g, 2, 1)
		AssertTemplateAtLevel(t, g, 4, "<div>2b</div>")
		// Possibly check for DeleteChange on old component?
	})
	t.Run("Multi loop", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Update
		}()
		g.RouterState.TransitionToPath("/level1/123/level2a")
		g.SingleLoop()
		g.SingleLoop()
		g.SingleLoop()
		g.SingleLoop()

		AssertMountsAtLevel(t, g, 0, 1)
		AssertTemplateAtLevel(t, g, 2, "<div>1<router-view></router-view></div>")
		AssertMountsAtLevel(t, g, 2, 1)
		AssertTemplateAtLevel(t, g, 4, "<div>2a</div>")
	})

	t.Run("Short path", func(t *testing.T) {
		// effectively a 404
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Update
		}()
		g.RouterState.TransitionToPath("/level1/")
		g.SingleLoop()

		AssertMountsAtLevel(t, g, 0, 1)
		AssertTemplateAtLevel(t, g, 2, "<div>404 - not found</div>")
	})

	t.Run("Test 404 fallback", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Update
		}()
		g.RouterState.TransitionToPath("/x")
		g.SingleLoop()

		AssertMountsAtLevel(t, g, 0, 1)
		AssertTemplateAtLevel(t, g, 2, "<div>404 - not found</div>")
	})

	t.Run("Not all routes resolved", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Update
		}()
		g.RouterState.TransitionToPath("/level1/123/")
		g.SingleLoop()

		AssertMountsAtLevel(t, g, 0, 1)
		AssertTemplateAtLevel(t, g, 2, "<div>1<router-view></router-view></div>")
	})

	t.Run("Transition up", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Update
			<-g.Update
		}()
		g.RouterState.TransitionToPath("/level1/123/level2a")
		g.SingleLoop()
		g.RouterState.TransitionToPath("/level1/123/")
		g.SingleLoop()

		AssertMountsAtLevel(t, g, 0, 1)
		AssertTemplateAtLevel(t, g, 2, "<div>1<router-view></router-view></div>")
		// At level 2 we still have the <x-component> mount, but itself has no mounts.
		AssertMountsAtLevel(t, g, 2, 1)
		AssertMountsAtLevel(t, g, 3, 0)
	})
	// Stuff to test:
	// Route doesn't change, param changes -> verify component updates (e.g. id)
}
