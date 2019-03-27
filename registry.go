package gadget

import "sync"

type Registry struct {
	services map[string]interface{}
}

func (r *Registry) Register(key string, service interface{}) {
	r.services[key] = service
}

func (r *Registry) Get(key string) interface{} {
	return r.services[key]
}

var once sync.Once
var registry *Registry

func GetRegistry() *Registry {
	once.Do(func() {
		registry = &Registry{make(map[string]interface{})}
	})
	return registry
}
