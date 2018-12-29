package vtree

import (
	"testing"
)

func TestElementIDUnique(t *testing.T) {
	if El("div").ID == El("dev").ID {
		t.Error("Expected ID's on elements to differ")
	}
}
func TestBasic(t *testing.T) {
	tree := El("div").A("class", "group").C(El("button").A("class", "button").T("Click me"))

	if len(tree.Children) != 1 {
		t.Errorf("Expected 1 child, got %d in stead", len(tree.Children))
	}
}

func TestMultiChildren(t *testing.T) {
	tree := El("div").C(El("a"), El("b"))

	if len(tree.Children) != 2 {
		t.Errorf("Expected 2 children, got %d in stead", len(tree.Children))
	}
}
