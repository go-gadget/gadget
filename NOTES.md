# Notes, Todos

## Small tasks
- implement actual "serve" subcommand on gadget command
- serve wasm_exec js from GO directly (since version specific)
- serve index.html from .. somewhere?
- implement forloop-vars (first, last, index, number, even, odd)
- implement forloop variables ("i in sequence")
- g-if: test generic truthyness
- g-for: work on "sequences" (strings, maps, arrays, something implementing some interface?)
- generic attribute support (e.g. href)
- test Text node stuff


## Bigger stuff
- a component (and its template/tree) probably always needs to be
  wrapped (?) For now we assume the Render always returns a single element
- figure out how to deal with channels (if/value/for, class?)
- only update nodes that change (e.g. through bound storage)
- run tests in browser (preferably with as little js tooling as possible)

## Large stories

- Routing
- Nested components
- Some sort of CSS support

## Building

GOARCH=wasm GOOS=js go build -o lib.wasm main.go

## Interesting resources

https://github.com/mattn/golang-wasm-example/blob/master/main.go
https://github.com/danieljoos/go-wasm-examples
https://github.com/albrow/vdom - 
A virtual dom implementation written in go which is compatible with gopherjs
https://brianketelsen.com/web-assembly-and-go-a-look-to-the-future/
https://tutorialedge.net/golang/writing-frontend-web-framework-webassembly-go/#a-full-example

https://medium.com/@alexmaisiura/how-to-create-a-users-search-app-for-github-on-webassembly-4f7346d9f138
https://godoc.org/syscall/js

go run -exec="$(go env GOROOT)/misc/wasm/go_js_wasm_exec" main.go


