package gadget

import "fmt"

func DumpMounts(c *ComponentInstance, level int) {
	for _, m := range c.State.Mounts {
		fmt.Printf("[%d] %s %v -> %d\n", level, m.Name, m.Component, len(m.Component.State.Mounts))
		DumpMounts(m.Component, level+1)
	}
}
