package vtree

import (
	"fmt"
	"testing"
)

type Storage struct {
	MyString string
	MyInt    int
	MyBool   bool
	MyArray  []int64
}

func TestValueWithExistingNode(t *testing.T) {
	renderer := NewRenderer()

	e := El("div").A("g-value", "SomeValue").C(El("div").T("Hello"))
	ctx := &Context{}
	ctx.Push("SomeValue", "Test")
	res := renderer.Render(e, ctx)[0]

	if len(res.Children) != 1 {
		t.Errorf("Expected exactly one child but got %d", len(res.Children))
	}
	AssertTextNode(t, res.Children[0], "Test")
}
func TestValueExpression(t *testing.T) {
	renderer := NewRenderer()

	TestCases := map[string]struct {
		Value    interface{}
		Expected string
	}{
		"Test String g-value":  {"Hello World", "Hello World"},
		"Test Integer g-value": {1234, "1234"},
	}

	e := El("div").A("g-value", "MyValue")
	ctx := &Context{}
	for Name, TestCase := range TestCases {
		t.Run(Name, func(t *testing.T) {
			defer ctx.Pop(ctx.Mark())
			ctx.Push("MyValue", TestCase.Value)
			res := renderer.Render(e, ctx)[0]

			if len(res.Children) != 1 {
				t.Errorf("Expected element to have exactly one element, got %d", len(res.Children))
			}
			tNode, ok := res.Children[0].(*Text)
			if !ok {
				t.Errorf("Expected child node to be Text, got %v", res.Children[0])
			}
			if tNode.Text != TestCase.Expected {
				t.Errorf("Unexpected text value, got '%s'", tNode.Text)
			}
		})
	}
}

func TestIfFalseExpression(t *testing.T) {
	renderer := NewRenderer()
	e := El("div").A("g-if", "MyBool").C(El("div").T("child 1"), El("div").T("child 2"))

	res := renderer.Render(e, MakeContext(&Storage{MyBool: false}))

	if res != nil {
		t.Error("Expected node to disappear but it didn't")
	}
}

func TestIfTrueExpression(t *testing.T) {
	renderer := NewRenderer()
	e := El("div").A("g-if", "MyBool").C(El("div").T("child 1"), El("div").T("child 2"))

	res := renderer.Render(e, MakeContext(&Storage{MyBool: true}))[0]

	if res == nil {
		t.Error("Expected to get node back")
	}

	if len(res.Children) != 2 {
		t.Errorf("Expected a node with 2 children but got %d", len(res.Children))
	}
}

// Test if with nested value (not on same node), and possibly on same node
// <b g-if="foo" g-value="bar">

func TestForExpression(t *testing.T) {
	renderer := NewRenderer()
	e := El("div").A("g-for", "MyArray").T("Hello World")

	res := renderer.Render(e, MakeContext(&Storage{MyArray: []int64{1, 2, 3, 4}}))

	AssertElementCount(t, res, 4)

	for _, n := range res {
		AssertElement(t, n, "div")
		AssertTextNode(t, n.Children[0], "Hello World")
	}
}
func TestEmptyForExpression(t *testing.T) {
	renderer := NewRenderer()
	e := El("div").A("g-for", "MyArray").T("Hello World")

	res := renderer.Render(e, MakeContext(&Storage{MyArray: []int64{}}))

	if res != nil {
		t.Errorf("Expected no nodes at all, got %d", len(res))
	}
}

func TestForLoopElementID(t *testing.T) {
	renderer := NewRenderer()
	e := El("div").A("g-for", "MyArray").A("g-value", "_")

	res := renderer.Render(e, MakeContext(&Storage{MyArray: []int64{1, 2, 3, 4}}))

	// Currently generate id's based on original + index
	for i, el := range res {
		expID := ElementID(fmt.Sprintf("%s-%d", e.ID, i))
		if el.ID != expID {
			t.Errorf("Expected id %s, got %s", expID, el.ID)
		}
	}
}
func TestForExpressionWithContext(t *testing.T) {
	renderer := NewRenderer()
	e := El("div").A("g-for", "MyArray").C(El("div").A("g-value", "_"))

	// Does not work with intarray because of g-value
	ctx := &Context{}
	ctx.Push("MyArray", []string{"1", "2", "3", "4"})
	// res := renderer.Render(e, MakeContext(&Storage{MyArray: []int64{1, 2, 3, 4}}))
	res := renderer.Render(e, ctx)

	AssertElementCount(t, res, 4)

	AssertElement(t, res[2].Children[0].(*Element), "div")
	AssertTextNode(t, res[2].Children[0].(*Element).Children[0], "3")
}

func TestForValue(t *testing.T) {
	renderer := NewRenderer()
	e := El("div").A("g-for", "MyArray").A("g-value", "_").T("x")
	ctx := &Context{}
	ctx.Push("MyArray", []string{"a", "bc", "c"})
	res := renderer.Render(e, ctx)

	AssertElementCount(t, res, 3)

	AssertTextNode(t, res[2].Children[0], "c")
	if len(res[2].Children) != 1 {
		t.Errorf("Too many childnodes in forloop")
	}
}

