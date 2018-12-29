package gadget

import (
	"github.com/go-gadget/gadget/j"
	"github.com/go-gadget/gadget/vtree"
)

type Action interface {
	Run()
	Component() *Component
	Node() vtree.Node
}

type Gadget struct {
	Chan       chan Action
	Components []*Component
	Trees      map[*Component]*vtree.Element
	Bridge     vtree.Subject
}

func NewGadget(bridge vtree.Subject) *Gadget {
	return &Gadget{
		Chan:   make(chan Action),
		Trees:  make(map[*Component]*vtree.Element),
		Bridge: bridge,
	}
}

func (g *Gadget) Mount(c *Component) {
	// Not sure if this really is mounting
	// probably needs lock
	g.Components = append(g.Components, c)
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

		 Click:
		 - gebruiker clickt op een button. In de backburner (?) wordt een UserAction geplaatst
		 - Backburner loop ontvangt action en zet deze in de queue
		 ... en of danwel, deze wordt opgepakt

		 De handler van de click wordt aangeroepen. Deze Set() variabelen (met mogelijk UI changes
		tot gevolg). Dus uit het afhandelen van de action worden nieuwe taken aangemaakt

		Eigenlijk wil je continue taken blijven afhandelen. Maar op een gegeven moment wil je er een klap
		op geven en de updates doorvoeren. Je wil niet (?) na iedere Set() de ui updaten.

		Wanneer is het werk "klaar"? In dit geval als de handler klaar is.

		Dus de Run() van Action produceert nieuwe taken

		Ember: MainLoop -> RunLoop
		Ember heeft meerdere priority Queues:
		 - sync, actions, route transitions, ...

		Nieuwe taken moeten consumed worden.
	*/

	var queue []Action
	wakeup := make(chan bool)

	go func() {

		for {
			msg := <-g.Chan

			size := len(queue)
			// lock?
			queue = append(queue, msg)

			if size == 0 {
				wakeup <- true
			}
		}
	}()

	// not sure if components can be added dynamically. For now
	// assume they're pre-built
	for _, c := range g.Components {
		tree := c.Render()
		g.Trees[c] = tree
		g.Bridge.Add(tree, nil)
	}

	for {
		workTrees := make(map[*Component]*vtree.Element)
		j.J("Sleeping until there's some work")
		<-wakeup
		j.J("There's work!", len(queue), queue[0])

		for len(queue) > 0 {

			work := queue[0]
			queue = queue[1:]

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
			newTree := c.Render()
			changes := vtree.Diff(tree, newTree)
			for _, c := range changes {
				j.J("Change ->", c)
			}
			j.J("That's a lot of changes:", len(changes))
			g.Trees[c] = newTree
			changes.ApplyChanges(g.Bridge)
		}
		j.J("Loop!")
	}
}
