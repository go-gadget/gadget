package gadget

import (
	"testing"

	"github.com/go-gadget/gadget/vtree"
)

/*
  Hoe werkt inner content in een component?

  Gewoon text/html is simpel -> zet children over naar component (placeholder)
  Maar een ander component? Deze wordt gerenderd in de context van de parent, lijkt me

  <comp-1><comp-2><div g-value="x"></div></comp-2></comp-1>
  met x=1 op comp-1, x=2 op comp-2

*/

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

	t.Run("Test render of static content", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		ChildComponentFactory := MakeDummyFactory(
			"<div>Hello <slot></slot> world</div>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			"<div>So, <test-child>friendly</test-child></div>",
			map[string]*ComponentFactory{"test-child": ChildComponentFactory},
			nil,
		))
		g.Mount(component)
		g.SingleLoop()

		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := FlattenComponents(g.App).ToString()

		if rendered != "<div>So, <test-child><div>Hello <slot>friendly</slot> world</div></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	t.Run("Test inner default slot", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		ChildComponentFactory := MakeDummyFactory(
			"<div>Hello <slot>kind</slot> world</div>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			"<div>So, <test-child></test-child></div>",
			map[string]*ComponentFactory{"test-child": ChildComponentFactory},
			nil,
		))
		g.Mount(component)
		g.SingleLoop()

		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := FlattenComponents(g.App).ToString()

		if rendered != "<div>So, <test-child><div>Hello <slot>kind</slot> world</div></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	// Stuff to test: named slots
}
