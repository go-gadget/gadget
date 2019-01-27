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
	Data() interface{}
	Handlers() map[string]Handler // Actions ?
	Components() map[string]Builder
}

type BaseComponent struct {
	Storage interface{}
}

func (b *BaseComponent) SetupStorage(Storage interface{}) {
	b.Storage = Storage
}

func (b *BaseComponent) Init() {
}

func (b *BaseComponent) Props() []string {
	return []string{}
}

func (b *BaseComponent) Data() interface{} {
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

func (a *ActionData) Component() *WrappedComponent {
	// how to avoid this repetition for all actions?
	return a.component
}

func (a *ActionData) Node() vtree.Node {
	// how to avoid this repetition for all actions?
	return a.node
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
}

func (g *WrappedComponent) RawSetValue(key string, val interface{}) {

	// return err?
	// use resolve to handle errors?
	// look at how json handles this
	storage := reflect.ValueOf(g.Comp.Data()).Elem()
	field := storage.FieldByName(key)

	// ValType := reflect.TypeOf(val)
	// FieldType := reflect.TypeOf(field)

	ValVal := reflect.ValueOf(val)

	// fmt.Printf("%s -> %v - %v\n", key, FieldType, ValType)
	field.Set(ValVal)
	// switch ValType.Kind() {
	// case reflect.String:
	// 	field.Set(ValVal)
	// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	// 	// tt := typ.Elem()
	// 	field.Set(ValVal)
	// }
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

	context := vtree.MakeContext(data)

	for _, variable := range props {
		context.PushValue(variable.Name, variable.Value)
	}
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

func (g *WrappedComponent) BuildDiff(handler vtree.ComponentRenderer, props []*vtree.Variable) (res vtree.ChangeSet) {
	tree := g.Execute(handler, props)

	if g.ExecutedTree == nil {
		res = vtree.ChangeSet{&vtree.AddChange{Parent: nil, Node: tree}}
	} else {
		res = vtree.Diff(g.ExecutedTree, tree)
	}
	g.ExecutedTree = tree
	return res
}

func (g *WrappedComponent) HandleEvent(event string) {
	g.Comp.Handlers()[event](g.Update)
}
