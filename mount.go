package gadget

import "github.com/go-gadget/gadget/vtree"

type Mount struct {
	Component   *WrappedComponent
	Point       *vtree.Element
	ToBeRemoved bool
}

func (m *Mount) HasComponent(componentElement *vtree.Element) bool {
	if m.Point == nil {
		return false
	}
	return m.Point.Equals(componentElement)
}
