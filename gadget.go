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
	Registry    *Registry
}

func NewGadget(bridge vtree.Subject) *Gadget {
	registry := NewRegistry()
	registry.Register("components", NewComponentRegistry())
	g := &Gadget{
		Registry:    registry,
		Update:      make(chan Action),
		Bridge:      bridge,
		Wakeup:      make(chan bool),
		RouterState: NewRouterState(registry),
	}
	g.App = g.NewComponent(GenerateComponentFactory("gadget.gadget.App", "<div>App<router-view></router-view></div>", nil, nil))
	g.Registry.Register("gadget", g)
	g.Registry.Register("bridge", bridge)
	g.RouterState.Update = g.Update
	return g
}

func (g *Gadget) Router(routes Router) {
	g.Registry.Register("router", &routes)
	g.Registry.Register("router-state", g.RouterState)
	// Make router register its components
	RegisterRouterComponents(g.Registry)

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

func GetGadget(registry *Registry) *Gadget {
	return registry.Get("gadget").(*Gadget)
}

func (g *Gadget) NewComponent(b *ComponentFactory) *ComponentInstance {
	state := &ComponentState{Registry: g.Registry}
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

		fmt.Printf("Work found: %#v, remain %d\n", work, len(g.Queue))
		work.Run()
		fmt.Printf("Done running\n")
	}

	g.Traverser = NewRouteTraverser(g.RouterState.CurrentRoute)

	fmt.Printf("New traverser created, level %d\n", g.Traverser.level)
	changes := g.App.BuildDiff(nil, g.Traverser)

	changes.ApplyChanges(g.Bridge)
	fmt.Println("===== Mounts after loop ======")
	DumpMounts(g.App, 0)
}

func (g *Gadget) MainLoop() {

	// Make sure there's aways a producer of actions
	go func() {
		for {
			// If nothing is feeding g.Wakeup, we get the "all goroutines are asleep"
			time.Sleep(100 * time.Second)
			j.J("Just slept 100 sec")
			g.Wakeup <- false
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
	if GetRouter(g.Registry) != nil {
		if url, err := url.Parse(g.Bridge.GetLocation()); err == nil {
			g.RouterState.TransitionToPath(url.Path)
		}
	}
	for {
		j.J("Loop!")
		g.SingleLoop()
		j.J("Sleeping until there's some work")
		<-g.Wakeup
	}
}
