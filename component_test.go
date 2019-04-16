package gadget

import (
	"testing"

	"github.com/go-gadget/gadget/vtree"
)

func FlattenComponents(base *ComponentInstance) *vtree.Element {
	executed := base.State.ExecutedTree
	// DeepClone changes id's, don't want that
	ee := executed.Clone().(*vtree.Element) // assume assertion is valid
	ee.Children = nil

	// string-replace won't work with duplicated components
	for _, c := range executed.Children {
		cc := c.Clone()
		ee.Children = append(ee.Children, cc)

		if el, ok := cc.(*vtree.Element); ok {
			if el.IsComponent() {
				for _, m := range base.State.Mounts {
					if m.HasComponent(el) {
						nested := FlattenComponents(m.Component)
						el.Children = append(el.Children, nested)
					}
				}
			}
		}
	}

	return ee
}

func TestComponentSlots(t *testing.T) {
	SetupTestGadget := func() (*Gadget, *TestBridge) {
		tb := NewTestBridge()
		g := NewGadget(tb)
		ChildComponentFactory := MakeDummyFactory(
			"<div>Hello <component-slot></component-slot> world</div>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			"<div>So, <test-child>friendly</test-child></div>",
			map[string]*ComponentFactory{"test-child": ChildComponentFactory},
			nil,
		))
		g.Mount(component)

		return g, tb
	}

	t.Run("Test render of content", func(t *testing.T) {
		g, _ := SetupTestGadget()
		g.SingleLoop()

		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := FlattenComponents(g.App).ToString()

		// Liefst wil je de hele tree gerenderd hebben
		if rendered != "<div>So, <test-child><div>Hello <component-slot>friendly</component-slot></div></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}

	})
}