func TestClass(t *testing.T) {
	renderer := NewRenderer()
	e := El("div").A("class", "some classes set").A("g-class", "MyClasses")
	ctx := &Context{}
	ctx.Push("MyClasses", "extra more")
	res := renderer.Render(e, ctx)

	AssertElementCount(t, res, 1)

	AssertAttribute(t, res[0], "class", "some classes set extra more")
}

func TestClassNoneSet(t *testing.T) {
	renderer := NewRenderer()
	e := El("div").A("g-class", "MyClasses")
	ctx := &Context{}
	ctx.Push("MyClasses", "extra more")
	res := renderer.Render(e, ctx)

	AssertElementCount(t, res, 1)

	AssertAttribute(t, res[0], "class", "extra more")
}

func TestIfValueClass(t *testing.T) {
	// if, value, class are all executed
	renderer := NewRenderer()
	e := El("div").A("g-class", "MyClass").A("g-if", "MyIf").A("g-value", "MyValue")
	res := renderer.Render(
		e,
		MakeContext(
			&struct {
				MyClass string
				MyIf    bool
				MyValue int
			}{"present", true, 42},
		),
	)

	AssertElementCount(t, res, 1)

	el := res[0]

	AssertElementAttributes(t, el, Attributes{"class": "present"})

	AssertTextNode(t, el.Children[0], "42")
}

func TestDeepNested(t *testing.T) {
	renderer := NewRenderer()
	tpl := `<div g-class="A"><ul g-if="B"><li g-for="C" g-value="_">x</li></ul></div>`
	tree := Parse(tpl)
	ctx := &Context{}
	ctx.Push("A", "test")
	ctx.Push("B", true)
	ctx.Push("C", []int{1})
	rendered := renderer.Render(tree, ctx)

	firstLi := rendered[0].Children[0].(*Element).Children[0].(*Element)
	tNode := firstLi.Children[0]
	if len(firstLi.Children) != 1 {
		t.Errorf("Too many textnodes in forloop, got %d", len(firstLi.Children))
	}
	AssertTextNode(t, tNode, "1")
}

func TestDeepNestedChange(t *testing.T) {
	renderer := NewRenderer()
	tpl := `<div><div g-class="A"><ul g-if="B"><li g-for="C" g-value="_">x</li></ul></div></div>`
	tree := Parse(tpl)
	ctx := &Context{}
	ctx.Push("A", "test")
	ctx.Push("B", true)
	ctx.Push("C", []int{1})

	// first render
	firstPass := renderer.Render(tree, ctx)

	ctx.Push("C", []int{1, 2})
	secondPass := renderer.Render(tree, ctx)

	changes := Diff(firstPass[0], secondPass[0])

	if len(changes) == 0 {
		t.Error("Expected at least one change")
	}
}

func TestBind(t *testing.T) {
	t.Run("Test regular case, long syntax", func(t *testing.T) {
		renderer := NewRenderer()
		e := El("div").A("g-bind:someprop", "somevalue")
		ctx := &Context{}
		ctx.Push("somevalue", "Hello World")
		res := renderer.Render(e, ctx)

		AssertElementCount(t, res, 1)

		AssertAttributeNotPresent(t, res[0], "g-bind:someprop")
		AssertAttribute(t, res[0], "someprop", "Hello World")
	})

	t.Run("Test regular case, short syntax", func(t *testing.T) {
		renderer := NewRenderer()
		e := El("div").A(":someprop", "somevalue")
		ctx := &Context{}
		ctx.Push("somevalue", "Hello World")
		res := renderer.Render(e, ctx)

		AssertElementCount(t, res, 1)

		AssertAttributeNotPresent(t, res[0], "g-bind:someprop")
		AssertAttribute(t, res[0], "someprop", "Hello World")
	})
	t.Run("Test overwrite", func(t *testing.T) {
		renderer := NewRenderer()
		e := El("div").A("someprop", "old value").A("g-bind:someprop", "somevalue")
		ctx := &Context{}
		ctx.Push("somevalue", "Hello World")
		res := renderer.Render(e, ctx)

		AssertElementCount(t, res, 1)

		AssertAttributeNotPresent(t, res[0], "g-bind:someprop")
		AssertAttribute(t, res[0], "someprop", "Hello World")
	})
	t.Run("Test no overwrite if missing", func(t *testing.T) {
		renderer := NewRenderer()
		e := El("div").A("someprop", "old value").A("g-bind:someprop", "somevalue")
		ctx := &Context{}
		res := renderer.Render(e, ctx)

		AssertElementCount(t, res, 1)

		AssertAttributeNotPresent(t, res[0], "g-bind:someprop")
		AssertAttribute(t, res[0], "someprop", "old value")
	})
	t.Run("Test mix with g-class", func(t *testing.T) {
		// If there's both a g-bind:class and g-class, the latter should "win" because it's more specific
		renderer := NewRenderer()
		e := El("div").A("g-bind:class", "somevalue").A("g-class", "specific-class")
		ctx := &Context{}
		ctx.Push("somevalue", "style1")
		ctx.Push("specific-class", "style2")
		res := renderer.Render(e, ctx)

		AssertElementCount(t, res, 1)

		AssertAttributeNotPresent(t, res[0], "g-bind:class")
		AssertAttributeNotPresent(t, res[0], "g-class")
		// g-class is additive
		AssertAttribute(t, res[0], "class", "style1 style2")
	})
}
