package vtree

import (
	"testing"
)

func TestTrivialNoChange(t *testing.T) {
	one := El("div").SetID("1")
	other := El("div").SetID("1")

	res := Diff(one, other)

	if len(res) != 0 {
		t.Errorf("Expected elements to be equal, but got change %v", res)
	}
}

func AssertChangeCount(t *testing.T, changes ChangeSet, count int) {
	t.Helper()

	if count == 0 && len(changes) > 0 {
		t.Error("Expected elements not to be equal")
	}
	if len(changes) != count {
		t.Errorf("Expected %d changes, got %d", count, len(changes))
	}
}

func AssertNodeID(t *testing.T, OneID, OtherID ElementID) {
	t.Helper()
	if OneID != OtherID {
		t.Errorf("Expected ElementID %s, got %s", OneID, OtherID)
	}
}

func AssertReplaceChange(t *testing.T, change Change, OldID, NewID ElementID) {
	t.Helper()

	c, ok := change.(*ReplaceChange)

	if !ok {
		t.Errorf("Expected an ReplaceChange, got %T", change)
	}

	AssertNodeID(t, c.Old.GetID(), OldID)
	AssertNodeID(t, c.New.GetID(), NewID)
}

func AssertAddChange(t *testing.T, change Change, ID ElementID) {
	t.Helper()

	c, ok := change.(*AddChange)

	if !ok {
		t.Errorf("Expected an AddChange, got %T", change)
	}

	// Only works on Elements
	_, ok = c.Node.(*Element)

	if !ok {
		t.Errorf("Expected an Element, got %T", c.Node)
	}
	AssertNodeID(t, c.Node.GetID(), ID)
}

func AssertAddTextChange(t *testing.T, change Change, text string) {
	t.Helper()

	c, ok := change.(*AddChange)

	if !ok {
		t.Errorf("Expected an AddChange, got %T", change)
	}

	tNode, ok := c.Node.(*Text)

	if !ok {
		t.Errorf("Expected a Text, got %T", c.Node)
	}

	if tNode.Text != text {
		t.Errorf("Expected content '%s', got '%s'", text, tNode.Text)
	}
}

func AssertReplaceFromTextChange(t *testing.T, change Change, text string) {
	t.Helper()

	c, ok := change.(*ReplaceChange)

	if !ok {
		t.Errorf("Expected an ReplaceChange, got %T", change)
	}

	tOld, ok := c.Old.(*Text)

	if !ok {
		t.Errorf("Expected Old Node to be Text, got %T", c.Old)
	}

	if tOld.Text != text {
		t.Errorf("Expected content '%s', got '%s'", text, tOld.Text)
	}

}

func AssertReplaceToTextChange(t *testing.T, change Change, text string) {
	t.Helper()

	c, ok := change.(*ReplaceChange)

	if !ok {
		t.Errorf("Expected an ReplaceChange, got %T", change)
	}

	tNew, ok := c.New.(*Text)

	if !ok {
		t.Errorf("Expected New Node to be Text, got %T", c.New)
	}

	if tNew.Text != text {
		t.Errorf("Expected content '%s', got '%s'", text, tNew.Text)
	}

}

func AssertAddChangeParent(t *testing.T, change Change, ID ElementID) {
	t.Helper()

	c, ok := change.(*AddChange)

	if !ok {
		t.Errorf("Expected an AddChange, got %T", change)
	}

	AssertNodeID(t, c.Parent.GetID(), ID)
}

func AssertAddChangeNilParent(t *testing.T, change Change) {
	t.Helper()

	c, ok := change.(*AddChange)

	if !ok {
		t.Errorf("Expected an AddChange, got %T", change)
	}

	if c.Parent == nil {
		t.Errorf("Expected AddChange parent to be nil, got %v", c.Parent)
	}

}
func AssertDeleteChange(t *testing.T, change Change, ID ElementID) {
	t.Helper()

	c, ok := change.(*DeleteChange)

	if !ok {
		t.Error("Expected a DeleteChange")
	}

	AssertNodeID(t, c.Node.GetID(), ID)
}

