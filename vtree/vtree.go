package vtree

import (
	"fmt"
	"strconv"
	"strings"
)

/*
 * Idea:
 * Build the following tree structure (explicitly or from template) from elements.
 *
 * An element:
 * - has attributes (class, id, etc)
 * - has children (child nodes, text)
 * - may contain "logic"? How to express this? E.g. v-for, v-value
 *
 * Sample expression:
 * el('div', class="foo" <- ?).attr("foo", "bar").attr("class", "button").children(c1, c2)
 *
 * Let's go for the chainy type definition for now
 *
 * Do we need id's on elements to identify them uniquely? We can just generate
 * id's and optionally allow the caller to set an id. But what if the caller
 * sets a non-unique id? We can detect this probably?
 *
 * ALternatively, we can keep the DOM element linked to the tree. Each time we
 * add a node we know the corresponding dom element, no need for lookups.
 *
 * Alternatively use data-<something-id, but I assume lookups won't be as fast.
 * Use internal id's to track diffs
 */

// ElementID uniquely enumerates ID's
type ElementID string

var _id int

func nextID() ElementID {
	_id++
	return ElementID(strconv.Itoa(_id))
}

type Attributes map[string]string

/*
 * A "normal" element has attributes, children (element or text), a type, etc.
 *
 * A text element only has text. It's quite different
 * from a "normal" element and hardly shares any functions, except:
 * - ToString()
 * - Equal(), perhaps?
 */
// Element describes an element (node) in the tree. It has attributes and
// children

type Node interface {
	GetID() ElementID
	ToString() string
	Equals(Node) bool
	Clone() Node
	DeepClone(ElementID) Node
}

type callable func()

// An Element node is a regular old html element, e.g. <div>
type Element struct {
	ID         ElementID // not sure if this is a good idea. E.g. rerunning for-loops shouldn't create new elements
	Type       string
	Attributes Attributes
	Children   NodeList
	Handlers   map[string]callable
	Setter     func(string)
}

func (e *Element) IsComponent() bool {
	return strings.Contains(e.Type, "-") && !strings.HasPrefix(e.Type, "g-")
}

// A Text node contains the text within an Element node. It doesn't have much special properties
type Text struct {
	ID   ElementID
	Text string
}

func (t *Text) GetID() ElementID {
	return t.ID
}

func (t *Text) ToString() string {
	return t.Text
}

func (t *Text) Clone() Node {
	// XXX Needs test LHF
	return &Text{t.ID, t.Text}
}

func (t *Text) DeepClone(ElementID) Node {
	// XXX Needs test LHF
	return &Text{t.ID, t.Text}
}

func (t *Text) Equals(other Node) bool {
	tt, ok := other.(*Text)
	if !ok {
		return false
	}

	if t.Text != tt.Text {
		return false
	}

	return t.GetID() == tt.GetID()
}

type NodeList []Node

// El constructs an element of a specific type
func El(Type string) *Element {
	return &Element{ID: nextID(), Type: Type, Attributes: make(Attributes), Handlers: make(map[string]callable)}
}

func (el *Element) Clone() Node {
	// Don't deep-copy, for now. In some cases (read-only?), a clone might not even be necessary
	// Clone children, link children? How to replace child with clone?
	// XXX Needs test LHF

	attrClone := make(Attributes)

	for atName, atVal := range el.Attributes {
		attrClone[atName] = atVal
	}

	// XXX test/assert attributes are copied
	return &Element{ID: el.ID,
		Type:       el.Type,
		Attributes: attrClone,
		Handlers:   el.Handlers,
		Children:   el.Children}
}

func (el *Element) DeepClone(newID ElementID) Node {
	// XXX needs test
	elClone := el.Clone().(*Element)

	elClone.ID = newID

	var childrenClones NodeList

	for _, c := range elClone.Children {
		cClone := c.DeepClone(ElementID(fmt.Sprintf("%s-%s", newID, c.GetID())))

		childrenClones = append(childrenClones, cClone)
	}

	elClone.Children = childrenClones

	return elClone
}

func (el *Element) GetID() ElementID {
	return el.ID
}

func (t *Element) Equals(other Node) bool {
	tt, ok := other.(*Element)
	if !ok {
		return false
	}
	if t.GetID() != tt.GetID() {
		return false
	}

	return t.Type == tt.Type
}
func (el *Element) SetID(id ElementID) *Element {
	// allow override of ID
	el.ID = id
	return el
}

// A adds an Attr to an element
func (el *Element) A(key string, value string) *Element {
	// make it a map? Overwrite previous?
	el.Attributes[key] = value
	return el
}

// T sets text on an element.
func (el *Element) T(text string) *Element {
	// An element holding text probably shouldn't have children?
	// Or should text itself also be an element?
	t := &Text{Text: text}
	t.ID = ElementID(string(el.ID) + "-" + strconv.Itoa(len(el.Children)))
	el.C(t)
	return el
}

// C adds one or more children
func (el *Element) C(child ...Node) *Element {
	el.Children = append(el.Children, child...)
	return el
}

// not used. Perhaps replace with string_bridge?
func (el *Element) ToString() string {
	res := "<" + el.Type
	res += fmt.Sprintf(` data-ID="%s"`, el.GetID())

	for k, v := range el.Attributes {
		res += " " + k + "=\"" + v + "\""
	}

	res += ">"

	for _, c := range el.Children {
		res += c.ToString()
	}
	res += "</" + el.Type + ">"

	return res
}
