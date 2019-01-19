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
	Mounts []*Mount // <- Mounts ?
	Bridge vtree.Subject
	Queue  []Action
	Wakeup chan bool
}

// Is a WrappedComponent actually a Mounted component?
// Is a Component actually (interface) Mountable?
type Mount struct {
	Component *WrappedComponent
	Point     *vtree.Element
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
	g.Mounts = append(g.Mounts, &Mount{c, point})
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
	workTrees := make(map[*WrappedComponent]*vtree.Element)

	for len(g.Queue) > 0 {
		j.J("There's work!", len(g.Queue), g.Queue[0])

		work := g.Queue[0]
		g.Queue = g.Queue[1:]

		c := work.Component()

		if _, ok := workTrees[c]; !ok {
			tree := c.ExecutedTree
			workTrees[c] = tree

			// Get data before doing work
			g.SyncState(tree)
		}

		work.Run()
	}

	var handled []*Mount

	for len(g.Mounts) > 0 {
		m := g.Mounts[0]
		handled = append(handled, m)

		g.Mounts = g.Mounts[1:]

		c := m.Component
		p := m.Point
		changes := c.BuildDiff(g.BuildCR(c))

		// See if components were removed. If so, remove them from mounts
		for _, ch := range changes {
			if dch, ok := ch.(*vtree.DeleteChange); ok {
				if el, ok := dch.Node.(*vtree.Element); ok && el.IsComponent() {
					for _, m := range g.Mounts {
						var newMounts []*Mount
						if m.Point.ID != el.ID {
							newMounts = append(newMounts, m)
						}
						g.Mounts = newMounts
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

	g.Mounts = handled

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
	return func(e *vtree.Element, context *vtree.Context) {
		// find the component that 'e' is referring to. Could be defined
		// on the parent component to which we don't have access right now (can be fixed).
		//
		childcomps := c.Comp.Components()
		cc, ok := childcomps[e.Type]

		// This can be optimized using a map. But since maps are not ordered,
		// we can't combine with g.Components
		for _, m := range g.Mounts {
			if m.Point.ID == e.ID {
				return
			}
		}
		if ok {
			// cc is a ComponentBuilder, resulting in a Copmonent, not a WrappedComponent
			wc := g.BuildComponent(cc)
			// e, or e's parent?
			g.Mount(wc, e)
		}
	}
}