func AssertMoveBeforeChange(t *testing.T, change Change, BeforeID, AfterID ElementID) {
	t.Helper()

	c, ok := change.(*MoveBeforeChange)

	if !ok {
		t.Error("Expected a MoveBeforeChange")
	}

	AssertNodeID(t, c.Node.GetID(), AfterID)
	AssertNodeID(t, c.Before.GetID(), BeforeID)
}

func AssertAttributeChange(t *testing.T, TargetID ElementID, change Change, Adds, Deletes, Updates Attributes) {
	t.Helper()

	c, ok := change.(*AttributeChange)
	if !ok {
		t.Error("Expected a MoveBeforeChange")
	}
	AssertNodeID(t, c.Target.GetID(), TargetID)

	if len(Deletes) != len(c.Deletes) {
		t.Errorf("Expected %d Deletes but got %d", len(Deletes), len(c.Deletes))
	}
	if len(Updates) != len(c.Updates) {
		t.Errorf("Expected %d Changes but got %d", len(Updates), len(c.Updates))
	}

	matchAttributes := func(name string, got, expected Attributes) {
		if len(got) != len(expected) {
			t.Errorf("%s: Expected %d but got %d", name, len(expected), len(got))
		}

		for k, expV := range expected {
			realV, ok := got[k]
			if !ok {
				t.Errorf("%s: Missing expected key %s", name, k)
				continue
			}
			if expV != realV {
				t.Errorf("%s: Key %s mismatch.Expected %s, got %s",
					name, k, expV, realV)
			}
		}
	}

	matchAttributes("Adds", c.Adds, Adds)
	matchAttributes("Deletes", c.Deletes, Deletes)
	matchAttributes("Updates", c.Updates, Updates)
}

func TestTrivialChange(t *testing.T) {
	one := El("div")
	other := El("p")

	changes := Diff(one, other)

	AssertChangeCount(t, changes, 1)
	AssertReplaceChange(t, changes[0], one.ID, other.ID)
}

func TestTextChange(t *testing.T) {
	one := El("div").SetID("1").T("old")
	other := El("div").SetID("1").T("new")

	changes := Diff(one, other)

	AssertChangeCount(t, changes, 1)
	// AssertReplaceChange(t, changes[0], one.ID, other.ID)
}
func TestAttributeAdded(t *testing.T) {
	one := El("div").SetID("1")
	other := El("div").SetID("1").A("class", "hello")

	changes := Diff(one, other)

	AssertChangeCount(t, changes, 1)
	AssertAttributeChange(t, one.GetID(), changes[0],
		Attributes{"class": "hello"},
		Attributes{},
		Attributes{})
}

func TestAttributeChanged(t *testing.T) {
	one := El("div").SetID("1").A("class", "welcome")
	other := El("div").SetID("1").A("class", "hello")

	changes := Diff(one, other)

	AssertChangeCount(t, changes, 1)
	AssertAttributeChange(t, one.GetID(), changes[0],
		Attributes{},
		Attributes{},
		Attributes{"class": "hello"})
}

func TestAttributeRemoved(t *testing.T) {
	one := El("div").SetID("1").A("class", "welcome")
	other := El("div").SetID("1")

	changes := Diff(one, other)

	AssertChangeCount(t, changes, 1)
	AssertAttributeChange(t, one.GetID(), changes[0],
		Attributes{},
		Attributes{"class": "welcome"},
		Attributes{})
}

/*
 * Tests on collections of elements
 */
func TestListEqual(t *testing.T) {
	one := NodeList{
		El("div").SetID("1"),
		El("p").SetID("2"),
	}
	other := NodeList{
		El("div").SetID("1"),
		El("p").SetID("2"),
	}

	changes := one.Diff(nil, other)
	AssertChangeCount(t, changes, 0)
}

func TestNodeAdded(t *testing.T) {
	one := NodeList{}
	other := NodeList{
		El("div").SetID("1"),
	}

	changes := one.Diff(nil, other)
	AssertChangeCount(t, changes, 1)
	AssertAddChange(t, changes[0], "1")
	AssertAddChangeNilParent(t, changes[0])
}

