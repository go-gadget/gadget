package main

import (
	"fmt"
	. "gadget"
	"j"
	"vtree"
)

/*
 * perhaps one day this will be a real gadget app
 * Assume this is a component, mounted somewhere in the router/page,
 * doing its thing
 */

/*
 * Let's imagine the router will mount it this way
 *
 * router.add("/foo/:id", SampleComponent)
 *
 * This will call some hooks on the component (which implements this Hooks interface),
 * e.g. comp.created()
 * html = comp.render()
 * app.mount(parent, html)
 * .. and whenever something changes update the html?
 */

type SampleComponent struct {
	Todos     []string
	NewTODO   string
	SomeValue string
	Frop      int
	Show      bool
	Color     string
	i         int
}

func BuildSampleComponent() ComponentInf {
	// create instance, do initialization if necessary

	s := &SampleComponent{
		Todos:     []string{"First entry!", "Second Entry!"},
		SomeValue: "Worst",
		Color:     "red",
		NewTODO:   "x",
	}

	return s
}

func (g *SampleComponent) Init(chan string) {
}

func (g *SampleComponent) Data() interface{} {
	return g
}

func (g *SampleComponent) Template() string {
	// could also be part of SampleComponent creation?
	// return `
	// <h1>TODO app</h1>
	// <input type="text" g-bind="NewTODO">
	// <button g-click="add_todo">Add</button>
	// <ul>
	//   <li g-for="todo in Todos">
	//     {{todo}}
	//   </li>
	// </ul>
	// `

	return `<div id="rootdiv">
	<b {{if .Color }}class="{{.Color}}"{{end}} id="h1-1">{{.SomeValue}} {{.Frop}} </b>
	<br>
	<input id="input-1" type="text" g-bind="NewTODO">
	<button id="button-1" g-click="add_todo">Add</button>
	{{ if .Show }}
	<i id="div-2" class="red">
	 Hello!
	</i>
	{{ end }}
	<ul id="ul-1">
	  {{range $i, $e := .Todos}}
	  <li id="li-{{$i}}">
	    {{.}}
	  </li>
	  {{end}}
	</ul>
	</div>`
}

func (g *SampleComponent) Handlers() map[string]Handler {
	return map[string]Handler{
		"add_todo": func(Updates chan Action) {
			g.Doit()
		},
	}
}

func (g *SampleComponent) Doit() {
	i := g.i
	g.i++

	// can't call SetValue() (it's defined on Component),
	// which means we can't track *which* value changed.
	g.Todos = append(g.Todos, g.NewTODO)
	j.J("add_todo called", g.Todos)
	// g.NewTODO = fmt.Sprintf("And another entry %d", i)
	g.SomeValue = "Kaas"
	g.Frop = i + 1000
	g.Show = i%2 == 1

	switch i % 3 {
	case 0:
		g.Color = "red"
	case 1:
		g.Color = "green"
	case 2:
		g.Color = ""
	}
}

func main() {
	fmt.Println("Go Go Gadget!")

	g := NewGadget(vtree.NewDomBridge())

	go g.MainLoop()
	component := NewComponent()
	component.Build(BuildSampleComponent)

	g.Mount(component)
	select {}
}
