package vtree

import (
	"strings"
	"syscall/js"

	"github.com/go-gadget/gadget/j"
)

// Make this a wasm_only file?

// A bridge implements the Subject interface
func init() {
	Builder = NewDomBridge
}

type DomBridge struct {
	Doc   js.Value
	Root  js.Value
	Nodes map[ElementID]js.Value
}

func NewDomBridge() Subject {
	// we can/should probably pass this in? Might even allow "mocking"
	// perhaps create an abstraction in general first.
	doc := js.Global().Get("document")
	root := doc.Call("getElementById", "gadget-content")

	b := &DomBridge{}
	b.Doc = doc
	b.Root = root
	b.Nodes = make(map[ElementID]js.Value)
	return b
}

func (b *DomBridge) createElement(node Node) js.Value {

	el := node.(*Element)

	e := b.Doc.Call("createElement", el.Type)
	b.Nodes[el.GetID()] = e
	for attr, value := range el.Attributes {
		if attr == "class" {
			attr = "className"
		}
		e.Set(attr, value)
	}

	e.Set("id", string(el.ID))

	b.HandleSpecialAttributes(node, el.Attributes)
	return e
}

func (b *DomBridge) SyncState(From Node) {
	el := From.(*Element)
	e := b.Nodes[From.GetID()]

	val := e.Get("value").String()
	if el.Setter != nil {
		el.Setter(val)
	}
}

func (b *DomBridge) HandleSpecialAttribute(Target Node, action string, value string) {
	// we need to register an onclick, install a handler
	// and call something on a component. Just a attr/value and
	// node/element isn't sufficnent, so we need to pre-handle things.
	// Basically we need a func to call

	// if it has attributes it must be an element
	el := Target.(*Element)
	e := b.Nodes[Target.GetID()]

	if action == "click" {
		handler, ok := el.Handlers[value]

		if !ok {
			return
		}

		cb := func(i []js.Value) {
			j.J("!")
			handler()
		}
		// value is not unique enough
		// set onclick etc

		// onclick if action == "click" XXX
		e.Set("onclick", js.NewCallback(cb))
		j.J("onclick set to ", value)
	}
}

func (b *DomBridge) HandleSpecialAttributes(Target Node, attrs Attributes) {
	for k, v := range attrs {
		if strings.HasPrefix(k, "g-") {
			b.HandleSpecialAttribute(Target, k[2:], v)
		}
	}
}

func (b *DomBridge) AttributeChange(Target Node, Adds, Deletes, Updates Attributes) error {
	e := b.Nodes[Target.GetID()]

	// Is it likely that a g- attribute gets dynamically added (resulting in a change)?
	b.HandleSpecialAttributes(Target, Adds)
	for attr, value := range Adds {
		// duplication with createElement
		// skip g-*
		if attr == "class" {
			attr = "className"
		}
		e.Set(attr, value)
	}

	b.HandleSpecialAttributes(Target, Updates)
	for attr, value := range Updates {
		// duplication with createElement
		// skip g-*
		if attr == "class" {
			attr = "className"
		}
		e.Set(attr, value)
	}

	for attr, _ := range Deletes {
		// but here className won't work..?
		if strings.HasPrefix(attr, "g-") {
			// XXX Also: delete node -> cleanup
			// b.ClearSpecialAttribute(e, attr, value)
		}
		e.Call("removeAttribute", attr)
	}
	return nil
}

func (b *DomBridge) Replace(old Node, new Node) error {
	// if it's a Text and only text changed, do an update
	t1, ok1 := old.(*Text)
	t2, ok2 := new.(*Text)
	if ok1 && ok2 && t1.Text != t2.Text {
		// t1.GetID() == t2.GetID() at this point
		e := b.Nodes[t1.GetID()]
		e.Set("nodeValue", t2.Text)
		return nil
	}
	// replace == delete, add
	oldE := b.Nodes[old.GetID()]
	newE := b.createElement(new)
	p := oldE.Get("parentElement")

	p.Call("replaceChild", oldE, newE)
	return nil
}

func (b *DomBridge) getParent(parent Node) js.Value {
	p := b.Root
	if parent != nil {
		p, ok := b.Nodes[parent.GetID()]
		if !ok {
			panic("Parent " + parent.GetID() + " not found in map")
		}
		return p
	}

	return p
}
func (b *DomBridge) Add(n Node, parent Node) error {
	// parent is 100% sure an element, can't be Text
	p := b.getParent(parent)

	t, ok := n.(*Text)
	if ok {
		e := b.Doc.Call("createTextNode", t.Text)
		b.Nodes[t.GetID()] = e
		p.Call("appendChild", e)

		return nil
	}
	el := n.(*Element)

	e := b.createElement(n)

	// fmt.Printf("Creating %s id %s parent %v\n", el.Type, el.ID, p)

	p.Call("appendChild", e)

	for _, c := range el.Children {
		b.Add(c, el)
	}
	return nil
}

func (b *DomBridge) Delete(el Node) error {
	child := b.Nodes[el.GetID()]
	p := child.Get("parentElement")

	p.Call("removeChild", child)

	return nil
}

func (b *DomBridge) InsertBefore(before Node, after Node) error {
	nBefore := b.Nodes[before.GetID()]
	nAfter := b.Nodes[after.GetID()]
	p := nAfter.Get("parentElement")

	p.Call("insertBefore", nBefore, nAfter)
	return nil
}