func TestNodeDeleted(t *testing.T) {
	one := NodeList{
		El("div").SetID("1"),
	}
	other := NodeList{}

	changes := one.Diff(nil, other)
	AssertChangeCount(t, changes, 1)
	AssertDeleteChange(t, changes[0], "1")
}
func TestListOrderSwapped(t *testing.T) {
	one := NodeList{
		El("div").SetID("1"),
		El("p").SetID("2"),
	}
	other := NodeList{
		El("p").SetID("2"),
		El("div").SetID("1"),
	}

	changes := one.Diff(nil, other)
	AssertChangeCount(t, changes, 1)
	AssertMoveBeforeChange(t, changes[0], "2", "1")
}

func TestListManyChanges(t *testing.T) {
	one := NodeList{
		El("div").SetID("1"),
		El("p").SetID("2"),
		El("p").SetID("3"),
	}
	other := NodeList{
		El("p").SetID("2"),
		El("div").SetID("1"),
		El("b").SetID("5"),
	}

	changes := one.Diff(nil, other)
	// 1 Delete, 1 Add, 2 reorders
	AssertChangeCount(t, changes, 4)
	AssertDeleteChange(t, changes[0], "3")
	AssertAddChange(t, changes[1], "5")
	AssertMoveBeforeChange(t, changes[2], "1", "5")
	AssertMoveBeforeChange(t, changes[3], "2", "1")
}

func TestNested(t *testing.T) {
	one := El("div").
		SetID("1").
		A("class", "a").
		C(
			El("div").SetID("2"),
			El("div").SetID("3"),
		)
	other := El("div").
		SetID("1").
		A("class", "b").
		C(
			El("div").SetID("3"),
			El("div").SetID("2"),
		)

	changes := Diff(one, other)

	AssertChangeCount(t, changes, 2)
	AssertAttributeChange(t, one.GetID(), changes[0],
		Attributes{},
		Attributes{},
		Attributes{"class": "b"})
	AssertMoveBeforeChange(t, changes[1], "3", "2")
}

func TestAddParent(t *testing.T) {
	one := El("div").SetID("1")
	two := El("div").SetID("1").C(El("div").SetID("2"))

	changes := Diff(one, two)

	AssertChangeCount(t, changes, 1)
	AssertAddChange(t, changes[0], "2")
	AssertAddChangeParent(t, changes[0], "1")
}

func TestAddText(t *testing.T) {
	one := El("div").SetID("1")
	two := El("div").SetID("1").T("Hello World")

	changes := Diff(one, two)

	AssertAddTextChange(t, changes[0], "Hello World")
}

func TestReplaceText(t *testing.T) {
	one := El("div").SetID("1").T("This is a test")
	two := El("div").SetID("1").T("Hello World")

	changes := Diff(one, two)

	change := changes[0]

	AssertReplaceFromTextChange(t, change, "This is a test")
	AssertReplaceToTextChange(t, change, "Hello World")

	rChange := change.(*ReplaceChange)

	// generated/internal id should remain the same
	if rChange.Old.GetID() != rChange.New.GetID() {
		t.Errorf("Expected id's to remain the same, but got %s %s",
			rChange.Old.GetID(), rChange.New.GetID())
	}
}

func TestReplaceToText(t *testing.T) {
	one := El("div").SetID("1").C(El("p").SetID("2"))
	two := El("div").SetID("1").T("Hello World")

	changes := Diff(one, two)

	AssertDeleteChange(t, changes[0], "2")
	AssertAddTextChange(t, changes[1], "Hello World")
}

func TestReplaceFromText(t *testing.T) {
	one := El("div").SetID("1").T("Hello World")
	two := El("div").SetID("1").C(El("p").SetID("2"))

	changes := Diff(one, two)

	AssertDeleteChange(t, changes[0], one.Children[0].GetID())
	AssertAddChange(t, changes[1], "2")
}
