package vtree

import (
	"reflect"
)

/*
 * Context should just hold a bunch of values of different types. Eventually in a nested/scoped
 * setup: a sub-context should lookup locally first, then consult it's parent. Each component/
 * loop/whatever gets its own context with optionally a link to a parent
 */
/*
 * Yes, inspired by text/template/exec.go
 */

var NotFound reflect.Value

type Variable struct {
	Name  string
	Value reflect.Value // better than interface{} ?
}

type Context struct {
	Variables []Variable
}

func MakeContext(vars ...Variable) *Context {
	ctx := &Context{}
	for _, v := range vars {
		ctx.PushValue(v.Name, v.Value)
	}
	return ctx
}

func (c *Context) PushValue(name string, value reflect.Value) {
	c.Variables = append(c.Variables, Variable{name, value})
}

func (c *Context) Push(name string, value interface{}) {
	c.PushValue(name, reflect.ValueOf(value))
}

func (c *Context) Mark() int {
	return len(c.Variables)
}

func (c *Context) Pop(mark int) {
	c.Variables = c.Variables[0:mark]
}

func (c *Context) Get(name string) reflect.Value {
	for i := len(c.Variables) - 1; i >= 0; i-- {
		if c.Variables[i].Name == name {
			return c.Variables[i].Value
		}
	}
	return NotFound
}
