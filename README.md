# Gadget - frontend framework for Go

"Gadget" is my feeble attempt at building a VueJS inspired frontend framework for Golang (wasm).

## Status

Currently it can run very simple single-component mini apps. It has its own template mini language (using g-something attributes), simple data bindings
and action handling through "handlers"

Expect the API and code to be pre-alpha. Things will change and break.

Check https://github.com/go-gadget/examples for actual sample code.

Or try the 'todo' example directly on https://go-gadget.github.io/index.html

## Running code

```
$ GOARCH=wasm GOOS=js go build -o lib.wasm github.com/go-gadget/examples

$ bin/gadget serve
```

For now make sure you copy src/github.com/go-gadget/gadget/cmd/gadget/{index.html,wasm_exec.js} to the directory you're starting bin/gadget from.
