package gadget

import (
	"fmt"

	"github.com/go-gadget/gadget/vtree"
)

type Mount struct {
	Component   *WrappedComponent
	Point       *vtree.Element
	ToBeRemoved bool
}

func (m *Mount) HasComponent(componentElement *vtree.Element) bool {
	fmt.Printf("Comparing mounts %v and %v\n", m.Point.ID, componentElement.ID)
	if m.Point == nil {
		return false
	}
	return m.Point.Equals(componentElement)
}
