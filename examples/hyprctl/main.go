// Limited reimplementation of hyprctl using hyprland-go to show an example
// on how it can be used.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/thiagokokada/hyprland-go"
)

// https://stackoverflow.com/a/28323276
type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlags) Set(v string) error {
	*i = append(*i, v)
	return nil
}

func must1[T any](v T, err error) T {
	must(err)
	return v
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustMarshalIndent(v any) []byte {
	return must1(json.MarshalIndent(v, "", "   "))
}

func main() {
	c := hyprland.MustClient()

	dispatchFS := flag.NewFlagSet("dispatch", flag.ExitOnError)
	var dispatch arrayFlags
	dispatchFS.Var(&dispatch, "c", "Command to dispatch. Please quote commands with arguments (e.g.: 'exec kitty')")

	setcursorFS := flag.NewFlagSet("setcursor", flag.ExitOnError)
	theme := setcursorFS.String("theme", "Adwaita", "Cursor theme")
	size := setcursorFS.Int("size", 32, "Cursor size")

	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("Expected subcommand.")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "activewindow":
		v := must1(c.ActiveWindow())
		fmt.Printf("%s\n", mustMarshalIndent(v))
	case "activeworkspace":
		v := must1(c.ActiveWorkspace())
		fmt.Printf("%s\n", mustMarshalIndent(v))
	case "dispatch":
		dispatchFS.Parse(os.Args[2:])
		v := must1(c.Dispatch(dispatch...))
		fmt.Printf("%s\n", v)
	case "kill":
		v := must1(c.Kill())
		fmt.Printf("%s\n", v)
	case "reload":
		v := must1(c.Reload())
		fmt.Printf("%s\n", v)
	case "setcursor":
		setcursorFS.Parse(os.Args[2:])
		v := must1(c.SetCursor(*theme, *size))
		fmt.Printf("%s\n", v)
	case "version":
		v := must1(c.Version())
		fmt.Printf("%s\n", mustMarshalIndent(v))
	default:
		fmt.Printf("[ERROR] Unknown command: %s\n", os.Args[1])
	}
}
