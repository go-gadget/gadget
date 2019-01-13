package gadget

import (
	"testing"

	"github.com/go-gadget/gadget/vtree"
)

type DummyComponent struct {
	DummyTemplate   string
	DummyComponents map[string]Builder
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

func TestGadgetComponent(t *testing.T) {

	g := NewGadget(vtree.Builder())
	component := g.BuildComponent(MakeDummyFactory("<div><p>Hi</p></div>", nil))
	g.Mount(component, nil)
	g.SingleLoop()

	if len(g.Components) != 1 {
		t.Errorf("Expected 1 mounted component, found %d", len(g.Components))
	}

	c := g.Components[0].Component

	rendered := c.ExecutedTree.ToString()

	if rendered != "<div><p>Hi</p></div>" {
		t.Errorf("Did not get expected rendered tree, got %s", rendered)
	}
}

func TestNestedComponent(t *testing.T) {
	g := NewGadget(vtree.Builder())
	ChildBuilder := MakeDummyFactory(
		"<b>I am the child</b>",
		nil,
	)
	component := g.BuildComponent(MakeDummyFactory(
		"<div><test-child></test-child></div>",
		map[string]Builder{"test-child": ChildBuilder},
	))
	g.Mount(component, nil)
	g.SingleLoop()

	if len(g.Components) != 2 {
		t.Errorf("Expected 2 mounted component, found %d", len(g.Components))
	}

	c := g.Components[0].Component

	rendered := c.ExecutedTree.ToString()

	if rendered != "<div><p>Hi</p></div>" {
		t.Errorf("Did not get expected rendered tree, got %s", rendered)
	}
}
