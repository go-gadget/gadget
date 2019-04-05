package gadget

import (
	"reflect"

	"github.com/go-gadget/gadget/j"
	"github.com/go-gadget/gadget/vtree"
)

/*
 * Gadget only cares about the component interface:
 * - it must render into something (may not even care about
 *   the template?)
 *  - let runtime be in control of rendering (extra values, nested components)
 * - I must be able to get values and set values
 *   - values from the js side should be set (e.g. input fields),
 *   - values for rendering should be gotten
 * - I want to be able to call handlers on it
 *
 * When parsing a template I will get
 * - bound values (can be 2-way): g-bind or {{val}}
 * - event handlers (basically id's)
 *
 * Since we can't observe property changes, we need a different mechanism
 * for handling changes. An option could be a channel, that will put it on a
 * queue that will be unique-d
 *
 * A component defines props it takes. These props can then be passed when using
 * the component. E.g. title="Hello World".
 * To (generically) dynamically bind a prop, use g-bind:title="var"
 */
type Handler func(chan Action)

type Component interface {
	Init()
	Props() []string
	Template() string
	Data() Storage
	Handlers() map[string]Handler // Actions ?
	Components() map[string]Builder
}

type BaseComponent struct {
	Storage Storage
}

func (b *BaseComponent) SetupStorage(storage Storage) {
	b.Storage = storage
}

func (b *BaseComponent) Init() {
}

func (b *BaseComponent) Props() []string {
	return []string{}
}

// interface{} or Storage?
func (b *BaseComponent) Data() Storage {
	return b.Storage
}

func (b *BaseComponent) Template() string {
	return ""
}

func (b *BaseComponent) Handlers() map[string]Handler {
	return nil
}

func (b *BaseComponent) Components() map[string]Builder {
	return nil
}

type UserAction struct {
	component *WrappedComponent
	node      vtree.Node
	handler   string
}

func (a *UserAction) Run() {
	a.component.HandleEvent(a.handler)
}

type WrappedComponent struct {
	Comp           Component
	UnexecutedTree *vtree.Element
	ExecutedTree   *vtree.Element
	Update         chan Action
	Mounts         []*Mount
	Gadget         *Gadget
}

func (g *WrappedComponent) RawSetValue(key string, val interface{}) {
	g.Comp.Data().RawSetValue(key, val)
}

func (g *WrappedComponent) SetValue(key string, val interface{}) {
	g.RawSetValue(key, val)
}

func (g *WrappedComponent) bindSpecials(node *vtree.Element) {
	// recusively do stuff
	for k, v := range node.Attributes {
		if k == "g-click" {
			vv := v
			f := func() {
				// One of the few actions that actually does stuff
				// But this should be reversed: A click on a control
				// creates an action (task). When handled, look up
				// any handlers for it.
				g.Update <- &UserAction{
					component: g,
					node:      node,
					handler:   vv,
				}
			}
			node.Handlers[v] = f
		}
		if k == "g-bind" {
			j.J("bindSpecials", k, v)
			vv := v
			f := func(value string) {
				// should probably do type conversions, return something if fails
				// RawSetValue doesn't trigger a new Action
				g.RawSetValue(vv, value)
			}
			node.Setter = f
		}
	}

	for _, c := range node.Children {
		if el, ok := c.(*vtree.Element); ok {
			g.bindSpecials(el)
		}
	}
}

func (g *WrappedComponent) Execute(handler vtree.ComponentRenderer, props []*vtree.Variable) *vtree.Element {
	// This is actually a 2-step proces, just like builtin templates:
	// - compile, compiles text to tree
	// - render, evaluates expressions
	data := g.Comp.Data()
	renderer := vtree.NewRenderer()
	renderer.Handler = handler

	for _, variable := range props {
		// context.PushValue(variable.Name, variable.Value)
		g.RawSetValue(variable.Name, variable.Value.Interface())
	}

	// This makes the props available in acontext, for template rendering.
	// But not on the component itself
	context := data.MakeContext()
	// What to do if multi-element (g-for), or nil (g-if)? XXX
	// always wrap component in <div> ?
	tree := renderer.Render(g.UnexecutedTree, context)[0]

	// we need to add a way for the "bridge" to call actions
	// this means just adding all Handlers() to all nodes,
	// to just the tree, or to already resolve the action
	// The bridge doesn't have access to the tree, only
	// to indivdual nodes
	// Not all nodes may have been cloned, duplication?
	g.bindSpecials(tree)

	return tree
}

// Mount a comonent somewhere within this component, and store it.
func (g *WrappedComponent) Mount(c *WrappedComponent, point *vtree.Element) *Mount {
	// probably needs lock

	// store node where mounted (or nil)
	mount := &Mount{Component: c, Point: point, ToBeRemoved: false}
	g.Mounts = append(g.Mounts, mount)
	c.Update = g.Update
	// c.Mounted() hook?
	return mount
}

