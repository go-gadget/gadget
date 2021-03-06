package gadget

import (
	"reflect"

	"github.com/go-gadget/gadget/j"
	"github.com/go-gadget/gadget/vtree"
)

type Handler func()

type Component interface {
	Init(*ComponentState)
	Props() []string
	Template() string
	Data() Storage
	Handlers() map[string]Handler // Actions ?
	Components() map[string]*ComponentFactory
}

type ComponentBuilder func() Component

type ComponentFactory struct {
	Name    string
	Builder ComponentBuilder
}

type BaseComponent struct {
	Storage Storage
	State   *ComponentState
}

func (b *BaseComponent) SetupStorage(storage Storage) {
	b.Storage = storage
}

func (b *BaseComponent) Init(s *ComponentState) {
	b.State = s
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

func (b *BaseComponent) Components() map[string]*ComponentFactory {
	return nil
}

type UserAction struct {
	component *ComponentInstance
	node      vtree.Node
	handler   string
}

func (a *UserAction) Run() {
	a.component.HandleEvent(a.handler)
}

type ComponentState struct {
	Registry       *Registry
	UnexecutedTree *vtree.Element
	ExecutedTree   *vtree.Element
	Mounts         []*Mount
}

type ComponentInstance struct {
	Comp      Component
	State     *ComponentState
	InnerTree vtree.NodeList // might become a map
}

func (ci *ComponentInstance) Init() {
	ci.Comp.Init(ci.State)
	ci.State.UnexecutedTree = vtree.Parse(ci.Comp.Template())
}

func (ci *ComponentInstance) RawSetValue(key string, val interface{}) {
	ci.Comp.Data().RawSetValue(key, val)
}

func (ci *ComponentInstance) SetValue(key string, val interface{}) {
	ci.RawSetValue(key, val)
}

func (ci *ComponentInstance) bindSpecials(node *vtree.Element) {
	// recusively do stuff
	for k, v := range node.Attributes {
		if k == "g-click" {
			vv := v
			f := func() {
				// One of the few actions that actually does stuff
				// But this should be reversed: A click on a control
				// creates an action (task). When handled, look up
				// any handlers for it.
				GetGadget(ci.State.Registry).Update <- &UserAction{
					component: ci,
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
				ci.RawSetValue(vv, value)
			}
			node.Setter = f
		}
	}

	for _, c := range node.Children {
		if el, ok := c.(*vtree.Element); ok {
			ci.bindSpecials(el)
		}
	}
}

func (ci *ComponentInstance) Execute(handler vtree.ComponentRenderer, props []*vtree.Variable) *vtree.Element {
	// This is actually a 2-step proces, just like builtin templates:
	// - compile, compiles text to tree
	// - render, evaluates expressions
	data := ci.Comp.Data()
	renderer := vtree.NewRenderer()
	renderer.Handler = handler
	renderer.InnerTree = ci.InnerTree

	for _, variable := range props {
		// context.PushValue(variable.Name, variable.Value)
		ci.RawSetValue(variable.Name, variable.Value.Interface())
	}

	// This makes the props available in acontext, for template rendering.
	// But not on the component itself
	context := data.MakeContext()
	// What to do if multi-element (g-for), or nil (g-if)? XXX
	// always wrap component in <div> ?
	tree := renderer.Render(ci.State.UnexecutedTree, context)[0]

	// we need to add a way for the "bridge" to call actions
	// this means just adding all Handlers() to all nodes,
	// to just the tree, or to already resolve the action
	// The bridge doesn't have access to the tree, only
	// to indivdual nodes
	// Not all nodes may have been cloned, duplication?
	ci.bindSpecials(tree)

	return tree
}

// Mount a comonent somewhere within this component, and store it.
func (ci *ComponentInstance) Mount(c *ComponentInstance, point *vtree.Element) *Mount {
	// probably needs lock

	// store node where mounted (or nil)
	mount := &Mount{Component: c, Point: point, ToBeRemoved: false}
	ci.State.Mounts = append(ci.State.Mounts, mount)
	// c.Mounted() hook?
	return mount
}

// ExtractProps checks which props a component accepts and fetches these from the elements attributes
func (ci *ComponentInstance) ExtractProps(componentElement *vtree.Element) []*vtree.Variable {
	var props []*vtree.Variable

	for _, propName := range ci.Comp.Props() {
		if val, ok := componentElement.Attributes[propName]; ok {
			props = append(props, &vtree.Variable{Name: propName, Value: reflect.ValueOf(val)})
		} else if val, ok := GetRouterState(ci.State.Registry).CurrentRoute.Params[propName]; ok {
			props = append(props, &vtree.Variable{Name: propName, Value: reflect.ValueOf(val)})
		}
	}

	return props
}

func (ci *ComponentInstance) BuildDiff(props []*vtree.Variable, rt *RouteTraverser) (res vtree.ChangeSet) {
	// collect changesets
	var cs []vtree.ChangeSet

	// Invoked when something component-like is encountered. Includes <router-view>
	ComponentHandler := func(componentElement *vtree.Element, innerElements vtree.NodeList) {
		// store, if anythingo
		// actually, not gonna work: <c>hello</c> -> <div><x></x><slot></slot></div>
		// <x> component wil already clear inner for slot
		var builder *ComponentFactory

		// Is this really about traversal? Or more about pre-call/render/?
		if t, ok := ci.Comp.(Traversable); ok {
			t.BeforeTraverse()
		}

		// First check if the component is already mounted. If so, it can be a router-view
		// that changes component, an existing component with different props
		for _, m := range ci.State.Mounts {
			if m.HasComponent(componentElement) {
				Props := m.Component.ExtractProps(componentElement)
				changes := m.Component.BuildDiff(Props, rt)
				cs = append(cs, changes)
				return
			}
		}

		// At this point, it was not an already mounted component
		childcomps := ci.Comp.Components()

		if builder = childcomps[componentElement.Type]; builder == nil {
			if cReg := GetComponentRegistry(ci.State.Registry); cReg != nil {
				builder = cReg.Get(componentElement.Type)
			}
		}

		if builder != nil {
			// builder is a ComponentComponentFactory, resulting in a Component, not a ComponentInstance
			cf := GetGadget(ci.State.Registry).NewComponent(builder)
			cf.InnerTree = innerElements
			m := ci.Mount(cf, componentElement)

			m.Name = builder.Name

			Props := m.Component.ExtractProps(componentElement)
			changes := m.Component.BuildDiff(Props, rt)
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

	// recusively calls BuildDiff through ComponentHandler
	tree := ci.Execute(ComponentHandler, props)

	var changes vtree.ChangeSet

	if ci.State.ExecutedTree == nil {
		changes = vtree.ChangeSet{&vtree.AddChange{Parent: nil, Node: tree}}
	} else {
		changes = vtree.Diff(ci.State.ExecutedTree, tree)
		for _, ch := range changes {
			if dch, ok := ch.(*vtree.DeleteChange); ok {
				if el, ok := dch.Node.(*vtree.Element); ok && el.IsComponent() {
					for _, m := range ci.State.Mounts {
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
	for _, m := range ci.State.Mounts {
		if m.ToBeRemoved {
			cs = append(cs, vtree.ChangeSet{&vtree.DeleteChange{Node: m.Component.State.ExecutedTree}})
			continue
			// call some hook?
		}
		FilteredMounts = append(FilteredMounts, m)
	}
	ci.State.Mounts = FilteredMounts
	ci.State.ExecutedTree = tree

	// reverse over cs, build res
	for i := len(cs) - 1; i >= 0; i-- {
		res = append(res, cs[i]...)
	}
	return res
}

func (ci *ComponentInstance) HandleEvent(event string) {
	ci.Comp.Handlers()[event]()
}

type ComponentRegistry struct {
	Components map[string]*ComponentFactory
}

func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{Components: make(map[string]*ComponentFactory)}
}

func (cr *ComponentRegistry) Register(Name string, Factory *ComponentFactory) {
	cr.Components[Name] = Factory
}

func (cr *ComponentRegistry) Get(Name string) *ComponentFactory {
	return cr.Components[Name]
}

func GetComponentRegistry(registry *Registry) *ComponentRegistry {
	if cr := registry.Get("components"); cr != nil {
		return cr.(*ComponentRegistry)
	}
	return nil
}

// A GeneratedComponent is a component that's dynamically built, not declaratively
type GeneratedComponent struct {
	BaseComponent
	gTemplate   string
	gComponents map[string]*ComponentFactory
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
func (g *GeneratedComponent) Components() map[string]*ComponentFactory {
	return g.gComponents
}

// GenerateComponent generates a Component (ComponentFactory) based on the supplied arguments
func GenerateComponentFactory(Name string, Template string, Components map[string]*ComponentFactory,
	Props []string) *ComponentFactory {
	return &ComponentFactory{
		Name: Name,
		Builder: func() Component {
			s := &GeneratedComponent{gTemplate: Template, gComponents: Components,
				gProps: Props}
			storage := NewMapStorage()
			s.SetupStorage(storage)
			return s
		},
	}
}
