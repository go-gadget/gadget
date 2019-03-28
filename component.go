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

type ActionData struct {
	component *WrappedComponent
	node      vtree.Node
}
type SetAction struct {
	ActionData
	property string
	value    interface{}
}

func (a *ActionData) Run() {
}

func (a *SetAction) Run() {
	j.J("SetAction.Run() called", a.property)
	// update nodes?
}

type UserAction struct {
	ActionData
	handler string
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
	// Perhaps set doesn't need to happen immediately if Get() is also
	// intercepted..
	g.RawSetValue(key, val)
	g.Update <- &SetAction{
		ActionData: ActionData{component: g,
			node: nil},
		property: key,
		value:    val}
}

func (g *WrappedComponent) bindSpecials(node *vtree.Element) {
	// recusively do stuff
	for k, v := range node.Attributes {
		if k == "g-click" {
			vv := v
			f := func() {
				g.Update <- &UserAction{
					ActionData: ActionData{
						component: g,
						node:      node},
					handler: vv,
				}
			}
			node.Handlers[v] = f
		}
		if k == "g-bind" {
			j.J("bindSpecials", k, v)
			vv := v
			f := func(value string) {
				// should probably do type conversions, return something if fails
				j.J("Setter", vv, value)
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
		} else if val, ok := g.Gadget.CurrentRoute.Params[propName]; ok {
			props = append(props, &vtree.Variable{Name: propName, Value: reflect.ValueOf(val)})
		}
	}

	return props
}

/*
 * Called if a component tag is encountered (e.g. <my-comp>)
 *
 * Either the component is already mounted -> find it and find its variables, or
 * it needs to be created. Find it, create instance, etc.
 */
// rename to Render?
func (g *WrappedComponent) BuildDiff(props []*vtree.Variable) (res vtree.ChangeSet) {

	// collect changesets
	var cs []vtree.ChangeSet

	/*
	 * The componentElement can be a <router-view>, which means the component
	 * comes from the router.
	 * This component can/will be different based on the route, of course.
	 * Perhaps store/index it based on the route? E.g. router-view-profile, router-view-posts
	 */
	ComponentHandler := func(componentElement *vtree.Element) {
		for _, m := range g.Mounts {
			if m.HasComponent(componentElement) {
				Props := m.Component.ExtractProps(componentElement)
				changes := m.Component.BuildDiff(Props)
				cs = append(cs, changes)
				return
			}
		}

		// Build the component, if possible
		childcomps := g.Comp.Components()

		var builder Builder
		if builder = childcomps[componentElement.Type]; builder == nil {
			builder = g.Gadget.GlobalComponent(componentElement.Type)
		}

		if builder != nil {
			// builder is a ComponentBuilder, resulting in a Component, not a WrappedComponent
			wc := g.Gadget.NewComponent(builder)
			m := g.Mount(wc, componentElement)
			Props := m.Component.ExtractProps(componentElement)
			changes := m.Component.BuildDiff(Props)
			for _, ch := range changes {
				if ach, ok := ch.(*vtree.AddChange); ok && ach.Parent == nil {
					ach.Parent = m.Point
				}
			}
			cs = append(cs, changes)
		} else {
			j.J("Could not find / match component " + componentElement.Type)
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