// ExtractProps checks which props a component accepts and fetches these from the elements attributes
func (g *WrappedComponent) ExtractProps(componentElement *vtree.Element) []*vtree.Variable {
	var props []*vtree.Variable

	for _, propName := range g.Comp.Props() {
		if val, ok := componentElement.Attributes[propName]; ok {
			props = append(props, &vtree.Variable{Name: propName, Value: reflect.ValueOf(val)})
		} else if val, ok := g.Gadget.RouterState.CurrentRoute.Params[propName]; ok {
			props = append(props, &vtree.Variable{Name: propName, Value: reflect.ValueOf(val)})
		}
	}

	return props
}

func (g *WrappedComponent) BuildDiff(props []*vtree.Variable, routeLevel int) (res vtree.ChangeSet) {

	// collect changesets
	var cs []vtree.ChangeSet

	// Invoked when something component-like is encountered. Includes <router-view>
	ComponentHandler := func(componentElement *vtree.Element) {
		var builder Builder

		for _, m := range g.Mounts {
			if m.HasComponent(componentElement) {
				// This will be true for a router-view, even if the inner component changes.
				if componentElement.Type == "router-view" {
					// PathID identifies the route. If it changes, we need to update the component and/or remove the old
					if crPathID := g.Gadget.RouterState.CurrentRoute.PathID(routeLevel); m.PathID != crPathID {
						cs = append(cs, vtree.ChangeSet{&vtree.DeleteChange{Node: m.Component.ExecutedTree}})
						builder = g.Gadget.GlobalComponent(componentElement.Type, routeLevel)
						if builder != nil {
							nc := g.Gadget.NewComponent(builder)
							m.Component = nc
							m.PathID = crPathID
						} else {
							m.ToBeRemoved = true
							return
						}
					}
				}
				Props := m.Component.ExtractProps(componentElement)
				changes := m.Component.BuildDiff(Props, routeLevel+1)
				cs = append(cs, changes)
				return
			}
		}

		// Build the component, if possible
		childcomps := g.Comp.Components()

		routeLevelInc := 0
		PathID := ""

		if builder = childcomps[componentElement.Type]; builder == nil {
			builder = g.Gadget.GlobalComponent(componentElement.Type, routeLevel)
			// XXX hacky, ugly
			// We need magic here to load the "index" route, if any.
			if builder != nil && componentElement.Type == "router-view" {
				routeLevelInc = 1
				PathID = g.Gadget.RouterState.CurrentRoute.PathID(routeLevel)
			}
		}

		if builder != nil {
			// builder is a ComponentBuilder, resulting in a Component, not a WrappedComponent
			wc := g.Gadget.NewComponent(builder)
			m := g.Mount(wc, componentElement)

			m.PathID = PathID

			Props := m.Component.ExtractProps(componentElement)
			changes := m.Component.BuildDiff(Props, routeLevel+routeLevelInc)
			for _, ch := range changes {
				if ach, ok := ch.(*vtree.AddChange); ok && ach.Parent == nil {
					ach.Parent = m.Point
				}
			}
			cs = append(cs, changes)
		} else {
			j.J("Could not find / match component "+componentElement.Type, builder, builder == nil, builder != nil)
		}
	}

	tree := g.Execute(ComponentHandler, props)

	var changes vtree.ChangeSet

	if g.ExecutedTree == nil {
		changes = vtree.ChangeSet{&vtree.AddChange{Parent: nil, Node: tree}}
	} else {
		changes = vtree.Diff(g.ExecutedTree, tree)
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
	}
	cs = append(cs, changes)
	var FilteredMounts []*Mount
	for _, m := range g.Mounts {
		if m.ToBeRemoved {
			continue
			// call some hook?
		}
		FilteredMounts = append(FilteredMounts, m)
	}
	g.Mounts = FilteredMounts
	g.ExecutedTree = tree

	// reverse over cs, build res
	for i := len(cs) - 1; i >= 0; i-- {
		res = append(res, cs[i]...)
	}
	return res
}

func (g *WrappedComponent) HandleEvent(event string) {
	g.Comp.Handlers()[event](g.Update)
}

// A GeneratedComponent is a component that's dynamically built, not declaratively
type GeneratedComponent struct {
	BaseComponent
	gTemplate   string
	gComponents map[string]Builder
	gProps      []string
}

// Props returns the BC's props, if any
func (g *GeneratedComponent) Props() []string {
	return g.gProps
}

// Template returns the BC's Template, if any
func (g *GeneratedComponent) Template() string {
	return g.gTemplate
}

// Components returns the BC's Components, if any
func (g *GeneratedComponent) Components() map[string]Builder {
	return g.gComponents
}

// GenerateComponent generates a Component (Builder) based on the supplied arguments
func GenerateComponent(Template string, Components map[string]Builder,
	Props []string) Builder {
	return func() Component {
		s := &GeneratedComponent{gTemplate: Template, gComponents: Components,
			gProps: Props}
		storage := NewMapStorage()
		s.SetupStorage(storage)
		return s
	}
}
