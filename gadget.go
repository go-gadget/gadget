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
}

// Is a WrappedComponent actually a Mounted component?
// Is a Component actually (interface) Mountable?
type Mount struct {
	Component   *WrappedComponent
	Point       *vtree.Element
	ToBeRemoved bool
}

func NewGadget(bridge vtree.Subject) *Gadget {
	return &Gadget{
		Chan:   make(chan Action),
		Bridge: bridge,
		Wakeup: make(chan bool),
	}
}

type Builder func() Component

func (g *Gadget) BuildComponent(b Builder) *WrappedComponent {
	comp := &WrappedComponent{Comp: b(), Update: nil}
	comp.Comp.Init()
	comp.UnexecutedTree = vtree.Parse(comp.Comp.Template())
	return comp
}

func (g *Gadget) Mount(c *WrappedComponent, point *vtree.Element) {
	// Not sure if this really is mounting
	// probably needs lock

	// store node where mounted (or nil)
	g.Mounts = append(g.Mounts, &Mount{c, point, false})
	c.Update = g.Chan
	// c.Mounted() hook?
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

	// Newly created components are added to the end of g.Mounts,
	// so it can grow. Newly added components also need to be handled
	for i := 0; i < len(g.Mounts); i++ {
		m := g.Mounts[i]

		c := m.Component
		p := m.Point
		changes := c.BuildDiff(g.BuildCR(c))

		// Check if diff shows components are removed. If so, mark them for removal
		for _, ch := range changes {
			if dch, ok := ch.(*vtree.DeleteChange); ok {
				if el, ok := dch.Node.(*vtree.Element); ok && el.IsComponent() {
					for _, m := range g.Mounts {
						if m.Point != nil && m.Point.ID == el.ID {
							m.ToBeRemoved = true
						}
					}
				}
			}
		}

		if p != nil {
			for _, ch := range changes {
				if ach, ok := ch.(*vtree.AddChange); ok && ach.Parent == nil {
					ach.Parent = p
				}
			}
		}
		changes.ApplyChanges(g.Bridge)
	}

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

	g.SingleLoop()
	for {
		j.J("Sleeping until there's some work")
		<-g.Wakeup
		g.SingleLoop()
		j.J("Loop!")
	}
}

// BuildCR is the callback called when executing on components.
func (g *Gadget) BuildCR(c *WrappedComponent) vtree.ComponentRenderer {
	return func(componentElement *vtree.Element, context *vtree.Context) {

		// This can be optimized using a map. But since maps are not ordered,
		// we can't combine with g.Components
		for _, m := range g.Mounts {
			if m.Point != nil && m.Point.ID == componentElement.ID { // XXX equals()?
				return
			}
		}

		// Build the component, if possible
		childcomps := c.Comp.Components()
		if builder, ok := childcomps[componentElement.Type]; ok {
			// builder is a ComponentBuilder, resulting in a Component, not a WrappedComponent
			wc := g.BuildComponent(builder)
			g.Mount(wc, componentElement)
		}
	}
}
