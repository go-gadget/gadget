package vtree


type DummyBridge struct {
}

func NewDomBridge() *DummyBridge {
	b := &DummyBridge{}
	return b
}


func (b *DummyBridge) SyncState(From Node) {}

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
