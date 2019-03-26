package gadget

import (
	"reflect"

	"github.com/go-gadget/gadget/j"
	"github.com/go-gadget/gadget/vtree"
)

type Storage interface {
	RawSetValue(key string, value interface{})
	RawGetValue(key string) interface{}
	MakeContext() *vtree.Context
}

type MapStorage struct {
	store map[string]interface{}
}

func (s *MapStorage) RawSetValue(key string, value interface{}) {
	s.store[key] = value
}

func (s *MapStorage) RawGetValue(key string) interface{} {
	return s.store[key]
}

func (s *MapStorage) MakeContext() *vtree.Context {
	ctx := &vtree.Context{}
	for k, v := range s.store {
		ctx.PushValue(k, reflect.ValueOf(v))
	}
	return ctx
}

func NewMapStorage() Storage {
	return &MapStorage{store: make(map[string]interface{})}
}

type StructStorage struct {
	Struct interface{}
}

// https://github.com/a8m/reflect-examples

func (s *StructStorage) RawSetValue(key string, value interface{}) {
	// return err?
	// use resolve to handle errors?
	// look at how json handles this
	storage := reflect.ValueOf(s.Struct).Elem()
	field := storage.FieldByName(key)

	// ValType := reflect.TypeOf(val)
	// FieldType := reflect.TypeOf(field)

	ValVal := reflect.ValueOf(value)

	// fmt.Printf("%s -> %v - %v\n", key, FieldType, ValType)
	if !field.IsValid() || !field.CanSet() {
		j.J("Could not RawSet field - not valid!", key)
	}
	field.Set(ValVal)
	// switch ValType.Kind() {
	// case reflect.String:
	// 	field.Set(ValVal)
	// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	// 	// tt := typ.Elem()
	// 	field.Set(ValVal)
	// }
}

func (s *StructStorage) RawGetValue(key string) interface{} {
	storage := reflect.ValueOf(s.Struct).Elem()
	field := storage.FieldByName(key)

	// Does this return the value of the field as interface{} ?
	return field.Interface()
}

func (s *StructStorage) MakeContext() *vtree.Context {
	ctx := &vtree.Context{}
	t := reflect.TypeOf(s.Struct)
	v := reflect.ValueOf(s.Struct)
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		ctx.PushValue(t.Field(i).Name, v.Field(i))
	}
	return ctx
}

func NewStructStorage(struc interface{}) Storage {
	return &StructStorage{Struct: struc}
}
