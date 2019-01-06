package vtree

import (
	"fmt"
	"reflect"
)

type ComponentRenderer func(*Element, *Context)

type Renderer struct {
	Handler ComponentRenderer
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
func (r *Renderer) RenderFor(e *Element, name string, context *Context) (result []*Element) {
	/*
	 * The tag containing the g-for will be duplicated,
	 * so effectively this can return nil, one or multiple
	 * elements. But not nodes, since it will duplicate
	 * the start tag.
	 *
	 * Also: set key / id
	 */
	value := context.Get(name) // XXX Check NotFound

	// must be an array of something
	// field.Kind is any of reflect.Array, reflect.Map, reflect.Slice, reflect.String:
	for i := 0; i < value.Len(); i++ {
		m := context.Mark()
		context.PushValue("_", value.Index(i))
		clone := e.Clone().(*Element)
		// "key"
		clone.ID = ElementID(fmt.Sprintf("%s-%d", clone.ID, i))
		delete(clone.Attributes, "g-for")

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

// Render the tree with root `e`, adding, cloning, removing nodes and handled
// g-<expressions> where necessary, leaving the original tree in tact.
func (r *Renderer) Render(e *Element, context *Context) []*Element {
	// render tree 'e' into a new tree, evaluating expressions,
	// with given context

	// if the node type is g-value itself, render into a text node directly.
	// e.g. <g-value g-value="myvalue">...</g-value>

	// a node can have multiple expressions, so immediate return isn't
	// entirely correct

	if gValue, ok := e.Attributes["g-for"]; ok {
		// g-for will recurse on itself for each itertion, which will
		// deal with g-value, g-if, g-class, etc.
		return r.RenderFor(e, gValue, context)
	}

	clone := e.Clone().(*Element)

	if gValue, ok := e.Attributes["g-if"]; ok {
		if !r.RenderIf(clone, gValue, context) {
			return nil
		}
	}

	if gValue, ok := clone.Attributes["g-class"]; ok {
		r.RenderClass(clone, gValue, context)
	}

	if gValue, ok := clone.Attributes["g-value"]; ok {
		// a g-value replaces all children, so don't don't recurse
		r.RenderValue(clone, gValue, context)
		return []*Element{clone}
	}

	// Don't recurse into Children on components. In stead,
	// call ComponentRenderer
	if e.IsComponent() {
		if r.Handler != nil {
			// Handler can decide if re-execution is actually necessary.
			// Alternatively, store copy of context on element, do
			// separate execute
			r.Handler(e, context)
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
