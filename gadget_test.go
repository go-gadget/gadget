package gadget

import (
	"strconv"
	"strings"
	"testing"

	"github.com/go-gadget/gadget/vtree"
)

type DummyComponent struct {
	GeneratedComponent
	BoolVal     bool
	IntArrayVal []int
	StringVal   string
}

func MakeDummyFactory(Template string, Components map[string]Builder, Props []string) Builder {
	return func() Component {
		s := &DummyComponent{
			GeneratedComponent: GeneratedComponent{gTemplate: Template,
				gComponents: Components, gProps: Props}}
		s.SetupStorage(NewStructStorage(s))
		return s
	}
}

type TestBridge struct {
	AttributeChangeCount uint16
	ReplaceCount         uint16
	AddCount             uint16
	DeleteCount          uint16
	InsertBeforeCount    uint16
	SyncStateCount       uint16
}

func NewTestBridge() *TestBridge {
	return &TestBridge{}
}

func (t *TestBridge) Reset() {
	t.AttributeChangeCount = 0
	t.ReplaceCount = 0
	t.AddCount = 0
	t.DeleteCount = 0
	t.InsertBeforeCount = 0
	t.SyncStateCount = 0
}

func (t *TestBridge) GetLocation() string {
	return "/"
}

func (t *TestBridge) SetLocation(string) {
}

func (t *TestBridge) AttributeChange(Target vtree.Node, Adds, Deletes, Updates vtree.Attributes) error {
	t.AttributeChangeCount++
	return nil
}
func (t *TestBridge) Replace(old vtree.Node, new vtree.Node) error {
	t.ReplaceCount++
	return nil
}
func (t *TestBridge) Add(el vtree.Node, parent vtree.Node) error {
	t.AddCount++
	return nil
}
func (t *TestBridge) Delete(el vtree.Node) error {
	t.DeleteCount++
	return nil
}
func (t *TestBridge) InsertBefore(before vtree.Node, after vtree.Node) error {
	t.InsertBeforeCount++
	return nil
}
func (t *TestBridge) SyncState(from vtree.Node) {
	t.SyncStateCount++
}

func TestGadgetComponent(t *testing.T) {

	g := NewGadget(NewTestBridge())
	component := g.NewComponent(MakeDummyFactory("<div><p>Hi</p></div>", nil, nil))
	g.Mount(component)
	g.SingleLoop()

	if len(g.App.Mounts) != 0 {
		t.Errorf("Expected 0 mounted component, found %d", len(g.App.Mounts))
	}

	rendered := g.App.ExecutedTree.ToString()

	if rendered != "<div><p>Hi</p></div>" {
		t.Errorf("Did not get expected rendered tree, got %s", rendered)
	}
}

func TestNestedComponents(t *testing.T) {
	SetupTestGadget := func() (*Gadget, *TestBridge) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildBuilder := MakeDummyFactory(
			"<b>I am the child</b>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			"<div><test-child></test-child></div>",
			map[string]Builder{"test-child": ChildBuilder},
			nil,
		))
		g.Mount(component)

		return g, tb
	}

	t.Run("Test single loop", func(t *testing.T) {
		g, _ := SetupTestGadget()
		g.SingleLoop()

		if len(g.App.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.Mounts))
		}

		rendered := g.App.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}

		rendered = g.App.Mounts[0].Component.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test double loop", func(t *testing.T) {
		g, _ := SetupTestGadget()

		g.SingleLoop()
		g.SingleLoop()

		if len(g.App.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.Mounts))
		}

		rendered := g.App.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.App.Mounts[0].Component.ExecutedTree.ToString()

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
		if len(g.App.Mounts) != 1 {
			t.Errorf("Expected 1 mounted components, found %d", len(g.App.Mounts))
		}
	})
}

