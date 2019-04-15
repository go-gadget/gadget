package gadget

type Registry struct {
	services map[string]interface{}
}

func (r *Registry) Register(key string, service interface{}) {
	r.services[key] = service
}

func (r *Registry) Get(key string) interface{} {
	return r.services[key]
}

func NewRegistry() *Registry {
	registry := &Registry{make(map[string]interface{})}
	return registry
}
