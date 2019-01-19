package gadget

import (
	"testing"

	"github.com/go-gadget/gadget/j"
	"github.com/go-gadget/gadget/vtree"
)

type DummyComponent struct {
	DummyTemplate   string
	DummyComponents map[string]Builder
	BoolVal         bool
}

func (d *DummyComponent) Init() {
}

func (d *DummyComponent) Data() interface{} {
	return d
}

func (d *DummyComponent) Template() string {
	return d.DummyTemplate
}

func (d *DummyComponent) Handlers() map[string]Handler {
	return nil
}

func (d *DummyComponent) Components() map[string]Builder {
	return d.DummyComponents
}

func MakeDummyFactory(Template string, Components map[string]Builder) Builder {
	return func() Component {
		s := &DummyComponent{DummyTemplate: Template, DummyComponents: Components}
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
	component := g.BuildComponent(MakeDummyFactory("<div><p>Hi</p></div>", nil))
	g.Mount(component, nil)
	g.SingleLoop()

	if len(g.Mounts) != 1 {
		t.Errorf("Expected 1 mounted component, found %d", len(g.Mounts))
	}

	c := g.Mounts[0].Component

	rendered := c.ExecutedTree.ToString()

	if rendered != "<div><p>Hi</p></div>" {
		t.Errorf("Did not get expected rendered tree, got %s", rendered)
	}
}

func SetupTestGadget() (*Gadget, *TestBridge) {
	tb := NewTestBridge()
	g := NewGadget(tb)
	ChildBuilder := MakeDummyFactory(
		"<b>I am the child</b>",
		nil,
	)
	component := g.BuildComponent(MakeDummyFactory(
		"<div><test-child></test-child></div>",
		map[string]Builder{"test-child": ChildBuilder},
	))
	g.Mount(component, nil)

	return g, tb
}

func TestNestedComponents(t *testing.T) {
	t.Run("Test single loop", func(t *testing.T) {
		g, _ := SetupTestGadget()
		g.SingleLoop()

		if len(g.Mounts) != 2 {
			t.Errorf("Expected 2 mounted component, found %d", len(g.Mounts))
			j.J(g.Mounts)
		}

		// Can we assume this order?
		rendered := g.Mounts[0].Component.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.Mounts[1].Component.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test double loop", func(t *testing.T) {
		g, _ := SetupTestGadget()
		// actually, g will already have accumulated an extra loop, so this will be tripple loop
		g.SingleLoop()
		g.SingleLoop()

		if len(g.Mounts) != 2 {
			t.Errorf("Expected 2 mounted component, found %d", len(g.Mounts))
		}

		// Can we assume this order?
		rendered := g.Mounts[0].Component.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.Mounts[1].Component.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test many loops", func(t *testing.T) {
		g, tb := SetupTestGadget()

		g.SingleLoop()

		count := tb.AddCount

		if count == 0 {
			t.Errorf("Didn't get any Add changes on bridge")
		}

		for i := 0; i < 5; i++ {
			g.SingleLoop()
		}
		if count != tb.AddCount {
			t.Errorf("Unexpected extra Add actions. Expected %d, got %d",
				count, tb.AddCount)
		}
	})
}

func TestConditionalComponent(t *testing.T) {
	tb := NewTestBridge()
	g := NewGadget(tb)
	ChildBuilder := MakeDummyFactory(
		"<b>I am the child</b>",
		nil,
	)
	component := g.BuildComponent(MakeDummyFactory(
		`<div><test-child g-if="BoolVal"></test-child></div>`,
		map[string]Builder{"test-child": ChildBuilder},
	))
	g.Mount(component, nil)

	t.Run("Test removed", func(t *testing.T) {
		component.RawSetValue("BoolVal", false)
		g.SingleLoop()

		if len(g.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.Mounts))
		}

		rendered := g.Mounts[0].Component.ExecutedTree.ToString()

		if rendered != "<div></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test present", func(t *testing.T) {
		component.RawSetValue("BoolVal", true)
		g.SingleLoop()

		if len(g.Mounts) != 2 {
			t.Errorf("Expected 2 mounted component, found %d", len(g.Mounts))
		}

		// Can we assume this order?
		rendered := g.Mounts[0].Component.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.Mounts[1].Component.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test toggle true -> false", func(t *testing.T) {
		component.RawSetValue("BoolVal", true)
		g.SingleLoop()
		component.RawSetValue("BoolVal", false)
		g.SingleLoop()

		if len(g.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.Mounts))
		}

		// Can we assume this order?
		rendered := g.Mounts[0].Component.ExecutedTree.ToString()

		if rendered != "<div></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test toggle false -> true", func(t *testing.T) {
		component.RawSetValue("BoolVal", false)
		g.SingleLoop()
		component.RawSetValue("BoolVal", true)
		g.SingleLoop()

		if len(g.Mounts) != 2 {
			t.Errorf("Expected 2 mounted component, found %d", len(g.Mounts))
		}

		// Can we assume this order?
		rendered := g.Mounts[0].Component.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.Mounts[1].Component.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test repeated toggle", func(t *testing.T) {
		val := false

		for i := 0; i < 4; i++ {
			component.RawSetValue("BoolVal", val)
			val = !val
			g.SingleLoop()
		}

		if len(g.Mounts) != 2 {
			t.Errorf("Expected 2 mounted component, found %d", len(g.Mounts))
		}

		// Can we assume this order?
		rendered := g.Mounts[0].Component.ExecutedTree.ToString()

		if rendered != "<div><test-child></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
		rendered = g.Mounts[1].Component.ExecutedTree.ToString()

		if rendered != "<b>I am the child</b>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
}
