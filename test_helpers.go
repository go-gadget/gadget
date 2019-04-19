package gadget

import "github.com/go-gadget/gadget/vtree"

type DummyComponent struct {
	GeneratedComponent
	BoolVal     bool
	IntArrayVal []int
	StringVal   string
}

func MakeDummyFactory(Template string, Components map[string]*ComponentFactory, Props []string) *ComponentFactory {
	return MakeNamedDummyFactory("DummyComponent", Template, Components, Props)
}

func MakeNamedDummyFactory(Name string, Template string, Components map[string]*ComponentFactory, Props []string) *ComponentFactory {
	return &ComponentFactory{
		Name: Name,
		Builder: func() Component {
			s := &DummyComponent{
				GeneratedComponent: GeneratedComponent{gTemplate: Template,
					gComponents: Components, gProps: Props}}
			s.SetupStorage(NewStructStorage(s))
			return s
		}}
}

type TestBridge struct {
	AttributeChangeCount uint16
	ReplaceCount         uint16
	AddCount             uint16
	DeleteCount          uint16
	InsertBeforeCount    uint16
	SyncStateCount       uint16
}

func NewTestBridge() *TestBridge {
	return &TestBridge{}
}

func (t *TestBridge) Reset() {
	t.AttributeChangeCount = 0
	t.ReplaceCount = 0
	t.AddCount = 0
	t.DeleteCount = 0
	t.InsertBeforeCount = 0
	t.SyncStateCount = 0
}

func (t *TestBridge) GetLocation() string {
	return "/"
}

func (t *TestBridge) SetLocation(string) {
}

func (t *TestBridge) AttributeChange(Target vtree.Node, Adds, Deletes, Updates vtree.Attributes) error {
	t.AttributeChangeCount++
	return nil
}
func (t *TestBridge) Replace(old vtree.Node, new vtree.Node) error {
	t.ReplaceCount++
	return nil
}
func (t *TestBridge) Add(el vtree.Node, parent vtree.Node) error {
	t.AddCount++
	return nil
}
func (t *TestBridge) Delete(el vtree.Node) error {
	t.DeleteCount++
	return nil
}
func (t *TestBridge) InsertBefore(before vtree.Node, after vtree.Node) error {
	t.InsertBeforeCount++
	return nil
}
func (t *TestBridge) SyncState(from vtree.Node) {
	t.SyncStateCount++
}

func FlattenComponents(base *ComponentInstance) *vtree.Element {
	executed := base.State.ExecutedTree
	// DeepClone changes id's, don't want that
	ee := executed.Clone().(*vtree.Element) // assume assertion is valid
	ee.Children = nil

	// string-replace won't work with duplicated components
	for _, c := range executed.Children {
		cc := c.Clone()
		ee.Children = append(ee.Children, cc)

		if el, ok := cc.(*vtree.Element); ok {
			if el.IsComponent() {
				for _, m := range base.State.Mounts {
					if m.HasComponent(el) {
						nested := FlattenComponents(m.Component)
						el.Children = append(el.Children, nested)
					}
				}
			}
		}
	}

	return ee
}
