package vtree

type DummyBridge struct {
}

type BridgeBuilder func() Subject

var Builder BridgeBuilder

func NewDummyBridge() Subject {
	b := &DummyBridge{}
	return b
}

func init() {
	Builder = NewDummyBridge
}

func (b *DummyBridge) SyncState(From Node) {}

func (b *DummyBridge) GetLocation() string {
	return "/"
}

func (b *DummyBridge) AttributeChange(Target Node, Adds, Deletes, Updates Attributes) error {
	return nil
}

func (b *DummyBridge) Replace(old Node, new Node) error {
	return nil
}

func (b *DummyBridge) Add(n Node, parent Node) error {
	return nil
}

func (b *DummyBridge) Delete(el Node) error {
	return nil
}

func (b *DummyBridge) InsertBefore(before Node, after Node) error {
	return nil
}
