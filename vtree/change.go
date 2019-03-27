package vtree

/*
 * Subject is the subject of changes, e.g. a DOM
 * It's broader than that now - it syncs state, gets location,
 * basically the bridge to the other side
 */
type Subject interface {
	AttributeChange(Target Node, Adds, Deletes, Updates Attributes) error
	Replace(old Node, new Node) error
	Add(el Node, parent Node) error
	Delete(el Node) error
	InsertBefore(before Node, after Node) error
	SyncState(from Node)
	GetLocation() string
	SetLocation(string)
}

/* Change should be on the 'other side' domtree, not on a local
 * Element based tree
 */
type Change interface {
	// this needs something to apply on, e.g. the real DOM
	Apply(subject Subject) error
}

// A ChangeSet is a collection of changes
type ChangeSet []Change

// ApplyChanges applies all changes in a ChangeSet
func (changes ChangeSet) ApplyChanges(subject Subject) {
	for _, change := range changes {
		change.Apply(subject)
	}
}

// AttributeChange tracks changes on an Elements attributes
type AttributeChange struct {
	Target  Node
	Adds    Attributes
	Deletes Attributes
	Updates Attributes
}

// ReplaceChange replaces old with new in-place
type ReplaceChange struct {
	Old Node
	New Node
}

// MoveBeforeChange means `Before` must be inserted before `Node`
type MoveBeforeChange struct {
	Node   Node // XXX Rename to After
	Before Node
}

// DeleteChange means `Node` must be removed
type DeleteChange struct {
	Node Node
}

// AddChange means `Node` must be placed at the end (may be moved
// later using MoveBeforeChange)
type AddChange struct {
	Parent Node
	Node   Node
}

// Apply an AttributeChange
func (c *AttributeChange) Apply(subject Subject) error {
	return subject.AttributeChange(c.Target, c.Adds, c.Deletes, c.Updates)
}

// Apply a ReplaceChange
func (c *ReplaceChange) Apply(subject Subject) error {
	return subject.Replace(c.Old, c.New)
}

// Apply a DeleteChange
func (c *DeleteChange) Apply(subject Subject) error {
	return subject.Delete(c.Node)
}

// Apply an AddChange
func (c *AddChange) Apply(subject Subject) error {
	return subject.Add(c.Node, c.Parent)
}

// Apply a MoveBeforeChange
func (c *MoveBeforeChange) Apply(subject Subject) error {
	return subject.InsertBefore(c.Before, c.Node)
}
