package gadget

import (
	"net/url"
	"time"

	"github.com/go-gadget/gadget/j"
	"github.com/go-gadget/gadget/vtree"
)

type Action interface {
	Run()
	Component() *WrappedComponent
	Node() vtree.Node
}

type Gadget struct {
	Chan         chan Action
	Bridge       vtree.Subject
	Queue        []Action
	Wakeup       chan bool
	Routes       Router
	RouteMatches []RouteMatch
	App          *WrappedComponent
}

func NewGadget(bridge vtree.Subject) *Gadget {
	return &Gadget{
		Chan:   make(chan Action),
		Bridge: bridge,
		Wakeup: make(chan bool),
		App:    NewComponent(GenerateComponent("<div>iHai</div>", nil, nil)),
	}
}

// A Builder is anything that creates s Component
type Builder func() Component

func (g *Gadget) Router(routes Router) {
	g.Routes = routes
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

	if g.Routes != nil {
		path := g.Bridge.GetLocation()
		if url, err := url.Parse(path); err == nil {
			g.RouteMatches = g.Routes.Parse(url.Path)
		}
	}
	// Make sure there's aways a producer of actions
	go func() {
		for {
			// If nothing is feeding g.Wakeup, we get the "all goroutines are asleep"
			g.Wakeup <- false
			time.Sleep(10 * time.Second)
			j.J("Just slept 10 sec")
		}
	}()
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
