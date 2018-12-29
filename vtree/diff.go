package vtree

func Diff(el Node, other Node) ChangeSet {
	changeSet := make(ChangeSet, 0)

	if !el.Equals(other) {
		// fmt.Printf("%s and %s are different", el.GetID(), other.GetID())
		changeSet = append(changeSet,
			&ReplaceChange{el, other})
		// we can stop now
		return changeSet
	}

	// The rest only makes sense with non-text nodes
	element, ok := el.(*Element)
	if !ok {
		return changeSet
	}

	oElement := other.(*Element) // This can't fail, right?

	AttrChange := &AttributeChange{Target: oElement,
		Adds:    make(Attributes),
		Deletes: make(Attributes),
		Updates: make(Attributes)}

	// Check for deleted attributes
	for k, v := range element.Attributes {
		if _, ok := oElement.Attributes[k]; !ok {
			// k was removed
			AttrChange.Deletes[k] = v
		}
	}
	// Check for added/changed attributes
	for k, v := range oElement.Attributes {
		if oldV, ok := element.Attributes[k]; !ok {
			// k was added
			AttrChange.Adds[k] = v
		} else if v != oldV {
			// k changed to v
			AttrChange.Updates[k] = v
		}
	}

	// has anything changed?
	if len(AttrChange.Adds)+len(AttrChange.Deletes)+len(AttrChange.Updates) > 0 {
		changeSet = append(changeSet, AttrChange)
	}

	childChanges := element.Children.Diff(element, oElement.Children)
	changeSet = append(changeSet, childChanges...)
	return changeSet
}

func (one NodeList) Diff(parent *Element, other NodeList) ChangeSet {
	// perhaps *Element diff is just this on a single-element array

	// loop over other - this is the order we want. Try to
	// find the elements
	//
	// [] -> [1]
	// [1, 2] -> [2, 1]
	//  find last node -> 1, work backwards and insertbefore
	// insertBefore and appendChild both remove from current
	// position. But append adds at the end which is problematic

	// find nodes to remove
	// find nodes to add
	// check if order is correct. If not, find last node
	// and insetBefore to the front

	// perhaps we need a map[ElementID]*Element (but doesn't keep order

	changeSet := make(ChangeSet, 0)

	oldElements := make(map[ElementID]Node)
	newElements := make(map[ElementID]Node)

	for _, el := range one {
		oldElements[el.GetID()] = el
	}

	for _, el := range other {
		newElements[el.GetID()] = el
	}

	orderChanged := false

	// Elements to delete
	for _, el := range one {
		if _, exists := newElements[el.GetID()]; !exists {
			changeSet = append(changeSet,
				&DeleteChange{el})
			// removing doesn't change order, but its easier to assume it does
			orderChanged = true
		}
	}

	// Elements to add
	for _, el := range other {
		if _, exists := oldElements[el.GetID()]; !exists {
			changeSet = append(changeSet,
				&AddChange{Parent: parent, Node: el})
			// adding may change order
			orderChanged = true
		} else {
			extra := Diff(oldElements[el.GetID()], el)
			if extra != nil {
				changeSet = append(changeSet, extra...)
			}
			// elements that were in both may have changed
			// check change el
		}
	}

	// check order
	if !orderChanged {
		// order didn't change through add/del, but can still be
		// different. Old/New must be same size
		for pos, el := range other {
			if el.GetID() != one[pos].GetID() {
				orderChanged = true
				break
			}
		}
	}

	if orderChanged && len(other) > 1 {
		last := other[len(other)-1]

		for i := len(other) - 2; i >= 0; i-- {
			changeSet = append(changeSet,
				&MoveBeforeChange{Node: last, Before: other[i]})
			last = other[i]
		}
	}

	return changeSet
}
