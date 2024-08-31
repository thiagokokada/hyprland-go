// Basic example on how to handle events in hyprland-go.
package main

import (
	"fmt"

	"github.com/thiagokokada/hyprland-go/event"
)

type ev struct {
	event.DefaultEventHandler
}

func (e *ev) Workspace(w event.WorkspaceName) {
	fmt.Printf("Workspace: %+v\n", w)
}

func (e *ev) ActiveWindow(w event.ActiveWindow) {
	fmt.Printf("ActiveWindow: %+v\n", w)
}

func main() {
	c := event.MustClient()
	defer c.Close()

	event.Subscribe(
		c, &ev{},
		event.EventWorkspace,
		event.EventActiveWindow,
	)

}
