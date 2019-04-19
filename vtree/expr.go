package vtree

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-gadget/gadget/j"
)

type ComponentRenderer func(*Element, NodeList)

type Renderer struct {
	Handler   ComponentRenderer
	InnerTree NodeList
}

func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderValue replaces g-value="expr" with the value from the context
func (r *Renderer) RenderValue(e *Element, name string, context *Context) {
	delete(e.Attributes, "g-value")

	value := context.Get(name) // XXX check NotFound
	// "de-reference" pointer? Necessary? Test!
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	e.Children = nil
	// We rely on the string printing capabilities of the value,
	// !v.Type().Implements(fmtStringerType)
	e.T(fmt.Sprint(value))
}

// RenderIf returns the node with children if the g-if expression evaluates to true
func (r *Renderer) RenderIf(e *Element, name string, context *Context) bool {
	value := context.Get(name) // XXX check NotFound

	if !value.Bool() {
		return false
	}
	delete(e.Attributes, "g-if")
	return true
}

// RenderFor handles g-for, duplicating the node for each iteration
func (r *Renderer) RenderFor(e *Element, expression string, context *Context) (result []*Element) {
	/*
	 * The tag containing the g-for will be duplicated,
	 * so effectively this can return nil, one or multiple
	 * elements. But not nodes, since it will duplicate
	 * the start tag.
	 *
	 * Also: set key / id
	 *
	 * Syntax:
	 * g-for="variable" - iterates over variable, assings to _
	 * g-for="variable in variable" - iterates over last variable, assigns to first
	 */
	name := expression
	assign := "_"
	if strings.Contains(expression, " in ") {
		parts := strings.Split(expression, " in ")
		assign = strings.Trim(parts[0], " ")
		name = strings.Trim(parts[1], " ")
	}

	value := context.Get(name) // XXX Check NotFound

	// must be an array of something
	// field.Kind is any of reflect.Array, reflect.Map, reflect.Slice, reflect.String:
	delete(e.Attributes, "g-for")
	for i := 0; i < value.Len(); i++ {
		m := context.Mark()
		context.PushValue(assign, value.Index(i))
		clone := e.DeepClone(ElementID(fmt.Sprintf("%s-%d", e.ID, i))).(*Element)

		res := r.Render(clone, context)
		result = append(result, res...)

		context.Pop(m)
	}
	return result
}

/*
 * RenderClass renders the g-class expression. It takes as value a string
 * (and later array of strings) that gets added to the already defined classes
 *
 * vue has <div v-bind:class="{ active: isActive }"></div>
 * the nice thing about this is that the class is in the template/expression,
 * not in some opaque variable. It's obivious what class might get set.
 * an alternative could be g-class:classname=variable, but this makes
 * the class attribute a special case, it won't work any attribute like
 * v-bind does.
 *
 * For now, g-class with special behaviour.It may also support a map,
 * etc.
 *
 * 1. g-class="varname" -> adds varname to class
 * 2. g-class:alert="varname" -> sets class alert if varname evals true
 * 3. g-class="map[string]bool" -> sets keys in map that are true
 *
 * For now, only 1. is supported
 */
func (r *Renderer) RenderClass(e *Element, classes string, context *Context) {
	// are attributes clones?
	delete(e.Attributes, "g-class")

	value := context.Get(classes) // XXX Check NotFound

	ClassAttr, ok := e.Attributes["class"]
	if ok {
		ClassAttr += " " + value.String()
	} else {
		ClassAttr = value.String()
	}
	e.Attributes["class"] = ClassAttr
	/// XXX TODO deduplicate
}

// RenderBind scans the attributes in e for g-bind:<attr> or :<attr>
// looks up the value in the context and sets it as an attribute on the
// element. If not found, nothing will be set (so a default may persist)
func (r *Renderer) RenderBind(e *Element, context *Context) {
	for k, v := range e.Attributes {
		if strings.HasPrefix(k, "g-bind:") || strings.HasPrefix(k, ":") {
			attr := strings.SplitN(k, ":", 2)[1]
			if value := context.Get(v); value != NotFound {
				// For now attrs are always strings XXX
				e.Attributes[attr] = fmt.Sprint(value)
			} else {
				j.J("Could not get value for g-bind attr " + attr)
			}
			delete(e.Attributes, k)
		}
	}
}

// RenderSlot replaces the special <slot></slot> tag with the contents
// of the component it's part of
func (r *Renderer) RenderSlot(e *Element, context *Context) []*Element {
	// If there's no content, keep whatever <slot> currently contains (default)
	if len(r.InnerTree) != 0 {
		e.Children = make(NodeList, len(r.InnerTree))
		for i, ie := range r.InnerTree {
			e.Children[i] = ie
		}
	}
	return []*Element{e}
}

func (r *Renderer) Render(e *Element, context *Context) []*Element {
	return r.RenderComp(e, context, false)
}

// Render the tree with root `e`, adding, cloning, removing nodes and handled
// g-<expressions> where necessary, leaving the original tree in tact.
func (r *Renderer) RenderComp(e *Element, context *Context, skipComponent bool) []*Element {
	// render tree 'e' into a new tree, evaluating expressions,
	// with given context

	// if the node type is g-value itself, render into a text node directly.
	// e.g. <g-value g-value="myvalue">...</g-value>

	// a node can have multiple expressions, so immediate return isn't
	// entirely correct
	clone := e.Clone().(*Element)

	if gValue, ok := clone.Attributes["g-for"]; ok {
		// g-for will recurse on itself for each itertion, which will
		// deal with g-value, g-if, g-class, etc.
		return r.RenderFor(clone, gValue, context)
	}

	if gValue, ok := e.Attributes["g-if"]; ok {
		if !r.RenderIf(clone, gValue, context) {
			return nil
		}
	}

	// g-bind:attr or just ":attr". Since we don't know
	// what attr looks like, we need to iterate over all attributes
	// Also, bind goes before class, so g-class wins from g-bind:class
	r.RenderBind(clone, context)

	if gValue, ok := clone.Attributes["g-class"]; ok {
		r.RenderClass(clone, gValue, context)
	}

	if gValue, ok := clone.Attributes["g-value"]; ok {
		// a g-value replaces all children, so don't recurse
		r.RenderValue(clone, gValue, context)
		return []*Element{clone}
	}

	// g-bind should take the property and set it as attribute on the rendered
	// element XXX

	if e.Type == "slot" {
		r.RenderSlot(clone, context)
	}

	// Render the contents of the component, then call the component handler with it.
	if !skipComponent && e.IsComponent() {
		inner := r.RenderComp(e, context, true)
		j.J("Rendered comp "+e.Type, inner)
		clone.Children = nil
		if r.Handler != nil {
			m := context.Mark()
			r.Handler(clone, inner[0].Children)
			context.Pop(m)
		}
	}
	// XXX make this optional: deep vs. shallow
	Children := clone.Children
	clone.Children = nil

	for _, c := range Children {
		if el, ok := c.(*Element); ok {
			for _, cc := range r.Render(el, context) {
				if cc != nil {
					clone.C(cc)
				}
			}
		} else {
			clone.C(c)
		}
	}
	return []*Element{clone}
}
