package gadget

import (
	"testing"

	"github.com/go-gadget/gadget/vtree"
)

type DummyComponent struct{}

func (d *DummyComponent) Init() {
}

func (d *DummyComponent) Data() interface{} { return nil }

func (d *DummyComponent) Template() string {
	return ""
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

}
