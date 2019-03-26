package gadget

// taken from vtree.context_test - rewrite XXX
// func TestContextCreate(t *testing.T) {
// 	data := struct {
// 		Foo string
// 		Bar int
// 	}{"Hello World", 42}
// 	ctx := MakeContext(data)

// 	if len(ctx.Variables) != 2 {
// 		t.Errorf("Expected 2 variables in context, got %d", len(ctx.Variables))
// 	}

// 	AssertValueString(t, ctx, "Foo", "Hello World")
// 	AssertValueInt(t, ctx, "Bar", 42)
// }
