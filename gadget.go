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
	Bridge vtree.Subject
	Queue  []Action
	Wakeup chan bool
	App    *WrappedComponent
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

func (g *Gadget) Mount(c *WrappedComponent) {
	g.App = c
	c.Update = g.Chan
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

	changes := g.App.BuildDiff(nil)

	changes.ApplyChanges(g.Bridge)

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
