package vtree

import (
	"testing"
)

func AssertValueString(t *testing.T, ctx *Context, name string, value string) {
	t.Helper()

	val := ctx.Get(name)
	if val == NotFound {
		t.Errorf("Could not get value for %s", name)
	}

	if val.String() != value {
		t.Errorf("Didn't get expected value '%s' for %s, got '%s'", value, name, val.String())
	}
}
func AssertValueInt(t *testing.T, ctx *Context, name string, value int64) {
	t.Helper()

	val := ctx.Get(name)
	if val == NotFound {
		t.Errorf("Could not get value for %s", name)
	}

	if val.Int() != value {
		t.Errorf("Didn't get expected value '%d' for %s, got '%d'", value, name, val.Int())
	}
}

func TestContextNotFound(t *testing.T) {
	ctx := &Context{}

	if v := ctx.Get("DoesNotExist"); v != NotFound {
		t.Errorf("Expected NotFound, got %v", v)
	}
}

func TestPush(t *testing.T) {
	ctx := &Context{}
	ctx.Push("Foo", 123)
	ctx.Push("Bar", "abc")

	if len(ctx.Variables) != 2 {
		t.Errorf("Expected 2 variables in context, got %d", len(ctx.Variables))
	}

	AssertValueString(t, ctx, "Bar", "abc")
	AssertValueInt(t, ctx, "Foo", 123)
}

func TestPushSame(t *testing.T) {
	ctx := &Context{}
	ctx.Push("Foo", 123)
	ctx.Push("Foo", "abc")

	if len(ctx.Variables) != 2 {
		t.Errorf("Expected 2 variables in context, got %d", len(ctx.Variables))
	}

	AssertValueString(t, ctx, "Foo", "abc")
}

func TestPushPop(t *testing.T) {
	ctx := &Context{}
	ctx.Push("Foo", 123)
	m := ctx.Mark()
	ctx.Push("Foo", "abc")
	ctx.Pop(m)

	if len(ctx.Variables) != 1 {
		t.Errorf("Expected 1 variable in context, got %d", len(ctx.Variables))
	}

	AssertValueInt(t, ctx, "Foo", 123)
}
