package vtree

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

func isComponent(Type string) bool {
	return strings.Contains(Type, "-") && !strings.HasPrefix(Type, "g-")
}

func Parse(s string) *Element {
	dec := xml.NewDecoder(bytes.NewBuffer([]byte(s)))

	dec.Entity = xml.HTMLEntity
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose

	var root *Element
	var current *Element
	// initialize with the toplevel parent nil, so we can properly "pop" it at the end

	parents := []*Element{nil}

	for {
		token, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		switch token.(type) {
		case xml.StartElement: // <li> and <my-component>
			element := token.(xml.StartElement)
			name := element.Name.Local

			// A component is a tag that contains a '-', but skip g-<whatever> tags
			// for now, since we (ab)use them to replace text values
			if isComponent(name) {
				if current != nil {
					comp := Comp(name)
					current.C(comp)
				} else {
					panic("Component " + name + " must be contained inside normal element")
				}
			} else {

				// fmt.Printf("It's a StartElement <%s>\n", name)
				if root == nil {
					root = El(name)
					current = root
				} else {
					parents = append([]*Element{current}, parents...)
					current = El(name)
					parents[0].C(current)
				}
				for _, attr := range element.Attr {
					current.A(attr.Name.Local, attr.Value)
					// not sure what to do with this...
					if attr.Name.Local == "id" {
						current.SetID(ElementID(attr.Value))
					}
				}
			}

		case xml.EndElement: // </li>
			name := token.(xml.EndElement).Name.Local
			if !isComponent(name) {
				current, parents = parents[0], parents[1:]
			}
		case xml.CharData:
			// ignore (whitespace) data before first actual tag
			if current != nil {
				b := token.(xml.CharData).Copy()
				current.T(string(b))
				// fmt.Printf("text: %s\n", string(b))
			}
		case xml.Comment:
			fmt.Println("It's a comment")
			// ProcInst, Directive?
		}
	}

	return root
}
