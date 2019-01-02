package vtree

import "testing"

func TestSimpleParse(t *testing.T) {
	el := Parse("<div></div>")

	AssertElement(t, el, "div")
}

func TestAttributeParse(t *testing.T) {
	el := Parse(`<div class="alert" data="foo"></div>`)

	AssertElementAttributes(t, el,
		Attributes{"class": "alert", "data": "foo"})
}

func TestTextSimpleParse(t *testing.T) {
	el := Parse("<div>Hello World</div>")

	if len(el.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(el.Children))
	}

	tNode, ok := el.Children[0].(*Text)

	if !ok {
		t.Errorf("Expected first child to be Text, got %T", el.Children[0])
	}
	if tNode.Text != "Hello World" {
		t.Errorf("Expected text to match '%s', got '%s'", "Hello World",
			tNode.Text)
	}
}

func TestSplitTextParse(t *testing.T) {
	el := Parse("<div>Hello <i>GO</i> World</div>")

	if len(el.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(el.Children))
	}

	AssertTextNode(t, el.Children[0], "Hello ")
	AssertTextNode(t, el.Children[2], " World")

	// test middle Node
	eNode, ok := el.Children[1].(*Element)

	if !ok {
		t.Errorf("Expected second child to be Element, got %T",
			el.Children[1])
	}

	if len(eNode.Children) != 1 {
		t.Errorf("Expected second child to have 1 child, got %d in stead",
			len(eNode.Children))
	}

	AssertTextNode(t, eNode.Children[0], "GO")
}

func TestComponent(t *testing.T) {
	el := Parse("<div><my-component></my-component></div>")
	if len(el.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(el.Children))
	}

	AssertComponentNode(t, el.Children[0], "my-component")
}
