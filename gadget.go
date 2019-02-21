package gadget

import (
	"github.com/go-gadget/gadget/j"
	"github.com/go-gadget/gadget/vtree"
)

type Action interface {
	Run()
	Component() *WrappedComponent
	Node() vtree.Node
}

type Gadget struct {
	Chan   chan Action
	Mounts []*Mount
	Bridge vtree.Subject
	Queue  []Action
	Wakeup chan bool
	App    *WrappedComponent
}

// Is a WrappedComponent actually a Mounted component?
// Is a Component actually (interface) Mountable?
type Mount struct {
	Component *WrappedComponent
	Point     *vtree.Element
	// Context is now (temp) stored on Mount which is ugly.
	// In stead, create separate queue of mounts to handle with their
	// contexts?
	Props       []*vtree.Variable
	ToBeRemoved bool
}

func (m *Mount) HasComponent(componentElement *vtree.Element) bool {
	if m.Point == nil {
		return false
	}
	return m.Point.Equals(componentElement)
}

func NewGadget(bridge vtree.Subject) *Gadget {
	return &Gadget{
		Chan:   make(chan Action),
		Bridge: bridge,
		Wakeup: make(chan bool),
	}
}

type Builder func() Component

// Move to component? NewWrappedComponent?
func (g *Gadget) BuildComponent(b Builder) *WrappedComponent {
	comp := &WrappedComponent{Comp: b(), Update: nil}
	comp.Comp.Init()
	comp.UnexecutedTree = vtree.Parse(comp.Comp.Template())
	comp.Gadget = g
	return comp
}

func (g *Gadget) Mount(c *WrappedComponent, point *vtree.Element) *Mount {
	// Not sure if this really is mounting
	// probably needs lock

	// store node where mounted (or nil)
	mount := &Mount{Component: c, Point: point, Props: nil, ToBeRemoved: false}
	g.Mounts = append(g.Mounts, mount)
	c.Update = g.Chan
	// c.Mounted() hook?
	return mount
}

func (g *Gadget) SyncState(Tree vtree.Node) {
	// Track which components actually have bindings
	el, ok := Tree.(*vtree.Element)
	if !ok {
		return
	}
	g.Bridge.SyncState(Tree)

	for _, c := range el.Children {
		g.SyncState(c)
	}
}

func (g *Gadget) SingleLoop() {

	for len(g.Queue) > 0 {
		// keep track of which trees have been synced
		syncedTrees := make(map[*WrappedComponent]bool)

		work := g.Queue[0]
		g.Queue = g.Queue[1:]

		c := work.Component()

		if !syncedTrees[c] {
			tree := c.ExecutedTree
			syncedTrees[c] = true

			// Get data before doing work
			g.SyncState(tree)
		}

		work.Run()
	}

	// Gadget no longer has mounts - it has one main
	// App component (possibly using a default)

	// -> changes := g.App.BuildDiff()

	// Newly created components are added to the end of g.Mounts,
	// so it can grow. Newly added components also need to be handled
	for i := 0; i < len(g.Mounts); i++ {
		m := g.Mounts[i]

		c := m.Component
		p := m.Point
		changes := c.BuildDiff(m.Props)

		// Check if diff shows components are removed. If so, mark them for removal
		for _, ch := range changes {
			if dch, ok := ch.(*vtree.DeleteChange); ok {
				if el, ok := dch.Node.(*vtree.Element); ok && el.IsComponent() {
					for _, m := range g.Mounts {
						if m.HasComponent(el) {
							m.ToBeRemoved = true
						}
					}
				}
			}
		}

		// This will add the rendered component tree to the component element
		if p != nil {
			for _, ch := range changes {
				if ach, ok := ch.(*vtree.AddChange); ok && ach.Parent == nil {
					ach.Parent = p
				}
			}
		}

		j.J("Changes end result", len(changes))
		changes.ApplyChanges(g.Bridge)
	}

	// Recursively call components for cleanup

	// Remove mountpoints that were marked for deletion
	// (make this a flag in stead?)
	var FilteredMounts []*Mount
	for _, m := range g.Mounts {
		if m.ToBeRemoved {
			continue
			// call some hook?
		}
		FilteredMounts = append(FilteredMounts, m)
	}
	g.Mounts = FilteredMounts

}

func (g *Gadget) MainLoop() {
	// Right now an update is triggered by sending something to the Chan channel.
	// If we'd do this on every SetValue, we'd get a lot of updates.
	// Better is to save / queue them and handle them in one go.
	// This probably means a go-routine needs to explicitly trigger an update
	/*
	 * Certain triggers cause the mainloop to loop. Normally the application
	 * is idle, except when:
	 * - a timer expires
	 * - an event (that's being listened to) triggers
	 * - IO ?
	 */

	go func() {

		for {
			msg := <-g.Chan

			size := len(g.Queue)
			// lock?
			g.Queue = append(g.Queue, msg)

			if size == 0 {
				g.Wakeup <- true
			}
		}
	}()

	for {
		j.J("Loop!")
		g.SingleLoop()
		<-g.Wakeup
		j.J("Sleeping until there's some work")
	}
}