func TestMultiNestedComponents(t *testing.T) {
	SetupTestGadget := func() (*Gadget, *TestBridge, *WrappedComponent) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildBuilder := MakeDummyFactory(
			"<b>I am the child</b>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			`<div><test-child g-if="BoolVal"></test-child>`+
				`<test-child></test-child></div>`,
			map[string]Builder{"test-child": ChildBuilder},
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
		if len(g.App.Mounts) != 1 {
			t.Errorf("Expected 1 mounted components, found %d", len(g.App.Mounts))
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
		if len(g.App.Mounts) != 2 {
			t.Errorf("Expected 2 mounted components, found %d", len(g.App.Mounts))
		}
	})
}

func TestConditionalComponent(t *testing.T) {
	SetupTestGadget := func() (*Gadget, *TestBridge, *WrappedComponent) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildBuilder := MakeDummyFactory(
			"<b>I am the child</b>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			`<div><test-child g-if="BoolVal"></test-child></div>`,
			map[string]Builder{"test-child": ChildBuilder},
			nil,
		))
		g.Mount(component)
		return g, tb, component
	}

	t.Run("Test removed", func(t *testing.T) {
		g, _, component := SetupTestGadget()
		component.RawSetValue("BoolVal", false)
		g.SingleLoop()

		if len(g.App.Mounts) != 0 {
			t.Errorf("Expected 0 mounted component, found %d", len(g.App.Mounts))
		}

		rendered := g.App.ExecutedTree.ToString()

		if rendered != "<div></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test present", func(t *testing.T) {
		g, _, component := SetupTestGadget()
		component.RawSetValue("BoolVal", true)
		g.SingleLoop()

		if len(g.App.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.Mounts))
		}

		rendered := g.App.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.App.Mounts[0].Component.ExecutedTree.ToString()

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

		if len(g.App.Mounts) != 0 {
			t.Errorf("Expected 0 mounted component, found %d", len(g.App.Mounts))
		}

		rendered := g.App.ExecutedTree.ToString()

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

		if len(g.App.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.Mounts))
		}

		rendered := g.App.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.App.Mounts[0].Component.ExecutedTree.ToString()

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

		if len(g.App.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.Mounts))
		}

		rendered := g.App.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.App.Mounts[0].Component.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
}

func TestForComponent(t *testing.T) {
	SetupTestGadget := func() (*Gadget, *TestBridge, *WrappedComponent) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildBuilder := MakeDummyFactory(
			"<b>I am the child</b>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			`<div><test-child g-for="IntArrayVal"></test-child></div>`,
			map[string]Builder{"test-child": ChildBuilder},
			nil,
		))
		g.Mount(component)
		return g, tb, component
	}

	t.Run("Test 3 elements", func(t *testing.T) {
		g, _, component := SetupTestGadget()
		component.RawSetValue("IntArrayVal", []int{1, 2, 3})
		g.SingleLoop()

		if len(g.App.Mounts) != 3 {
			t.Errorf("Expected 3 mounted component, found %d", len(g.App.Mounts))
		}

		rendered := g.App.ExecutedTree.ToString()
		if rendered != "<div><test-child></test-child><test-child></test-child><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}

		for _, m := range g.App.Mounts {
			rendered = m.Component.ExecutedTree.ToString()

			if rendered != "<b>I am the child</b>" {
				t.Errorf("Did not get expected rendered tree, got %s", rendered)
			}
		}
	})
}

