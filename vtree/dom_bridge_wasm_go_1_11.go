// +build wasm,!go1.12

package vtree

import "syscall/js"

func jsHandler(handler callable) js.Func {
	cb := func([]js.Value) {
		handler()
	}

	return js.NewCallback(cb)
}
