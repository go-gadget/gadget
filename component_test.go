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
	/*

			Does not work - nested components are too complex. For example, under which component would
			the inner component be mounted?

			Besides that, since component tags are preserved, they will be re-rendered inside the component
			slot which fails in horrible ways

		t.Run("Test nested components", func(t *testing.T) {
			g := NewGadget(NewTestBridge())
			ChildComponentFactory := MakeDummyFactory(
				"<div>I am child 1 <slot></slot></div>",
				nil,
				nil,
			)
			Child2ComponentFactory := MakeDummyFactory(
				"<div>I am child 2 <slot></slot></div>",
				nil,
				nil,
			)

			component := g.NewComponent(MakeDummyFactory(
				`<div><test-child><test2-child><div g-value="StringVal">123</div></test2-child></test-child></div>`,
				map[string]*ComponentFactory{"test-child": ChildComponentFactory, "test2-child": Child2ComponentFactory},
				nil,
			))
			g.Mount(component)
			component.SetValue("StringVal", "Friendly")
			g.SingleLoop()

			if len(g.App.State.Mounts) != 2 {
				t.Errorf("Expected 1 mounted component, found %d", len(g.App.State.Mounts))
			}

			rendered := FlattenComponents(g.App).ToString()

			if rendered != "<div><test-child><div>I am child 1 <slot><test2-child><div>I am child 2 <slot>Friendly</slot></test2-child></slot></test-child></div>" {
				t.Errorf("Did not get expected rendered tree, got %s", rendered)
			}
		})
	*/
}
