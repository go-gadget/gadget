package vtree

import (
	"strings"
	"testing"
)

// AssertElement asserts that el is of the specified type
func AssertElement(t *testing.T, el *Element, Type string) {
	t.Helper()

	if el.Type != Type {
		t.Errorf("Expected <%s> tag, got %s in stead", Type, el.Type)
	}
}

// AssertElementAttributes asserts that element has exactly the specified
// attributes, excluding g-* expression attributes
func AssertElementAttributes(t *testing.T, el *Element, Attributes Attributes) {
	t.Helper()

	attrCount := 0
	for k := range el.Attributes {
		if !strings.HasPrefix(k, "g-") {
			attrCount++
		}
	}
	if attrCount != len(Attributes) {
		t.Errorf("Amount of attributes doesn't match, expected %d but got %d",
			len(Attributes), attrCount)
	}

	for k, v := range Attributes {
		gotV, ok := el.Attributes[k]

		if !ok {
			t.Errorf("Didn't get expected attribute/value %s->%s",
				k, v)
		} else if v != gotV {
			t.Errorf("Value for %s didn't match, expected %s but got %s",
				k, v, gotV)
		}
	}
}

// AssertTextNode asserts that node is a TextNode with the specified string content
func AssertTextNode(t *testing.T, node Node, content string) {
	t.Helper()

	tNode, ok := node.(*Text)
	if !ok {
		t.Errorf("Expected node to be Text, got %T", node)
	}

	if tNode.Text != content {
		t.Errorf("Expected text to match '%s', got '%s'", content,
			tNode.Text)

	}
}
