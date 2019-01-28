package vtree

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

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

			if root == nil {
				root = El(name)
				current = root
			} else {
				parents = append([]*Element{current}, parents...)
				current = El(name)
				parents[0].C(current)
			}
			for _, attr := range element.Attr {
				attrName := attr.Name.Local
				if strings.HasPrefix(attr.Name.Space, "g-") {
					attrName = attr.Name.Space + ":" + attr.Name.Local
				}
				current.A(attrName, attr.Value)
				// not sure what to do with this...
				if attr.Name.Local == "id" {
					current.SetID(ElementID(attr.Value))
				}
			}

		case xml.EndElement: // </li>
			current, parents = parents[0], parents[1:]
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
