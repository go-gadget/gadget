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

func TestIsComponentComponentElement(t *testing.T) {
	if !El("my-component").IsComponent() {
		t.Error("Expected IsComponent to be true on 'my-component'")
	}
}
func TestIsComponentGElement(t *testing.T) {
	if El("g-my-component").IsComponent() {
		t.Error("Expected IsComponent not to be true on 'g-my-component'")
	}
}

func TestIsComponentElement(t *testing.T) {
	if El("div").IsComponent() {
		t.Error("Expected IsComponent not to be true on 'div'")
	}
}