func TestComponentArgs(t *testing.T) {
	SetupTestGadget := func(Props []string) (*Gadget, *TestBridge, *WrappedComponent) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildBuilder := MakeDummyFactory(
			`<b g-value="StringVal">I am the child</b>`,
			nil,
			Props,
		)
		component := g.NewComponent(MakeDummyFactory(
			`<div><test-child StringVal="Hello World"></test-child></div>`,
			map[string]Builder{"test-child": ChildBuilder}, nil,
		))
		g.Mount(component)
		return g, tb, component
	}

	t.Run("Test direct attribute", func(t *testing.T) {
		g, _, _ := SetupTestGadget([]string{"StringVal"})
		g.SingleLoop()

		if len(g.App.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.Mounts))
		}

		rendered := g.App.Mounts[0].Component.ExecutedTree.ToString()
		if rendered != "<b>Hello World</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})

	t.Run("Test bound attribute", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		ChildBuilder := MakeDummyFactory(
			`<b g-value="StringVal">I am the child</b>`,
			nil,
			[]string{"StringVal"},
		)
		// Because the parser assumes the ":" is actually a namespace separator,
		// it will get removed. Hence, in a template, you need to use a double ::
		// (or use g-bind:attr)
		component := g.NewComponent(MakeDummyFactory(
			`<div><test-child g-bind:StringVal="StringVal"></test-child></div>`,
			map[string]Builder{"test-child": ChildBuilder}, nil,
		))

		component.RawSetValue("StringVal", "Hello World")
		g.Mount(component)
		g.SingleLoop()

		if len(g.App.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.Mounts))
		}

		rendered := g.App.Mounts[0].Component.ExecutedTree.ToString()
		if rendered != "<b>Hello World</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
}

