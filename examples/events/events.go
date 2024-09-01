// Basic example on how to handle events in hyprland-go.
package main

import (
	"context"
	"fmt"
	"time"

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
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	c := event.MustClient()
	defer c.Close()

	// Will listen for events for 5 seconds and exit
	c.Subscribe(
		ctx,
		&ev{},
		event.EventWorkspace,
		event.EventActiveWindow,
	)

	fmt.Println("Bye!")
}
