// i3-like navigation between windows and groups.
// See https://github.com/hyprwm/Hyprland/discussions/2517 for more details.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/thiagokokada/hyprland-go"
)

func must1[T any](v T, err error) T {
	must(err)
	return v
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: hypr-i3-move <focus|move> <direction>")
		os.Exit(1)
	}
	mode := os.Args[1]
	direction := os.Args[2]
	client := hyprland.MustClient()

	aWindow := must1(client.ActiveWindow())
	grouped := aWindow.Grouped
	addr := aWindow.Address

	switch mode {
	case "focus":
		if len(grouped) == 0 {
			client.Dispatch(fmt.Sprintf("movefocus %s", direction))
			return
		}

		switch direction {
		case "l", "u":
			if addr == grouped[0] {
				client.Dispatch(fmt.Sprintf("movefocus %s", direction))
			} else {
				client.Dispatch("changegroupactive b")
			}
		case "r", "d":
			if addr == grouped[len(grouped)-1] {
				client.Dispatch(fmt.Sprintf("movefocus %s", direction))
			} else {
				client.Dispatch("changegroupactive f")
			}
		default:
			log.Printf("Unknown direction '%s'. Valid options are: l, r, u, d.", direction)
			os.Exit(1)
		}
	case "move":
		if len(grouped) == 0 {
			client.Dispatch(fmt.Sprintf("movewindoworgroup %s", direction))
			return
		}
		switch direction {
		case "l", "u":
			if addr == grouped[0] {
				client.Dispatch(fmt.Sprintf("movewindoworgroup %s", direction))
			} else {
				client.Dispatch("movegroupwindow b")
			}
		case "r", "d":
			if addr == grouped[len(grouped)-1] {
				client.Dispatch(fmt.Sprintf("movewindoworgroup %s", direction))
			} else {
				client.Dispatch("movegroupwindow f")
			}
		default:
			log.Printf("Unknown direction '%s'. Valid options are: l, r, u, d.", direction)
			os.Exit(1)
		}
	default:
		fmt.Println("Invalid mode. Use 'focus' or 'move'.")
		os.Exit(1)
	}
}