func TestForBindComponent(t *testing.T) {
	SetupTestGadget := func() (*Gadget, *TestBridge, *WrappedComponent) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildBuilder := MakeDummyFactory(
			`<b g-value="StringVal"></b>`,
			nil,
			[]string{"StringVal"},
		)
		component := g.NewComponent(MakeDummyFactory(
			`<div><p g-for="IntArrayVal"><test-child ::StringVal="_"></test-child></p></div>`,
			map[string]Builder{"test-child": ChildBuilder},
			nil,
		))
		g.Mount(component)
		return g, tb, component
	}

	t.Run("Test passing _ to val to child", func(t *testing.T) {
		g, _, component := SetupTestGadget()
		component.RawSetValue("IntArrayVal", []int{1, 2, 3})
		g.SingleLoop()

		if len(g.App.Mounts) != 3 {
			t.Errorf("Expected 3 mounted component, found %d", len(g.App.Mounts))
		}

		rendered := g.App.ExecutedTree.ToString()
		if c := strings.Count(rendered, "<test-child"); c != 3 {
			t.Errorf("Did not get expected number of components, got %d", c)
		}

		ids := make(map[vtree.ElementID]bool)
		for i, m := range g.App.Mounts {
			e := m.Component.ExecutedTree
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

func TestRoutes(t *testing.T) {
	Level1Component := MakeDummyFactory(`<div>1<router-view></router-view></div>`, nil, nil)
	Level2aComponent := MakeDummyFactory("<div>2a</div>", nil, nil)
	Level2bComponent := MakeDummyFactory("<div>2b</div>", nil, nil)

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
			<-g.Chan
		}()
		g.RouterState.TransitionToPath("/level1/123/level2a")
		g.SingleLoop()

		if l := len(g.App.Mounts); l != 1 {
			t.Errorf("Didn't get expected amount of level1 mounts: %d", l)
		}
		if r := g.App.Mounts[0].Component.ExecutedTree.ToString(); r != "<div>1<router-view></router-view></div>" {
			t.Errorf("Didn't get expected level1 template, got %s", r)
			return
		}
		if l := len(g.App.Mounts[0].Component.Mounts); l != 1 {
			t.Errorf("Didn't get expected amount of level2 mounts: %d", l)
			return
		}
		if r := g.App.Mounts[0].Component.Mounts[0].Component.ExecutedTree.ToString(); r != "<div>2a</div>" {
			t.Errorf("Didn't get expected level1 template, got %s", r)
		}

	})
	t.Run("Transition a->b", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Chan
			<-g.Chan
		}()
		g.RouterState.TransitionToPath("/level1/123/level2a")
		g.SingleLoop()
		g.RouterState.TransitionToPath("/level1/123/level2b")
		g.SingleLoop()

		if l := len(g.App.Mounts); l != 1 {
			t.Errorf("Didn't get expected amount of level1 mounts: %d", l)
		}
		if r := g.App.Mounts[0].Component.ExecutedTree.ToString(); r != "<div>1<router-view></router-view></div>" {
			t.Errorf("Didn't get expected level1 template, got %s", r)
		}
		if l := len(g.App.Mounts[0].Component.Mounts); l != 1 {
			t.Errorf("Didn't get expected amount of level2 mounts: %d", l)
		}
		if r := g.App.Mounts[0].Component.Mounts[0].Component.ExecutedTree.ToString(); r != "<div>2b</div>" {
			t.Errorf("Didn't get expected level1 template, got %s", r)
		}
		// Possibly check for DeleteChange on old component?
	})
	t.Run("Multi loop", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Chan
		}()
		g.RouterState.TransitionToPath("/level1/123/level2a")
		g.SingleLoop()
		g.SingleLoop()
		g.SingleLoop()
		g.SingleLoop()

		if l := len(g.App.Mounts); l != 1 {
			t.Errorf("Didn't get expected amount of level1 mounts: %d", l)
		}
		if r := g.App.Mounts[0].Component.ExecutedTree.ToString(); r != "<div>1<router-view></router-view></div>" {
			t.Errorf("Didn't get expected level1 template, got %s", r)
		}
		if l := len(g.App.Mounts[0].Component.Mounts); l != 1 {
			t.Errorf("Didn't get expected amount of level2 mounts: %d", l)
		}
		if r := g.App.Mounts[0].Component.Mounts[0].Component.ExecutedTree.ToString(); r != "<div>2a</div>" {
			t.Errorf("Didn't get expected level1 template, got %s", r)
		}
	})

	t.Run("Short path", func(t *testing.T) {
		// effectively a 404
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Chan
		}()
		g.RouterState.TransitionToPath("/level1/")
		g.SingleLoop()

		if l := len(g.App.Mounts); l != 1 {
			t.Errorf("Didn't get expected amount of level1 mounts: %d", l)
		}
		if r := g.App.Mounts[0].Component.ExecutedTree.ToString(); r != "<div>404 - not found</div>" {
			t.Errorf("Didn't get expected level1 template, got %s", r)
		}
	})

	t.Run("Test 404 fallback", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Chan
		}()
		g.RouterState.TransitionToPath("/x")
		g.SingleLoop()

		if l := len(g.App.Mounts); l != 1 {
			t.Errorf("Didn't get expected amount of level1 mounts: %d", l)
		}
		if r := g.App.Mounts[0].Component.ExecutedTree.ToString(); r != "<div>404 - not found</div>" {
			t.Errorf("Didn't get expected level1 template, got %s", r)
		}
	})

	t.Run("Not all routes resolved", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Chan
		}()
		g.RouterState.TransitionToPath("/level1/123/")
		g.SingleLoop()

		if l := len(g.App.Mounts); l != 1 {
			t.Errorf("Didn't get expected amount of level1 mounts: %d", l)
		}
		if r := g.App.Mounts[0].Component.ExecutedTree.ToString(); r != "<div>1<router-view></router-view></div>" {
			t.Errorf("Didn't get expected level1 template, got %s", r)
			return
		}
	})

	t.Run("Transition up", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		g.Router(router)

		go func() {
			<-g.Chan
			<-g.Chan
		}()
		g.RouterState.TransitionToPath("/level1/123/level2a")
		g.SingleLoop()
		g.RouterState.TransitionToPath("/level1/123/")
		g.SingleLoop()

		if l := len(g.App.Mounts); l != 1 {
			t.Errorf("Didn't get expected amount of level1 mounts: %d", l)
		}
		if r := g.App.Mounts[0].Component.ExecutedTree.ToString(); r != "<div>1<router-view></router-view></div>" {
			t.Errorf("Didn't get expected level1 template, got %s", r)
		}
		// The component disappeared
		if l := len(g.App.Mounts[0].Component.Mounts); l != 0 {
			t.Errorf("Didn't get expected amount of level2 mounts: %d", l)
		}
	})
	// Stuff to test:
	// Route doesn't change, param changes -> verify component updates (e.g. id)
}
