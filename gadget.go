package gadget

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-gadget/gadget/j"
	"github.com/go-gadget/gadget/vtree"
)

type Action interface {
	Run()
}

type Gadget struct {
	Chan        chan Action
	Bridge      vtree.Subject
	Queue       []Action
	Wakeup      chan bool
	Routes      Router
	App         *WrappedComponent
	RouterState *RouterState
}

func NewGadget(bridge vtree.Subject) *Gadget {
	g := &Gadget{
		Chan:        make(chan Action),
		Bridge:      bridge,
		Wakeup:      make(chan bool),
		RouterState: &RouterState{},
	}
	g.App = g.NewComponent(GenerateComponent("<div>App<router-view></router-view></div>", nil, nil))
	GetRegistry().Register("gadget", g)
	GetRegistry().Register("bridge", bridge)
	g.RouterState.Update = g.Chan
	return g
}

// A Builder is anything that creates s Component
type Builder func() Component

func (g *Gadget) Router(routes Router) {
	g.Routes = routes
	g.RouterState.Router = routes
	// So either we make this global, or we structurally pass Gadget around
	SetRouter(&routes)
	SetRouterState(g.RouterState)
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

// GlobalComponent attempts to map a globally registered component (e.g. routing)
func (g *Gadget) GlobalComponent(ElementType string, routeLevel int) Builder {
	// But since it's no longer router-unaware (because of the routelevel),
	// fix?
	// Delegate to Router, Store, ...
	if ElementType == "router-view" {
		if rm := g.RouterState.CurrentRoute.Get(routeLevel); rm != nil {
			return rm.Route.Component
		}
	} else if ElementType == "router-link" {
		return RouterLinkBuilder
	}
	return nil
}

// NewComponent creates a new WrappedComponent through the supplied Builder,
// calling relevant hooks and doing necessary initialization
func (g *Gadget) NewComponent(b Builder) *WrappedComponent {
	comp := &WrappedComponent{Comp: b(), Update: nil}
	comp.Gadget = g

	comp.Comp.Init()
	comp.UnexecutedTree = vtree.Parse(comp.Comp.Template())
	return comp
}

func (g *Gadget) SingleLoop() {

	// Just sync entire tree. We can optimize this later
	if g.App.ExecutedTree != nil {
		tree := g.App.ExecutedTree
		g.SyncState(tree)
	}

	for len(g.Queue) > 0 {
		// continue until queue is completely empty (could be infinite, so cap?)

		work := g.Queue[0]
		g.Queue = g.Queue[1:]

		fmt.Printf("Work found: %#v\n", work)
		work.Run()
	}

	changes := g.App.BuildDiff(nil, 0)

	changes.ApplyChanges(g.Bridge)
}

func (g *Gadget) MainLoop() {

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
			fmt.Println("Ready to read tasks")
			msg := <-g.Chan

			size := len(g.Queue)
			// lock?
			g.Queue = append(g.Queue, msg)

			if size == 0 {
				g.Wakeup <- true
			}
		}
	}()

	// Set initial route
	if g.Routes != nil {
		if url, err := url.Parse(g.Bridge.GetLocation()); err == nil {
			g.RouterState.TransitionToPath(url.Path)
		}
	}
	for {
		j.J("Loop!")
		g.SingleLoop()
		<-g.Wakeup
		j.J("Sleeping until there's some work")
	}
}
