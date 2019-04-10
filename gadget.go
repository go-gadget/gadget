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
	Update      chan Action
	Bridge      vtree.Subject
	Queue       []Action
	Wakeup      chan bool
	App         *ComponentInstance
	RouterState *RouterState
	Traverser   *RouteTraverser
}

func NewGadget(bridge vtree.Subject) *Gadget {
	g := &Gadget{
		Update:      make(chan Action),
		Bridge:      bridge,
		Wakeup:      make(chan bool),
		RouterState: NewRouterState(),
	}
	g.App = g.NewComponent(GenerateComponentFactory("gadget.gadget.App", "<div>App<router-view></router-view></div>", nil, nil))
	GetRegistry().Register("gadget", g)
	GetRegistry().Register("bridge", bridge)
	g.RouterState.Update = g.Update
	return g
}

func (g *Gadget) Router(routes Router) {
	g.RouterState.Router = routes
	// So either we make this global, or we structurally pass Gadget around
	SetRouter(&routes)
	SetRouterState(g.RouterState)
}

func (g *Gadget) Mount(c *ComponentInstance) {
	g.App = c
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

func (g *Gadget) NewComponent(b *ComponentFactory) *ComponentInstance {
	state := &ComponentState{Update: g.Update, Gadget: g}
	comp := &ComponentInstance{Comp: b.Builder(), State: state}

	comp.Init()
	return comp
}

func (g *Gadget) SingleLoop() {

	// Just sync entire tree. We can optimize this later
	if g.App.State.ExecutedTree != nil {
		tree := g.App.State.ExecutedTree
		g.SyncState(tree)
	}

	for len(g.Queue) > 0 {
		// continue until queue is completely empty (could be infinite, so cap?)

		work := g.Queue[0]
		g.Queue = g.Queue[1:]

		fmt.Printf("Work found: %#v\n", work)
		work.Run()
	}

	g.Traverser = NewRouteTraverser(g.RouterState.CurrentRoute)
	changes := g.App.BuildDiff(nil, g.Traverser)

	changes.ApplyChanges(g.Bridge)
}

func (g *Gadget) MainLoop() {

	// Make sure there's aways a producer of actions
	go func() {
		for {
			// If nothing is feeding g.Wakeup, we get the "all goroutines are asleep"
			g.Wakeup <- false
			time.Sleep(100 * time.Second)
			j.J("Just slept 100 sec")
		}
	}()

	go func() {
		for {
			fmt.Println("Ready to read tasks")
			msg := <-g.Update

			size := len(g.Queue)
			// lock?
			g.Queue = append(g.Queue, msg)

			if size == 0 {
				g.Wakeup <- true
			}
		}
	}()

	// Set initial route
	if g.RouterState.Router != nil {
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
