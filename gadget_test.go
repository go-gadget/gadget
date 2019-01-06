package gadget

import (
	"testing"

	"github.com/go-gadget/gadget/vtree"
)

type DummyComponent struct{}

func (d *DummyComponent) Init() {
}

func (d *DummyComponent) Data() interface{} {
	return d
}

func (d *DummyComponent) Template() string {
	return "<div></div>"
}

func (d *DummyComponent) Handlers() map[string]Handler {
	return nil
}

func (d *DummyComponent) Components() map[string]Builder {
	return nil
}

func DummyComponentFactory() Component {
	s := &DummyComponent{}
	return s
}

func TestGadgetComponent(t *testing.T) {

	g := NewGadget(vtree.Builder())
	component := g.BuildComponent(DummyComponentFactory)
	g.Mount(component)
	g.RenderComponents()
}
