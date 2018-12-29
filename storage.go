package gadget

type Storage interface {
	SetValue(key string, value interface{})
	GetValue(key string) interface{}
}
