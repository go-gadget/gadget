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

// Make it accept a ...[]*Variable in stead?
// and add a PushStruct in stead?
func MakeContext(data interface{}) *Context {
	ctx := &Context{}

	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		ctx.PushValue(t.Field(i).Name, v.Field(i))
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
