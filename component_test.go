package gadget

import (
	"testing"
)

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
	t.Run("Test dynamic value", func(t *testing.T) {
		g := NewGadget(NewTestBridge())
		ChildComponentFactory := MakeDummyFactory(
			"<div>Hello <slot>kind</slot> world</div>",
			nil,
			nil,
		)
		component := g.NewComponent(MakeDummyFactory(
			`<div>So, <test-child><div g-value="StringVal"></div></test-child></div>`,
			map[string]*ComponentFactory{"test-child": ChildComponentFactory},
			nil,
		))
		g.Mount(component)
		component.SetValue("StringVal", "Friendly")
		g.SingleLoop()

		if len(g.App.State.Mounts) != 1 {
			t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
		}

		rendered := FlattenComponents(g.App).ToString()

		if rendered != "<div>So, <test-child><div>Hello <slot><div>Friendly</div></slot> world</div></test-child></div>" {
			t.Errorf("Did not get expected rendered tree, got %s", rendered)
		}
	})
	// Stuff to test: named slots
}
