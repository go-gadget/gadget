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
	Chan       chan Action
	Components []*WrappedComponent
	Trees      map[*WrappedComponent]*vtree.Element
	Bridge     vtree.Subject
	Queue      []Action
	Wakeup     chan bool
}

func NewGadget(bridge vtree.Subject) *Gadget {
	return &Gadget{
		Chan:   make(chan Action),
		Trees:  make(map[*WrappedComponent]*vtree.Element),
		Bridge: bridge,
		Wakeup: make(chan bool),
	}
}

type Builder func() Component

func (g *Gadget) BuildComponent(b Builder) *WrappedComponent {
	comp := &WrappedComponent{Comp: b(), Update: nil}
	comp.Comp.Init()
	comp.Tree = vtree.Parse(comp.Comp.Template())
	return comp
}

func (g *Gadget) Mount(c *WrappedComponent) {
	// Not sure if this really is mounting
	// probably needs lock
	g.Components = append(g.Components, c)
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

func (g *Gadget) RenderComponents() {
	// first render of components
	// Build the first tree. There's nothing to diff against, so force
	// an Add change on the entire tree
	for _, c := range g.Components {
		tree := c.Render(g.BuildCR(c))
		g.Trees[c] = tree
		g.Bridge.Add(tree, nil)
	}

}

func (g *Gadget) SingleLoop() {
	workTrees := make(map[*WrappedComponent]*vtree.Element)
	j.J("There's work!", len(g.Queue), g.Queue[0])

	for len(g.Queue) > 0 {

		work := g.Queue[0]
		g.Queue = g.Queue[1:]

		c := work.Component()

		if _, ok := workTrees[c]; !ok {
			tree := g.Trees[c]
			workTrees[c] = tree

			// Get data before doing work
			g.SyncState(tree)
		}

		work.Run()
	}
	// done looping, start updating
	for c := range workTrees {
		tree := g.Trees[c]
		newTree := c.Render(g.BuildCR(c))
		changes := vtree.Diff(tree, newTree)
		for _, c := range changes {
			j.J("Change ->", c)
		}
		j.J("That's a lot of changes:", len(changes))
		g.Trees[c] = newTree
		changes.ApplyChanges(g.Bridge)
	}

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

	g.RenderComponents()

	for {
		j.J("Sleeping until there's some work")
		<-g.Wakeup
		g.SingleLoop()
		j.J("Loop!")
	}
}

func (g *Gadget) BuildCR(c *WrappedComponent) vtree.ComponentRenderer {
	return func(e *vtree.Element, context *vtree.Context) {
		// find the component that 'e' is referring to. Could be defined
		// on the parent component to which we don't have access right now (can be fixed).
		//
		j.J("CR", e, context)
		childcomps := c.Comp.Components()
		j.J("ChildC", childcomps)
		cc, ok := childcomps[e.Type]
		if ok {
			ccc := cc()
			j.J("Yeah! Found it!", cc, ccc)
			// res := c.Render(g.BuildCR(ccc))
		}
	}
}
